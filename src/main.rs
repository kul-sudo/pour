mod consts;
mod mode;
mod peer;

use bincode::{Decode, Encode, config::standard, decode_from_slice, encode_to_vec};
use consts::*;
use peer::{Chunk, Config, Mode, Peer};
use rand::{SeedableRng, rngs::SmallRng, seq::IteratorRandom};
use std::{
    collections::HashSet,
    fs::{read_to_string, write},
    io::{Read, Write},
    net::{SocketAddr, TcpListener, TcpStream},
    sync::{Arc, RwLock},
    thread::spawn,
};
use toml::from_str;

#[derive(Encode, Decode, PartialEq, Debug)]
struct Join {
    contribution: usize,
    address: SocketAddr,
}

fn main() {
    let config: Config = from_str(&read_to_string("config.toml").unwrap()).unwrap();

    let mut peer = Peer::new(config);

    match peer.mode {
        Mode::Seeder(ref mut seeder) => {
            let listener = Arc::new(RwLock::new(TcpListener::bind(seeder.address).unwrap()));
            let lc = listener.clone();
            // spawn(move || {
            let listener_read = lc.read().unwrap();
            for stream in listener_read.incoming() {
                let config = standard();

                let mut buf = [0; 100];
                let len = stream.unwrap().read(&mut buf).unwrap();
                let slice = &buf[..len];
                let (node, _): (Join, _) = decode_from_slice(slice, config).unwrap();

                let mut rng = SmallRng::from_os_rng();
                let random_chunks = peer
                    .chunks
                    .iter()
                    .choose_multiple(&mut rng, node.contribution);

                for (index, chunk) in random_chunks {
                    let chunk_encoded = encode_to_vec((index, chunk), config).unwrap();

                    if let Ok(ref mut stream) = TcpStream::connect(node.address) {
                        stream.write_all(&chunk_encoded).unwrap();
                        seeder.nodes.insert(node.address, DEFAULT_RATING);
                        seeder
                            .distributed_storage
                            .entry(*index)
                            .and_modify(|contributors| {
                                contributors.insert(node.address);
                            })
                            .or_insert(HashSet::from([node.address]));
                    } else {
                        println!("Couldn't connect to server...");
                    }
                }
            }
            // });
        }
        Mode::Node(node) => {
            let listener = TcpListener::bind(node.address).unwrap();

            if let Ok(mut stream) = TcpStream::connect(node.seeder_address) {
                let config = standard();

                let encoded = encode_to_vec(
                    &Join {
                        contribution: node.contribution,
                        address: node.address,
                    },
                    config,
                )
                .unwrap();

                stream.write_all(&encoded).unwrap();

                for stream_incoming in listener.incoming() {
                    let mut buf = Vec::new();

                    let len = stream_incoming.unwrap().read_to_end(&mut buf).unwrap();
                    let slice = &buf[..len];
                    let ((index, chunk), _): ((usize, Chunk), _) =
                        decode_from_slice(slice, config).unwrap();
                    peer.chunks.insert(index, chunk);
                }

                println!("Connected to the server!");
            } else {
                println!("Couldn't connect to server...");
            }
        }
    }
}
