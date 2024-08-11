use crate::p2p::behaviour::{Behaviour, BehaviourEvent};
use futures::{prelude::*, select};
use libp2p::kad::store::MemoryStore;
use libp2p::kad::Mode;
use libp2p::swarm::SwarmEvent;
use libp2p::{kad, mdns, noise, tcp, yamux, Multiaddr};
use std::env;
use std::error::Error;
use std::time::Duration;

pub(crate) async fn init_swarm() -> Result<(), Box<dyn Error>> {
    let default_swarm_port = 7001;
    let swarm_port: u16 = env::var("SWARM_PORT")
        .ok()
        .and_then(|p| p.parse().ok())
        .unwrap_or(default_swarm_port);

    let default_swarm_host = String::from("0.0.0.0");
    let swarm_host = env::var("SWARM_HOST").ok().unwrap_or(default_swarm_host);

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
                    mdns::Config {
                        ttl: Duration::from_secs(20),
                        query_interval: Duration::from_secs(5),
                        enable_ipv6: false,
                    },
                    key.public().to_peer_id(),
                )?,
            })
        })?
        .with_swarm_config(|c| c.with_idle_connection_timeout(Duration::from_secs(60)))
        .build();

    swarm.behaviour_mut().kademlia.set_mode(Some(Mode::Server));

    // Listen on all interfaces and whatever port the OS assigns.

    let swarm_addr: Multiaddr = format!("/ip4/{}/tcp/{}", swarm_host, swarm_port).parse()?;
    match swarm.listen_on(swarm_addr) {
        Ok(listener_id) => {
            println!("Swarm active, listener_id={}", listener_id);
        }
        Err(err) => {
            eprintln!("failed to initialise swarm {:?}", err);
        }
    }

    if let Some(swarm_peer_url) = env::var("SWARM_PEER").ok() {
        let peer_addr: Multiaddr = swarm_peer_url.parse()?;
        match swarm.dial(peer_addr.clone()) {
            Ok(_) => {
                println!("dialed peer={}", peer_addr);
            }
            Err(err) => {
                eprintln!("failed to dial peer={}, err={:?}", peer_addr, err);
            }
        }
    }

    tokio::spawn(async move {
        loop {
            select! {
                event = swarm.select_next_some() => match event {
                    SwarmEvent::NewListenAddr { address, .. } => {
                        println!("Swarm listening in {address:?}");
                    },
                    SwarmEvent::Behaviour(BehaviourEvent::Mdns(mdns::Event::Discovered(list))) => {
                        for (peer_id, multiaddr) in list {
                            println!("Connected to peer: peer_id={}, addr={}", peer_id, multiaddr);
                            swarm.behaviour_mut().kademlia.add_address(&peer_id, multiaddr);
                        }
                    }
                    SwarmEvent::Behaviour(BehaviourEvent::Mdns(mdns::Event::Expired(list))) => {
                        for (peer_id, multiaddr) in list {
                            println!("Disconnected from peer: peer_id={}, addr={}", peer_id, multiaddr);
                            swarm.behaviour_mut().kademlia.remove_address(&peer_id, &multiaddr);
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
