package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Mappings map[string]string `json:mappings`
}

func LoadConfig(path string) (cfg Config, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	return
}
