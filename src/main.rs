mod consts;
mod mode;
mod peer;

use bincode::{Decode, Encode, config::standard, decode_from_slice, encode_to_vec};
use consts::*;
use peer::{Chunk, Config, Mode, Peer};
use rand::{SeedableRng, rngs::SmallRng, seq::IteratorRandom};
use std::{
    collections::{HashSet, VecDeque},
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

#[derive(Encode, Decode, PartialEq, Debug)]
struct Retrieve {
    chunk_index: usize,
    address: SocketAddr,
}

#[derive(Encode, Decode, PartialEq, Debug)]
enum Packet {
    Join(Join),
    Retrieve(Retrieve),
}

fn main() {
    let toml_config: Config = from_str(&read_to_string("config.toml").unwrap()).unwrap();

    let mut peer = Peer::new(toml_config);

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
                let (packet, _): (Packet, _) = decode_from_slice(slice, config).unwrap();

                dbg!(&packet);

                match packet {
                    Packet::Join(join) => {
                        let mut rng = SmallRng::from_os_rng();
                        let random_chunks = peer
                            .chunks
                            .iter()
                            .choose_multiple(&mut rng, join.contribution);

                        if let Ok(ref mut stream) = TcpStream::connect(join.address) {
                            let chunk_encoded = encode_to_vec(&random_chunks, config).unwrap();

                            stream.write_all(&chunk_encoded).unwrap();
                            seeder.nodes.insert(join.address, DEFAULT_RATING);
                            for (index, chunk) in random_chunks {
                                seeder
                                    .distributed_storage
                                    .entry(*index)
                                    .and_modify(|contributors| {
                                        contributors.insert(join.address);
                                    })
                                    .or_insert(HashSet::from([join.address]));
                            }
                        } else {
                            println!("Couldn't connect to server...");
                        }
                    }
                    Packet::Retrieve(retrieve) => {
                        let retrieved_chunk = peer.chunks.get(&retrieve.chunk_index);
                        if let Ok(ref mut stream) = TcpStream::connect(retrieve.address) {
                            let encoded = encode_to_vec(
                                peer.chunks.get(&retrieve.chunk_index).unwrap(),
                                config,
                            )
                            .unwrap();
                            stream.write_all(&encoded).unwrap();
                        } else {
                            println!("Couldn't connect to server...");
                        }
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
                    Packet::Join(Join {
                        contribution: node.contribution,
                        address: node.address,
                    }),
                    config,
                )
                .unwrap();

                stream.write_all(&encoded).unwrap();

                let (mut stream_incoming, _) = listener.accept().unwrap();
                let mut buf = vec![];

                let len = stream_incoming.read_to_end(&mut buf).unwrap();
                let slice = &buf[..len];
                let (data, _): (Vec<(usize, Chunk)>, _) = decode_from_slice(slice, config).unwrap();
                for (index, chunk) in data {
                    peer.chunks.insert(index, chunk);
                }
            }

            // if let Ok(mut stream) = TcpStream::connect(node.seeder_address) {
            //     let config = standard();
            //
            //     let mut needed_chunk_index = 0;
            //     let mut buffer = VecDeque::with_capacity(BUFFER_SIZE);
            //     // back => newest
            //     while buffer.len() < BUFFER_SIZE {
            //         dbg!(buffer.len());
            //         match peer.chunks.get(&needed_chunk_index) {
            //             Some(chunk) => {
            //                 buffer.push_back(chunk.clone());
            //                 needed_chunk_index += 1;
            //             }
            //             None => {
            //                 // let packet_encoded = encode_to_vec(
            //                 //     Packet::Retrieve(Retrieve {
            //                 //         chunk_index: needed_chunk_index,
            //                 //         address: node.address,
            //                 //     }),
            //                 //     config,
            //                 // )
            //                 // .unwrap();
            //                 //
            //                 // stream.write_all(&packet_encoded).unwrap();
            //                 //
            //                 // for stream_incoming in listener.incoming() {
            //                 //     let mut buf = Vec::new();
            //                 //
            //                 //     let len = stream_incoming.unwrap().read_to_end(&mut buf).unwrap();
            //                 //     let slice = &buf[..len];
            //                 //     let (chunk, _): (Chunk, _) =
            //                 //         decode_from_slice(slice, config).unwrap();
            //                 //
            //                 //     buffer.push_back(chunk.clone());
            //                 //     needed_chunk_index += 1;
            //                 // }
            //             }
            //         }
            //     }
            //
            //     dbg!(buffer.len());
            //
            //     println!("Connected to the server!");
            // } else {
            //     println!("Couldn't connect to the server...");
            // }
        }
    }
}
