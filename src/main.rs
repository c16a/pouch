use std::sync::Arc;

use tokio::io::{AsyncReadExt, AsyncWriteExt};
use tokio::net::{TcpListener, TcpStream};
use tokio::sync::RwLock;

use command::Command;
use wal::WAL;

use crate::db::InMemoryDb;

mod command;
mod wal;
mod response;
mod db;

#[tokio::main]
async fn main() {
    let tcp_listener = TcpListener::bind("127.0.0.1:6379").await.unwrap();

    let wal = Arc::new(RwLock::new(WAL::new("wal.log").unwrap()));
    let db = Arc::new(RwLock::new(InMemoryDb::new().unwrap()));

    {
        let mut db2 = db.write().await;

        match wal.read().await.replay(&mut db2) {
            Ok(count) => {
                println!("restored {} entries from WAL", count)
            }
            Err(err) => {
                eprintln!("failed to read WAL entries; err = {:?}", err);
            }
        }
    }

    loop {
        let (socket, _) = tcp_listener.accept().await.unwrap();
        let wal = wal.clone();
        let db = db.clone();
        tokio::spawn(async move {
            process(socket, db, wal).await;
        });
    }
}

async fn process(mut socket: TcpStream, db: Arc<RwLock<InMemoryDb>>, wal: Arc<RwLock<WAL>>) {
    let mut buf = vec![0; 1024];

    loop {
        let n = match socket
            .read(&mut buf)
            .await {
            Ok(n) if n == 0 => return,
            Ok(n) => n,
            Err(err) => {
                eprintln!("failed to read from socket; err = {:?}", err);
                return;
            }
        };

        // let data = Bytes::copy_from_slice(&buf[..n]);
        let response = match Command::from_slice(&buf[..n]) {
            None => continue,
            Some(cmd) => match cmd {
                Command::Get { ref key } => {
                    db.read().await.get(key)
                }
                Command::Set { ref key, ref value } => {
                    wal.write().await.log(&cmd).unwrap();
                    db.write().await.insert(key, value)
                }
                Command::Delete { ref key } => {
                    wal.write().await.log(&cmd).unwrap();
                    db.write().await.remove(key)
                }
                Command::LPush { ref key, ref value } => {
                    wal.write().await.log(&cmd).unwrap();
                    db.write().await.lpush(key, value)
                }
                Command::RPush { ref key, ref value } => {
                    wal.write().await.log(&cmd).unwrap();
                    db.write().await.rpush(key, value)
                }
                Command::LRange { ref key, start, end } => {
                    db.read().await.lrange(key, start, end)
                }
            }
        };

        socket.write_all(response.to_vec().as_slice()).await
            .expect("failed to write data to socket");
    }
}