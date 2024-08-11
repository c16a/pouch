use futures_util::{SinkExt, StreamExt};
use std::env;
use std::ops::DerefMut;
use std::sync::Arc;
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use tokio::net::{TcpListener, TcpStream};
use tokio::sync::RwLock;
use tokio_tungstenite::accept_async;
use tokio_tungstenite::tungstenite::protocol::Message;

use crate::handlers::swarm::init_swarm;
use crate::handlers::tcp::handle_tcp;
use crate::handlers::ws::handle_websocket;
use crate::processor::db::InMemoryDb;
use crate::processor::spec::Processor;
use pouch_sdk::command::Command;
use pouch_sdk::response::Error::UnknownCommand;
use pouch_sdk::response::Response;
use wal::WAL;

mod handlers;
mod p2p;
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

    {
        init_swarm().await;
    }

    {
        let wal = wal.clone();
        let db = db.clone();
        handle_tcp(db, wal).await;
    }

    {
        let wal = wal.clone();
        let db = db.clone();
        handle_websocket(db, wal).await;
    }
}
