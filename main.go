package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"github.com/kul-sudo/packet"
)

const SEGMENT_DURATION = 3
const CONFIG_FILE = "config.json"

type Config struct {
	Mode string `json:"mode"`
	Node struct {
		Address      string `json:"address"`
		Seeder       string `json:"seeder"`
		Contribution int    `json:"contribution"`
	} `json:"node"`
	Seeder struct {
		Address string `json:"address"`
	} `json:"seeder"`
}

type Node struct {
	Chunks [][]byte
}

func segmentation() {
	cmd := exec.Command("ffmpeg", "-i", "rtmp://localhost:1935/live/playpath", "-c", "copy", "-f", "segment", "-segment_time", fmt.Sprintf("%d", SEGMENT_DURATION), "-reset_timestamps", "1", "%d.mp4")
	err := cmd.Run()
	if err != nil {
		fmt.Println("failed to run segmentation process")
		return
	}
}

func main() {
	configFile, err := os.Open(CONFIG_FILE)
	if err != nil {
		fmt.Println("failed to open config file, error: %v", err)
		return
	}

	defer configFile.Close()

	configData, err := io.ReadAll(configFile)
	if err != nil {
		fmt.Println("failed to read config file, error: %v", err)
		return
	}

	config := Config{}
	if err := json.Unmarshal(configData, &config); err != nil {
		fmt.Println("failed to unmarshal config file, error: %v", err)
		return
	}

	switch config.Mode {
	case "seeder":
		go segmentation()

		ln, err := net.Listen("tcp", config.Seeder.Address)
		if err != nil {
			fmt.Println("failed to listen on port %s", config.Seeder.Address)
			return
		}
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("failed to accept connection")
				return
			}

			dec := gob.NewDecoder(conn)
			packet := Packet{}
			err = dec.Decode(&packet)
		}
	case "node":
		conn, err := net.Dial("tcp", config.Node.Seeder)
		if err != nil {
			fmt.Println("failed to dial seeder %v", err)
			return
		}

		encoder := gob.NewEncoder(conn)

		packet := Packet{Type: PacketJoin, Join: Join{Address: config.Node.Address}}

		err = encoder.Encode(packet)
		if err != nil {
			fmt.Println("failed to send data to seeder")
			return
		}
	}
	// http.ListenAndServe(":8080", nil)
}
