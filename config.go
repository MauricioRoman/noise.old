package main

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Port        int     `json:"port"`
	Workers     int     `json:"workers"`
	DBFile      string  `json:"dbfile"`
	Factor      float32 `json:"factor"`
	Strict      bool    `json:"strict"`
	Periodicity int     `json:"periodicity"`
}

// Create config with default values
func NewConfigWithDefaults() *Config {
	cfg := new(Config)
	cfg.Port = 9000
	cfg.Workers = 1
	cfg.DBFile = "noise.db"
	cfg.Factor = 0.06
	cfg.Strict = true
	cfg.Periodicity = 24 * 3600
	return cfg
}

// Create config from json bytes.
func NewConfigWithJSONBytes(data []byte) (*Config, error) {
	cfg := NewConfigWithDefaults()
	err := json.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// Create config from json file.
func NewConfigWithJSONFile(fileName string) (*Config, error) {
	log.Printf("reading config from %s..", fileName)
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	return NewConfigWithJSONBytes(data)
}
