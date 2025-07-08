package main

import (
	"pour/bootstrap"
	"pour/node"
	"pour/seeder"
)

func main() {
	config, err := bootstrap.ConfigGen()
	if err != nil {
		return
	}

	switch config.Mode {
	case "seeder":
		seeder.Setup(&config)
	case "node":
		node.Setup(&config)
	}
}
