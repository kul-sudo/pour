package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"pour/bootstrap"
	"pour/packet"
	"pour/seeder"
	"pour/workers"
	"sync"
)

func main() {
	config, err := bootstrap.ConfigGen()
	if err != nil {
		return
	}

	switch config.Mode {
	case "seeder":
		go workers.Segmentation()

		seeder := seeder.Seeder{make([]string, 0)}
		var wg sync.WaitGroup

		ln, err := net.Listen("tcp", config.Seeder.Address)
		if err != nil {
			fmt.Printf("failed to listen on port\n")
			return
		}

		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Printf("failed to accept connection\n")
				return
			}

			wg.Add(1)
			go packet.HandleConnection(conn, &seeder, &wg)
		}
	case "node":
		conn, err := net.Dial("tcp", config.Node.Seeder)
		if err != nil {
			fmt.Printf("failed to dial seeder %v\n", err)
			return
		}

		encoder := gob.NewEncoder(conn)

		packet := packet.Packet{Type: packet.PacketJoin, Join: packet.Join{Contributor: true, Address: config.Node.Address}}

		err = encoder.Encode(packet)
		if err != nil {
			fmt.Printf("failed to send data to seeder\n")
			return
		}
	}
	// http.ListenAndServe(":8080", nil)
}
