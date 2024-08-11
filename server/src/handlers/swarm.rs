use crate::p2p::behaviour::{Behaviour, BehaviourEvent};
use async_std::io;
use futures::{prelude::*, select};
use libp2p::kad::store::MemoryStore;
use libp2p::kad::Mode;
use libp2p::swarm::SwarmEvent;
use libp2p::{kad, mdns, noise, tcp, yamux};
use std::error::Error;
use std::time::Duration;

pub(crate) async fn init_swarm() -> Result<(), Box<dyn Error>> {
    let mut swarm = libp2p::SwarmBuilder::with_new_identity()
        .with_async_std()
        .with_tcp(
            tcp::Config::default(),
            noise::Config::new,
            yamux::Config::default,
        )?
        .with_behaviour(|key| {
            Ok(Behaviour {
                kademlia: kad::Behaviour::new(
                    key.public().to_peer_id(),
                    MemoryStore::new(key.public().to_peer_id()),
                ),
                mdns: mdns::async_io::Behaviour::new(
                    mdns::Config::default(),
                    key.public().to_peer_id(),
                )?,
            })
        })?
        .with_swarm_config(|c| c.with_idle_connection_timeout(Duration::from_secs(60)))
        .build();

    swarm.behaviour_mut().kademlia.set_mode(Some(Mode::Server));

    // Listen on all interfaces and whatever port the OS assigns.
    swarm.listen_on("/ip4/0.0.0.0/tcp/0".parse()?)?;

    tokio::spawn(async move {
        loop {
            select! {
                event = swarm.select_next_some() => match event {
                    SwarmEvent::NewListenAddr { address, .. } => {
                        println!("Swarm listening in {address:?}");
                    },
                    SwarmEvent::Behaviour(BehaviourEvent::Mdns(mdns::Event::Discovered(list))) => {
                        for (peer_id, multiaddr) in list {
                            swarm.behaviour_mut().kademlia.add_address(&peer_id, multiaddr);
                        }
                    }
                    SwarmEvent::Behaviour(BehaviourEvent::Kademlia(kad::Event::OutboundQueryProgressed { result, ..})) => {
                        match result {
                            kad::QueryResult::GetProviders(Ok(kad::GetProvidersOk::FoundProviders { key, providers, .. })) => {
                                for peer in providers {
                                    println!(
                                        "Peer {peer:?} provides key {:?}",
                                        std::str::from_utf8(key.as_ref()).unwrap()
                                    );
                                }
                            }
                            kad::QueryResult::GetProviders(Err(err)) => {
                                eprintln!("Failed to get providers: {err:?}");
                            }
                            kad::QueryResult::GetRecord(Ok(
                                kad::GetRecordOk::FoundRecord(kad::PeerRecord {
                                    record: kad::Record { key, value, .. },
                                    ..
                                })
                            )) => {
                                println!(
                                    "Got record {:?} {:?}",
                                    std::str::from_utf8(key.as_ref()).unwrap(),
                                    std::str::from_utf8(&value).unwrap(),
                                );
                            }
                            kad::QueryResult::GetRecord(Ok(_)) => {}
                            kad::QueryResult::GetRecord(Err(err)) => {
                                eprintln!("Failed to get record: {err:?}");
                            }
                            kad::QueryResult::PutRecord(Ok(kad::PutRecordOk { key })) => {
                                println!(
                                    "Successfully put record {:?}",
                                    std::str::from_utf8(key.as_ref()).unwrap()
                                );
                            }
                            kad::QueryResult::PutRecord(Err(err)) => {
                                eprintln!("Failed to put record: {err:?}");
                            }
                            kad::QueryResult::StartProviding(Ok(kad::AddProviderOk { key })) => {
                                println!(
                                    "Successfully put provider record {:?}",
                                    std::str::from_utf8(key.as_ref()).unwrap()
                                );
                            }
                            kad::QueryResult::StartProviding(Err(err)) => {
                                eprintln!("Failed to put provider record: {err:?}");
                            }
                            _ => {}
                        }
                    }
                    _ => {}
                }
            }
        }
    });

    Ok(())
}
