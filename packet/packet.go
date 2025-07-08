package packet

import (
	"encoding/gob"
	"fmt"
	"net"
	"pour/seeder"
	"sync"
)

type PacketType int

const (
	PacketChunk PacketType = iota
	PacketJoin
)

type Chunk struct {
	Bytes []byte
}

type Join struct {
	Contributor bool
	Address string
}

type Packet struct {
	Type  PacketType
	Chunk Chunk
	Join  Join
}

func HandleConnection(conn net.Conn, seeder *seeder.Seeder, wg *sync.WaitGroup) {
	dec := gob.NewDecoder(conn)
	packetReceived := Packet{}
	err := dec.Decode(&packetReceived)
	if err != nil {
		fmt.Printf("failed to decode packet, error %v\n", err)
		return
	}

	switch packetReceived.Type {
	case PacketJoin:
		if packetReceived.Join.Contributor {
			seeder.Contributors = append(seeder.Contributors, packetReceived.Join.Address)
		}
	}

	wg.Done()
}
