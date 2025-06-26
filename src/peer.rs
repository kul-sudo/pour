use crate::consts::*;
use serde::Deserialize;
use std::{
    collections::HashMap,
    fs::{create_dir, read, read_dir, remove_dir_all},
    io::Write,
    net::{SocketAddr, TcpStream},
    path::Path,
    process::Command,
};

use crate::mode::{node::Node, seeder::Seeder};
use bincode::{Decode, Encode, config::standard, decode_from_slice, encode_to_vec};
use sha2::{Digest, Sha512};

#[derive(Deserialize, Debug)]
pub enum Mode {
    Seeder(Seeder),
    Node(Node),
}

#[derive(Encode, Decode, PartialEq, Deserialize, Default, Clone, Debug)]
pub struct Chunk {
    pub bytes: Vec<u8>,
    pub hash: Vec<u8>,
}

#[derive(Deserialize, Debug)]
#[serde(tag = "mode", content = "options")]
pub enum Config {
    Seeder(Seeder),
    Node(Node),
}

#[derive(Deserialize, Debug)]
pub struct Peer {
    pub mode: Mode,
    pub chunks: HashMap<usize, Chunk>,
}

impl Peer {
    pub fn new(config: Config) -> Self {
        Self {
            mode: match config {
                Config::Seeder(ref data) => Mode::Seeder(data.clone()),
                Config::Node(ref data) => Mode::Node(data.clone()),
            },
            chunks: {
                match config {
                    Config::Seeder(seeder) => {
                        let file = Path::new(SHARE_DIR).join(seeder.file);
                        let dir = Path::new(SHARE_DIR).join(TMP_DIR);
                        create_dir(&dir).unwrap();
                        let file_name = file.file_name().unwrap();

                        Command::new("ffmpeg")
                            .current_dir(SHARE_DIR)
                            .arg("-i")
                            .arg(file_name)
                            .arg("-c")
                            .arg("copy")
                            .arg("-map")
                            .arg("0")
                            .arg("-segment_time")
                            .arg(CHUNK_DURATION)
                            .arg("-f")
                            .arg("segment")
                            .arg("-reset_timestamps")
                            .arg("1")
                            .arg(Path::new(&TMP_DIR).join("%d.webm"))
                            .output()
                            .unwrap();

                        let chunk_dir_read = read_dir(&dir).unwrap().collect::<Vec<_>>();
                        let mut chunks = HashMap::with_capacity(chunk_dir_read.len());

                        for chunk in &chunk_dir_read {
                            let chunk_path = chunk.as_ref().unwrap().path();
                            let stem = chunk_path.file_stem().unwrap();
                            let bytes = read(&chunk_path).unwrap();

                            let hash = Sha512::digest(&bytes)[..].to_vec();

                            // chunks.insert(
                            //     stem.to_string_lossy().parse::<usize>().unwrap(),
                            //     bytes
                            //         .chunks(SUBCHUNK_N)
                            //         .map(|pieces| Chunk {
                            //             bytes: pieces.iter().cloned().collect::<Vec<_>>(),
                            //             hash: Sha512::digest(&bytes)[..].to_vec(),
                            //         })
                            //         .collect::<Vec<_>>(), // Chunk { bytes, hash },
                            // );

                            chunks.insert(
                                stem.to_string_lossy().parse::<usize>().unwrap(),
                                Chunk { bytes, hash },
                            );
                        }

                        remove_dir_all(&dir).unwrap();
                        chunks
                    }
                    Config::Node(..) => HashMap::new(),
                }
            },
        }
    }
}
