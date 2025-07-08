package bootstrap

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const CONFIG_FILE = "config.json"

type Config struct {
	Mode string `json:"mode"`
	Node struct {
		Address     string `json:"address"`
		Seeder      string `json:"seeder"`
		Contributor bool   `json:"contributor"`
	} `json:"node"`
	Seeder struct {
		Address string `json:"address"`
	} `json:"seeder"`
}

func ConfigGen() (Config, error) {
	configFile, err := os.Open(CONFIG_FILE)
	if err != nil {
		fmt.Printf("failed to open config file, %v\n", err)
		return Config{}, err
	}

	defer configFile.Close()

	configData, err := io.ReadAll(configFile)
	if err != nil {
		fmt.Printf("failed to read config file, %v\n", err)
		return Config{}, err
	}

	config := Config{}
	if err := json.Unmarshal(configData, &config); err != nil {
		fmt.Printf("failed to unmarshal config file, %v\n", err)
		return Config{}, err
	}

	return config, nil
}
