package seeder

import (
	"net"
	"pour/packet"
	"pour/workers"
	"pour/bootstrap"
	"pour/dashboard"
	"encoding/gob"
	"sync"
	"fmt"
	"strings"
	"strconv"
	"path/filepath"
	"os"
	"math/rand"
)

type Seeder struct {
	Nodes []string
}

func Setup(config *bootstrap.Config) {
	go workers.Segmentation()

	seeder := Seeder{make([]string, 0)}
	go seeder.HandleNewChunks()

	var wg sync.WaitGroup

	page := dashboard.Page { Nodes: &seeder.Nodes, Dashboard: config.Dashboard  }
	go dashboard.ShowSeederInfo(&page)

	ln, err := net.Listen("tcp", config.Seeder.Address)
	if err != nil {
		fmt.Printf("failed to listen on port\n")
		return
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("failed to accept connection\n")
		}

		wg.Add(1)
		go seeder.HandleConnection(conn, &wg)
	}
}

func (seeder *Seeder) HandleNewChunks() {
	lastNum := -1

	for {
		if len(seeder.Nodes) == 0 {
			continue	
		}

		files, err := os.ReadDir(workers.SEGMENTS_DIR)
		if err != nil {
			fmt.Printf("failed to read the segments dir")
			return
		}

		maxNum := -1
		for _, f := range files {
			filename := strings.TrimSuffix(f.Name(), ".mp4")
			if num, err := strconv.Atoi(filename); err == nil {
				if num > maxNum {
					maxNum = num
				}
			}
		}

		maxNum -= 1
		maxNum = 5

		if maxNum >= 0 {
			if lastNum != maxNum {
				lastNum = maxNum
				fullName := filepath.Join(workers.SEGMENTS_DIR, fmt.Sprintf("%d.mp4", maxNum));
				file, err := os.Open(fullName)
				defer file.Close()
				if err != nil {
					fmt.Printf("failed to open the file")
					return
				}

				randomNode := seeder.Nodes[rand.Intn(len(seeder.Nodes))]
				bytes, err := os.ReadFile(fullName)
				if err != nil {
					fmt.Printf("failed to read the file")
					return
				}

				packetSend := packet.Packet{Type: packet.PacketChunk, Chunk: packet.Chunk{Bytes: bytes}}

				conn, err := net.Dial("tcp", randomNode)
				if err != nil {
					fmt.Printf("failed to dial seeder %v\n", err)
					return
				}

				encoder := gob.NewEncoder(conn)
				err = encoder.Encode(packetSend)
				if err != nil {
					fmt.Printf("failed to send data to seeder\n")
					return
				}
			}

		}

	}
}

func (seeder *Seeder) HandleConnection(conn net.Conn, wg *sync.WaitGroup) {
	dec := gob.NewDecoder(conn)
	packetReceived := packet.Packet{}
	err := dec.Decode(&packetReceived)
	if err != nil {
		fmt.Printf("failed to decode packet, error %v\n", err)
		return
	}

	switch packetReceived.Type {
	case packet.PacketJoin:
		seeder.Nodes = append(seeder.Nodes, packetReceived.Join.Address)
	}

	wg.Done()
}
