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
	detector := flag.Bool("detector", false, "start detector")
	// webapp := flag.Bool("webapp", false, "start webapp")
	// alerter := flag.Bool("alerter", false, "start alerter")
	fileName := flag.String("config", "", "config path")
	version := flag.Bool("verion", false, "show version")
	flag.Parse()
	if flag.NFlag() != 2 {
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
	switch {
	case *detector:
		StartDetector(cfg.Detector)
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func StartDetector(cfg *config.SectionDetector) {
	detector := detector.New(cfg)
	detector.Start()
}
