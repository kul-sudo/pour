package node		

import (
	"encoding/gob"
	"pour/bootstrap"
	"pour/dashboard"
	"pour/packet"
	"fmt"
	"net"
	"sync"
)

type Node struct {
	LatestChunk []byte
}

func Setup(config *bootstrap.Config) {
	node := Node{}

	conn, err := net.Dial("tcp", config.Node.Seeder)
	if err != nil {
		fmt.Printf("failed to dial seeder %v\n", err)
		return
	}
	
	page := dashboard.Page { Dashboard: config.Dashboard, LatestChunk: &node.LatestChunk }
	go dashboard.ShowNodeInfo(&page)
	
	packetSend := packet.Packet{Type: packet.PacketJoin, Join: packet.Join{Address: config.Node.Address}}
	encoder := gob.NewEncoder(conn)
	err = encoder.Encode(packetSend)
	if err != nil {
		fmt.Printf("failed to send data to seeder\n")
		return
	}

	var wg sync.WaitGroup

	ln, err := net.Listen("tcp", config.Node.Address)
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
		go node.HandleConnection(conn, config, &wg)
	}
}

func (node *Node) HandleConnection(conn net.Conn, config *bootstrap.Config, wg *sync.WaitGroup) {
	dec := gob.NewDecoder(conn)
	packetReceived := packet.Packet{}
	err := dec.Decode(&packetReceived)
	if err != nil {
		fmt.Printf("failed to decode packet, error %v\n", err)
		return
	}

	switch packetReceived.Type {
	case packet.PacketPassChunk:
		packet.PassChunkToNode(&packetReceived.PassChunk)
	case packet.PacketChunk:
		node.LatestChunk = packetReceived.Chunk.Bytes
	}

	wg.Done()
}
