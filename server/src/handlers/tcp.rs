use crate::processor::db::InMemoryDb;
use crate::processor::spec::Processor;
use crate::wal::WAL;
use pouch_sdk::command::Command;
use pouch_sdk::response::Error::UnknownCommand;
use pouch_sdk::response::Response;
use std::env;
use std::ops::DerefMut;
use std::sync::Arc;
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use tokio::net::{TcpListener, TcpStream};
use tokio::sync::RwLock;

pub(crate) async fn handle_tcp(db: Arc<RwLock<InMemoryDb>>, wal: Arc<RwLock<WAL>>) {
    let enable_tcp: bool = env::var("ENABLE_TCP")
        .ok()
        .and_then(|p| p.parse().ok())
        .unwrap_or(true);
    if enable_tcp {
        let default_tcp_port = 6379;
        let tcp_port: u16 = env::var("TCP_PORT")
            .ok()
            .and_then(|p| p.parse().ok())
            .unwrap_or(default_tcp_port);

        let default_tcp_host = String::from("0.0.0.0");
        let tcp_host = env::var("TCP_HOST").ok().unwrap_or(default_tcp_host);
        let tcp_address = format!("{}:{}", tcp_host, tcp_port);

        let tcp_listener = TcpListener::bind(&tcp_address).await.unwrap();

        let wal_tcp = wal.clone();
        let db_tcp = db.clone();

        tokio::spawn(async move {
            println!("Started TCP listener on {}", tcp_address);
            loop {
                let (socket, _) = tcp_listener.accept().await.unwrap();
                let wal = wal_tcp.clone();
                let db = db_tcp.clone();
                tokio::spawn(async move {
                    process_tcp(socket, db, wal).await;
                });
            }
        });
    }
}

async fn process_tcp(mut socket: TcpStream, db: Arc<RwLock<dyn Processor>>, wal: Arc<RwLock<WAL>>) {
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
                Response::Err {
                    error: UnknownCommand,
                }
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
