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
)

type Seeder struct {
	Contributors []string
}

func Setup(config *bootstrap.Config) {
	go workers.Segmentation()

	seeder := Seeder{make([]string, 0)}

	var wg sync.WaitGroup

	page := dashboard.Page { Contributors: &seeder.Contributors, Dashboard: config.Dashboard  }
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
		if packetReceived.Join.Contributor {
			seeder.Contributors = append(seeder.Contributors, packetReceived.Join.Address)
		}
	}

	wg.Done()
}
