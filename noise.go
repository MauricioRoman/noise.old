// Copyright 2015. Chao Wang <hit9@icloud.com>
// Noise - Metric outliers detection.

package main

import (
	"encoding/json"
	"flag"
	// "fmt"
	"io/ioutil"
	"log"
	"os"
	// "net"
)

type Conf struct {
	InPort  int // input tcp port
	OutPort int // output tco port
}

var conf Conf

func main() {
	var path string
	flag.StringVar(&path, "c", "conf.json", "conf path")
	flag.Parse()
	if flag.NFlag() != 1 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	log.Printf("reading conf from %s..", path)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read %s: %v", path, err)
	}

	err = json.Unmarshal(content, &conf)
	if err != nil {
		log.Fatalf("failed to parse conf: %v", err)
	}
}
