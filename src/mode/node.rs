use serde::Deserialize;
use std::net::SocketAddr;

#[derive(Deserialize, Debug, Clone)]
pub struct Node {
    pub address: SocketAddr,
    pub seeder_address: SocketAddr,
    pub contribution: usize,
}
