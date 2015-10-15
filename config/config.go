// Copyright (c) 2015, Chao Wang <hit9@ele.me>

package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
)

type Config struct {
	Detector *SectionDetector `json:"detector"`
	Web      *SectionWeb      `json:"web"`
	Alerter  *SectionAlerter  `json:"alerter"`
}

type SectionDetector struct {
	Port        int      `json:"port"`
	DBFile      string   `json:"dbfile"`
	Factor      float64  `json:"factor"`
	Strict      bool     `json:"strict"`
	Periodicity [2]int   `json:"periodicity"`
	StartSize   int      `json:"start size"`
	WhiteList   []string `json:"whitelist"`
	BlackList   []string `json:"blacklist"`
}

type SectionWeb struct {
	Port   int    `json:"port"`
	Auth   string `json:"auth"`
	DBFile string `json:"dbfile"`
}

type SectionAlerter struct {
	DBFile  string `json:"dbfile"`
	Command string `json:"command"`
}

func NewWithDefaults() *Config {
	cfg := new(Config)
	cfg.Detector = new(SectionDetector)
	cfg.Web = new(SectionWeb)
	cfg.Alerter = new(SectionAlerter)
	cfg.Detector.Port = 9000
	cfg.Detector.DBFile = "stats.db"
	cfg.Detector.Factor = 0.07
	cfg.Detector.Strict = true
	cfg.Detector.Periodicity = [2]int{480, 180}
	cfg.Detector.StartSize = 32
	cfg.Detector.WhiteList = []string{"*"}
	cfg.Detector.BlackList = []string{"statsd.*"}
	cfg.Web.Port = 9001
	cfg.Web.Auth = "admin:admin"
	cfg.Web.DBFile = "rules.db"
	cfg.Alerter.Command = ""
	cfg.Alerter.DBFile = cfg.Web.DBFile
	return cfg
}

func NewWithJsonBytes(data []byte) (*Config, error) {
	cfg := NewWithDefaults()
	err := json.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}
	if err = Validate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func NewWithJsonFile(fileName string) (*Config, error) {
	log.Printf("reading config from %s..", fileName)
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	return NewWithJsonBytes(data)
}

func Validate(cfg *Config) error {
	if cfg.Detector.Factor >= 1.0 || cfg.Detector.Factor <= 0 {
		return ErrDetectorFactor
	}
	if len(cfg.Web.Auth) > 0 && len(strings.Split(cfg.Web.Auth, ":")) != 2 {
		return ErrWebAuth
	}
	return nil
}
