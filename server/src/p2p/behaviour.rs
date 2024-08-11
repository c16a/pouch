use libp2p::kad::store::MemoryStore;
use libp2p::swarm::NetworkBehaviour;
use libp2p::{kad, mdns};

#[derive(NetworkBehaviour)]
pub(crate) struct Behaviour {
    pub(crate) kademlia: kad::Behaviour<MemoryStore>,
    pub(crate) mdns: mdns::async_io::Behaviour,
}
