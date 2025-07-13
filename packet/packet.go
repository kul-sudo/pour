package packet

import (
	"encoding/gob"
	"fmt"
	"net"
)

type PacketType int

const (
	PacketChunk PacketType = iota
	PacketPassChunk
	PacketJoin
)

type Chunk struct {
	Bytes []byte
}

type PassChunk struct {
	Chunk Chunk
	DestinationAddress string
}

type Join struct {
	Address string
}

type Packet struct {
	Type  PacketType
	Chunk Chunk
	PassChunk PassChunk
	Join  Join
}

func PassChunkToNode(data *PassChunk) {
	conn, err := net.Dial("tcp", data.DestinationAddress)
	if err != nil {
		fmt.Printf("failed to dial seeder %v\n", err)
		return
	}

	packetSend := Packet{Type: PacketChunk, Chunk: Chunk{Bytes: data.Chunk.Bytes}}
	encoder := gob.NewEncoder(conn)
	err = encoder.Encode(packetSend)
	if err != nil {
		fmt.Printf("failed to send data to seeder\n")
		return
	}
}
