package packet

type PacketType uint

const (
	PacketChunk PacketType = iota
	PacketJoin
)

type Chunk struct {
	Bytes []byte
}

type Join struct {
	Address string
}

type Packet struct {
	Type  PacketType
	Chunk Chunk
	Join  Join
}
