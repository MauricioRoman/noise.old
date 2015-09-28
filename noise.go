// Copyright (c) 2015, Chao Wang <hit9@icloud.com>
// All rights reserved.
//
// Noise is a simple daemon to detect anomalous stats.

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
)

const (
	ACTION_PUB = "pub" // detail command of action pub
	ACTION_SUB = "sub" // detail command of action sub
)

var (
	ErrInvalidInput     = errors.New("invalid input from pub end")
	ErrInvalidDBVal     = errors.New("invalid value is found in db")
	ErrInvalidCfgFactor = errors.New("invalid factor in config (require 0~1)")
)

type Config struct {
	Port        int      `json:"port"`        // port to bind
	Workers     int      `json:"workers"`     // number of workers to start
	DBPath      string   `json:"dbpath"`      // leveldb dir path
	Factor      float64  `json:"factor"`      // weighted moving average factor
	Strict      bool     `json:"strict"`      // if weaken latest stat value
	Periodicity [2]int   `json:"periodicity"` // metric periodicity: grid * numGrid
	StartSize   int      `json:"start size"`  // start detecting minimum stats count
	WhiteList   []string `json: "whitelist"`  // allow passing pattern list
	BlackList   []string `json: "blacklist"`  // disallow passing pattern list
}

type App struct {
	cfg  *Config                  // ref of app config
	db   *leveldb.DB              // ref of leveldb handle
	outs map[*net.Conn]chan *Stat // output channels map
}

type Stat struct {
	Name  string  // stat name, e.g. timer.count_ps.api
	Stamp int     // stat timestamp, in second e.g. 1412762335
	Value float64 // stat value, in float, e.g. 14.3
	Anoma float64 // stat anomalous factor, e.g. 1.2 (abs>=1 => anomaly)
}

// Main entry.
func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "help: noise <path/to/config.json>")
		os.Exit(1)
	}
	fileName := flag.Args()[0]
	cfg, err := NewConfigWithJSONFile(fileName)
	if err != nil {
		log.Fatalf("failed to open %s: %v", fileName, err)
	}
	app := NewApp(cfg)
	app.Start()
}

// Create stat with default values.
func NewStatWithDefaults() *Stat {
	stat := new(Stat)
	stat.Stamp = 0
	stat.Anoma = 0
	return stat
}

// Create stat with arguments.
func NewStat(name string, stamp int, value float64) *Stat {
	stat := NewStatWithDefaults()
	stat.Name = name
	stat.Stamp = stamp
	stat.Value = value
	return stat
}

// Create stat with protocol like string.
func NewStatWithString(s string) (*Stat, error) {
	words := strings.Fields(s)
	if len(words) != 3 {
		return nil, ErrInvalidInput
	}
	name := words[0]
	stamp, err := strconv.Atoi(words[1])
	if err != nil {
		return nil, err
	}
	value, err := strconv.ParseFloat(words[2], 32)
	if err != nil {
		return nil, err
	}
	return NewStat(name, stamp, value), nil
}

// Dump or format stat into string.
func (stat *Stat) String() string {
	return fmt.Sprintf("%s %d %.3f %.3f",
		stat.Name, stat.Stamp, stat.Value, stat.Anoma)
}

// Create config with default values.
func NewConfigWithDefaults() *Config {
	cfg := new(Config)
	cfg.Port = 9000
	cfg.Workers = 1
	cfg.DBPath = "noise.db"
	cfg.Factor = 0.06
	cfg.Strict = true
	cfg.Periodicity = [2]int{240, 360}
	cfg.StartSize = 23
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

// Create an app instance by config.
func NewApp(cfg *Config) *App {
	db, err := leveldb.OpenFile(cfg.DBPath, nil)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	app := new(App)
	app.cfg = cfg
	app.outs = make(map[*net.Conn]chan *Stat)
	app.db = db
	return app
}

// Start app server.
func (app *App) Start() {
	addr := fmt.Sprintf("0.0.0.0:%d", app.cfg.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to bind %s: %v", addr, err)
	}
	log.Printf("listening on %s..", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("failed to accept new conn: %v", err)
		}
		go app.Handle(conn)
	}
}

