package main

import (
	"bufio"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"math"
	"net"
	"path/filepath"
	"strings"
)

type App struct {
	cfg  *Config                  // app config
	db   *leveldb.DB              // leveldb handle
	outs map[*net.Conn]chan *Stat // output channels map
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

// Detect anomaly
func (app *App) Detect(stat *Stat) error {
	key, val, f := stat.Name, stat.Value, app.cfg.Factor
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
		avgNew = (1-f)*avgOld + f*val
		stdNew = math.Sqrt((1-f)*stdOld*stdOld + f*(val-avgOld)*(val-avgNew))
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
