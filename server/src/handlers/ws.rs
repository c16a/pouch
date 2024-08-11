use crate::processor::db::InMemoryDb;
use crate::processor::spec::Processor;
use crate::wal::WAL;
use futures_util::{SinkExt, StreamExt};
use pouch_sdk::command::Command;
use pouch_sdk::response::Error::UnknownCommand;
use pouch_sdk::response::Response;
use std::env;
use std::ops::DerefMut;
use std::sync::Arc;
use tokio::net::{TcpListener, TcpStream};
use tokio::sync::RwLock;
use tokio_tungstenite::accept_async;
use tokio_tungstenite::tungstenite::Message;

pub(crate) async fn handle_websocket(db: Arc<RwLock<InMemoryDb>>, wal: Arc<RwLock<WAL>>) {
    let enable_ws: bool = env::var("ENABLE_WS")
        .ok()
        .and_then(|p| p.parse().ok())
        .unwrap_or(false);
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
            println!("Started WS listener on {}", ws_address);
            loop {
                let (socket, _) = ws_listener.accept().await.unwrap();
                let wal = wal_ws.clone();
                let db = db_ws.clone();
                tokio::spawn(async move {
                    process_websocket(socket, db, wal).await;
                });
            }
        }).await.unwrap();
    }
}

async fn process_websocket(
    stream: TcpStream,
    db: Arc<RwLock<dyn Processor>>,
    wal: Arc<RwLock<WAL>>,
) {
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
                                Ok(cmd) => db
                                    .write()
                                    .await
                                    .cmd(cmd, Some(wal.write().await.deref_mut())),
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