// Handle connection request
func (app *App) Handle(conn net.Conn) {
	addr := conn.RemoteAddr()
	log.Printf("conn %s established", addr)

	defer func() {
		conn.Close()
		log.Printf("conn %s disconnected", addr)
	}()

	scanner := bufio.NewScanner(conn)

	if scanner.Scan() {
		if err := scanner.Err(); err != nil {
			log.Printf("failed to read data: %v, closing conn..", err)
			return
		}
		s := scanner.Text()
		switch strings.ToLower(s) {
		case ACTION_PUB:
			log.Printf("conn %s action pub", addr)
			app.HandlePub(conn)
		case ACTION_SUB:
			log.Printf("conn %s action sub", addr)
			app.HandleSub(conn)
		default:
			log.Printf("conn %s action unknown", addr)
		}
	}
}

// Handle connection request for pub
func (app *App) HandlePub(conn net.Conn) {
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			log.Printf("failed to read data: %v, closing conn..", err)
			break
		}
		s := scanner.Text()
		stat, err := NewStatWithString(s)
		if err != nil {
			log.Printf("invalid input, skipping..")
			continue
		}

		if !app.Match(stat) {
			log.Println("not match")
			continue
		}

		err = app.Detect(stat)
		if err != nil {
			log.Printf("failed to detect %s: %v, skipping..", stat.Name, err)
			continue
		}
		for _, out := range app.outs {
			if math.Abs(stat.Anoma) >= 1.0 {
				out <- stat
			}
		}
	}
}

// Handle connection request for sub
func (app *App) HandleSub(conn net.Conn) {
	app.outs[&conn] = make(chan *Stat)
	defer delete(app.outs, &conn)
	for {
		stat := <-app.outs[&conn]
		bytes := []byte(stat.String())
		bytes = append(bytes, '\n')
		_, err := conn.Write(bytes)
		if err != nil {
			log.Printf("failed to write conn: %v", err)
			break
		}
	}
}

// Match the whitelist and blacklist.
func (app *App) Match(stat *Stat) bool {
	wl := app.cfg.WhiteList
	bl := app.cfg.BlackList
	for i := 0; i < len(wl); i++ {
		matched, err := filepath.Match(wl[i], stat.Name)
		if err != nil {
			log.Printf("!bad whitelist pattern: %s, %v, skipping..",
				wl[i], err)
			continue
		}
		if matched {
			for j := 0; j < len(bl); j++ {
				matched, err := filepath.Match(bl[j], stat.Name)
				if err != nil {
					log.Printf("!bad blacklist pattern: %s, %v, skipping..",
						bl[j], err)
					continue
				}
				if matched {
					return false
				}
			}
			return true
		}
	}
	return false
}

// Detect if this stat is an anomaly
func (app *App) Detect(stat *Stat) error {
	key := app.getDBKey(stat)
	val := stat.Value
	fct := app.cfg.Factor

	data, err := app.db.Get([]byte(key), nil)
	if err != nil && err != leveldb.ErrNotFound {
		return err
	}
	var avgOld, stdOld, avgNew, stdNew float64
	var numOld, numNew int
	var anoma float64

	if data == nil || err == leveldb.ErrNotFound {
		avgNew = val
		stdNew = 0
		numNew = 0
	} else {
		n, err := fmt.Sscanf(string(data), "%f %f %d", &avgOld, &stdOld, &numOld)
		if err != nil || n != 3 {
			return ErrInvalidDBVal
		}
		if !app.cfg.Strict {
			val = (val + avgOld) / float64(2)
		}
		avgNew = (1-fct)*avgOld + fct*val
		stdNew = math.Sqrt((1-fct)*stdOld*stdOld + fct*(val-avgOld)*(val-avgNew))
		if numOld < app.cfg.StartSize {
			numNew = numOld + 1
			anoma = 0
		} else {
			numNew = numOld
			anoma = (val - avgNew) / float64(3*stdNew)
		}
	}
	dataNew := []byte(fmt.Sprintf("%.5f %.5f %d", avgNew, stdNew, numNew))
	err = app.db.Put([]byte(key), dataNew, nil)
	if err != nil {
		return err
	}
	stat.Anoma = anoma
	return nil
}

// Get leveldb key by stat name and timestamp, per periodicity is divided
// into multiple grids, with each grid shares the same time span, this
// function will find the grid for this stat by its stamp. By this way,
// only the stats on the same phase will be considered.
func (app *App) getDBKey(stat *Stat) string {
	grid := app.cfg.Periodicity[0]
	numGrids := app.cfg.Periodicity[1]
	periodicity := grid * numGrids
	gridNo := (stat.Stamp % periodicity) / grid
	return fmt.Sprintf("%s:%d", stat.Name, gridNo)
}
