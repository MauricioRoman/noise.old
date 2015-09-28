// Copyright (c) Chao Wang <hit9@icloud.com>
// Noise - Stats anomalies detection.

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"io/ioutil"
	"log"
	"math"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	ACTION_PUB = "pub"
	ACTION_SUB = "sub"
)

var (
	ErrInvalidInput     = errors.New("invalid input from pub end")
	ErrInvalidDBVal     = errors.New("invalid value is found in db")
	ErrInvalidCfgFactor = errors.New("invalid factor in config (require 0~1)")
)

type Config struct {
	Port        int      `json:"port"`
	Workers     int      `json:"workers"`
	DBPath      string   `json:"dbpath"`
	Factor      float64  `json:"factor"`
	Strict      bool     `json:"strict"`
	Periodicity [2]int   `json:"periodicity"`
	StartSize   int      `json:"start size"`
	WhiteList   []string `json: "whitelist"`
	BlackList   []string `json: "blacklist"`
}

type App struct {
	cfg  *Config                  // app config
	db   *leveldb.DB              // leveldb handle
	outs map[*net.Conn]chan *Stat // output channels map
}

type Stat struct {
	Name  string  // metric name
	Stamp int     // stat timestamp
	Value float64 // stat value
	Anoma float64 // stat anomalous factor
}

// Main
func main() {
	fileName := flag.String("c", "config.json", "config file")
	flag.Parse()
	if flag.NFlag() != 1 && flag.NArg() > 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}
	cfg, err := NewConfigWithJSONFile(*fileName)
	if err != nil {
		log.Fatalf("failed to read %s: %v", *fileName, err)
	}
	app := NewApp(cfg)
	app.Start()
}

// Create stat with default values
func NewStatWithDefaults() *Stat {
	stat := new(Stat)
	stat.Stamp = 0
	stat.Anoma = 0
	return stat
}

// Create stat with arguments
func NewStat(name string, stamp int, value float64) *Stat {
	stat := NewStatWithDefaults()
	stat.Name = name
	stat.Stamp = stamp
	stat.Value = value
	return stat
}

// Create stat with string.
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

// Dump stat into string
func (stat *Stat) String() string {
	return fmt.Sprintf("%s %d %.3f %.3f",
		stat.Name, stat.Stamp, stat.Value, stat.Anoma)
}

// Create config with default values
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

// Create app by config.
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

// Start server.
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
			log.Printf("action pub from conn %s", addr)
			app.HandlePub(conn)
		case ACTION_SUB:
			log.Printf("action sub from conn %s", addr)
			app.HandleSub(conn)
		default:
			log.Printf("action unknown from conn %s", addr)
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

func (app *App) GetKey(stat *Stat) string {
	offset := app.cfg.Periodicity[0]
	nOffset := app.cfg.Periodicity[1]
	periodicity := offset * nOffset
	suffix := (stat.Stamp % periodicity) / offset
	return fmt.Sprintf("%s-%d", stat.Name, suffix)
}

// Detect anomaly
func (app *App) Detect(stat *Stat) error {
	key := app.GetKey(stat)
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
