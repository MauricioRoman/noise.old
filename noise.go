// Copyright 2015. Chao Wang <hit9@icloud.com>
// Noise - Metric outliers detection.

package main

import (
	"flag"
	"log"
	"os"
)

func main() {
	fileName := flag.String("c", "config.json", "config file")
	flag.Parse()

	if flag.NFlag() != 1 && flag.NArg() > 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	cfg, err := NewConfigWithFile(*fileName)

	if err != nil {
		log.Fatalf("error to read %s: %v", *fileName, err)
	}

	app := NewApp(cfg)
	app.Serve()
}
