// Copyright 2015. Chao Wang <hit9@icloud.com>
// Noise - Metric outliers detection.

package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
)

type Conf struct {
	PortIn  int // port to receive data
	PortOut int // port to output dara
}

var conf Conf

func main() {
	var confPath string
	flag.StringVar(&confPath, "conf", "conf.json", "conf file path")
	flag.Parse()
	log.Printf("reading config file from %s..", confPath)
	content, err := ioutil.ReadFile(confPath)
	if err != nil {
		log.Fatalf("failed to read file %s: %v", confPath, err)
	}
	err = json.Unmarshal(content, &conf)
	if err != nil {
		log.Fatalf("failed to parse json: %v", err)
	}
}
