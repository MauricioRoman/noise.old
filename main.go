package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/eleme/noise/config"
	"github.com/eleme/noise/detector"
)

func main() {
	fileName := flag.String("c", "config.json", "config path")
	version := flag.Bool("v", false, "show version")
	flag.Parse()
	if flag.NArg() > 0 && flag.NFlag() != 1 {
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *version {
		fmt.Fprintln(os.Stderr, VERSION)
		os.Exit(1)
	}
	cfg, err := config.NewWithJsonFile(*fileName)
	if err != nil {
		log.Fatalf("failed to read %s: %v", *fileName, err)
	}
	StartDetector(cfg.Detector)
}

func StartDetector(cfg *config.SectionDetector) {
	detector := detector.New(cfg)
	detector.Start()
}
