use futures_util::{SinkExt, StreamExt};
use std::env;
use std::ops::DerefMut;
use std::sync::Arc;
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use tokio::net::{TcpListener, TcpStream};
use tokio::sync::RwLock;
use tokio_tungstenite::accept_async;
use tokio_tungstenite::tungstenite::protocol::Message;

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
    let default_wal_file = "wal.log".to_string();
    let wal_file = env::var("WAL_FILE").unwrap_or(default_wal_file);
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

    let enable_tcp: bool = env::var("ENABLE_TCP").ok().and_then(|p| p.parse().ok()).unwrap_or(true);
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
            format!("started TCP listener on {}", tcp_address);
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

    let enable_ws: bool = env::var("ENABLE_WS").ok().and_then(|p| p.parse().ok()).unwrap_or(false);
    if enable_ws {
        let default_ws_port = 6389;
        let ws_port: u16 = env::var("WS_PORT")
            .ok()
            .and_then(|p| p.parse().ok())
            .unwrap_or(default_ws_port);

        let default_ws_host = String::from("0.0.0.0");
        let ws_host = env::var("WS_HOST").ok().unwrap_or(default_ws_host);
        let ws_address = format!("{}:{}", ws_host, ws_port);
        let ws_listener = TcpListener::bind(&ws_address).await.unwrap();

        let wal_ws = wal.clone();
        let db_ws = db.clone();

        tokio::spawn(async move {
            format!("started WS listener on {}", ws_address);
            loop {
                let (socket, _) = ws_listener.accept().await.unwrap();
                let wal = wal_ws.clone();
                let db = db_ws.clone();
                tokio::spawn(async move {
                    handle_websocket(socket, db, wal).await;
                });
            }
        }).await.unwrap();
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

async fn handle_websocket(stream: TcpStream, db: Arc<RwLock<dyn Processor>>, wal: Arc<RwLock<WAL>>) {
    match accept_async(stream).await {
        Ok(ws_stream) => {
            let (mut ws_sender, mut ws_receiver) = ws_stream.split();

            while let Some(msg) = ws_receiver.next().await {
                match msg {
                    Ok(msg) => {
                        if msg.is_text() {
                            let json_str = msg.into_text().unwrap();

                            let response = match Command::from_json(&json_str) {
                                Err(err) => {
                                    eprintln!("error parsing command: {}", err);
                                    Response::Err {
                                        error: UnknownCommand,
                                    }
                                }
                                Ok(cmd) => {
                                    db.write()
                                        .await
                                        .cmd(cmd, Some(wal.write().await.deref_mut()))
                                }
                            };

                            let json_str = response.to_json().unwrap();

                            if let Err(err) = ws_sender.send(Message::text(json_str)).await {
                                eprintln!("failed to send WebSocket message; err = {:?}", err);
                            }
                        } else if msg.is_close() {
                            break;
                        }
                    }
                    Err(err) => {
                        eprintln!("WebSocket error: {:?}", err);
                        break;
                    }
                }
            }
        }
        Err(err) => {
            eprintln!("Failed to accept WebSocket connection; err = {:?}", err);
        }
    }
}