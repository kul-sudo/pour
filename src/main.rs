mod consts;
mod mode;
mod peer;

use bincode::{Decode, Encode, config::standard, decode_from_slice, encode_to_vec};
use peer::{Chunk, Config, Mode, Peer};
use rand::{SeedableRng, rngs::SmallRng, seq::IteratorRandom};
use std::{
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
        Mode::Seeder(ref seeder) => {
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
                for chunk in random_chunks {
                    let chunk_encoded = encode_to_vec(chunk, config).unwrap();
                    peer.send_to(node.address, chunk_encoded);
                }
                // let mut encoder = ZlibEncoder::new(Vec::new(), CompressionOptions {
                //     max_hash_checks: 32768,
                //     lazy_if_less_than: 258,
                //     matching_type: MatchingType::Lazy,
                //     special: SpecialOptions::Normal
                // });
                // encoder.write_all(&chunk_encoded).expect("Write error!");
                // let compressed_data = encoder.finish().expect("Failed to finish compression!");
                // dbg!(compressed_data.len(), chunk_encoded.len());
                // peer.send_to(node.address, chunk_encoded);
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

                for stream1 in listener.incoming() {
                    let mut buf = Vec::new();

                    let len = stream1.unwrap().read_to_end(&mut buf).unwrap();
                    let slice = &buf[..len];
                    let ((index, chunk), a): ((usize, Chunk), _) =
                        decode_from_slice(slice, config).unwrap();
                    write(format!("share/test{}.webm", index), chunk.bytes.clone()).unwrap(); 
                    peer.chunks.insert(index, chunk);
                }

                println!("Connected to the server!");
            } else {
                println!("Couldn't connect to server...");
            }
        }
    }
}
