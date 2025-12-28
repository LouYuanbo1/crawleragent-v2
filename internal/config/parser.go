package config

import (
	"encoding/json"
)

func ParseConfig(byteConfig []byte) (*Config, error) {
	var cfg Config
	err := json.Unmarshal(byteConfig, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
