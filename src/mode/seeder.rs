use serde::Deserialize;
use std::{
    collections::{HashMap, HashSet},
    net::SocketAddr,
    path::PathBuf,
};

pub type Rating = u8;

#[derive(Deserialize, Debug, Clone)]
pub struct Seeder {
    pub address: SocketAddr,
    pub file: PathBuf,
    #[serde(skip)]
    pub nodes: HashMap<SocketAddr, Rating>,
    #[serde(skip)]
    pub distributed_storage: HashMap<usize, HashSet<SocketAddr>>,
}
