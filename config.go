package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Config struct {
	Port        int      `json:"port"`
	Workers     int      `json:"workers"`
	DBPath      string   `json:"dbpath"`
	Factor      float64  `json:"factor"`
	Strict      bool     `json:"strict"`
	Periodicity int      `json:"periodicity"`
	StartSize   int      `json:"start size"`
	WhiteList   []string `json: "whitelist"`
	BlackList   []string `json: "blacklist"`
}

// Create config with default values
func NewConfigWithDefaults() *Config {
	cfg := new(Config)
	cfg.Port = 9000
	cfg.Workers = 1
	cfg.DBPath = "noise.db"
	cfg.Factor = 0.06
	cfg.Strict = true
	cfg.Periodicity = 24 * 3600
	cfg.StartSize = 50
	cfg.WhiteList = []string{"*"}
	cfg.BlackList = []string{"statsd.*"}
	return cfg
}

// Create config from json bytes.
func NewConfigWithJSONBytes(data []byte) (*Config, error) {
	cfg := NewConfigWithDefaults()
	err := json.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}
	if cfg.Factor >= 1.0 || cfg.Factor <= 0 {
		return nil, ErrInvalidCfgFactor
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
