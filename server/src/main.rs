use std::env;
use std::ops::DerefMut;
use std::sync::Arc;
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use tokio::net::{TcpListener, TcpStream};
use tokio::sync::RwLock;

use crate::processor::db::InMemoryDb;
use crate::processor::spec::Processor;
use pouch_sdk::command::Command;
use pouch_sdk::response::Error::UnknownCommand;
use pouch_sdk::response::Response;
use wal::WAL;

mod processor;
mod structures;
mod wal;

#[tokio::main]
async fn main() {
    // Default values
    let default_port = 6379;
    let default_wal_file = "wal.log".to_string();

    let port: u16 = env::var("PORT")
        .ok()
        .and_then(|p| p.parse().ok())
        .unwrap_or(default_port);

    let wal_file = env::var("WAL_FILE").unwrap_or(default_wal_file);

    let address = format!("127.0.0.1:{}", port);
    let tcp_listener = TcpListener::bind(&address).await.unwrap();

    let wal = Arc::new(RwLock::new(WAL::new(&wal_file).unwrap()));
    let db = Arc::new(RwLock::new(InMemoryDb::new().unwrap()));

    {
        let mut db = db.write().await;

        match wal.read().await.replay(&mut db) {
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

async fn process(mut socket: TcpStream, db: Arc<RwLock<dyn Processor>>, wal: Arc<RwLock<WAL>>) {
    let mut buf = vec![0; 1024];

    loop {
        let n = match socket.read(&mut buf).await {
            Ok(n) if n == 0 => return, // Connection closed
            Ok(n) => n,
            Err(err) => {
                eprintln!("failed to read from socket; err = {:?}", err);
                return;
            }
        };

        // Slice the buffer to only include the bytes that were read
        let json_str = match std::str::from_utf8(&buf[..n]) {
            Ok(json_str) => json_str,
            Err(err) => {
                eprintln!("failed to convert buffer to string; err = {:?}", err);
                continue;
            }
        };

        // Parse the JSON command
        let response = match Command::from_json(json_str) {
            Err(err) => {
                eprintln!("error parsing command: {}", err);
                Response::Err { error: UnknownCommand }
            }
            Ok(cmd) => {
                // Process the command
                db.write()
                    .await
                    .cmd(cmd, Some(wal.write().await.deref_mut()))
            }
        };

        let json_str = response.to_json().unwrap();

        // Write the response to the socket
        if let Err(err) = socket.write_all(format!("{}\n", json_str).as_bytes()).await {
            eprintln!("failed to write data to socket; err = {:?}", err);
            return;
        }

        // Clear the buffer for the next read
        buf.clear();
        buf.resize(1024, 0); // Reset the buffer size
    }
}
