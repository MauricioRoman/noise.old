package main

import (
	"bufio"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"math"
	"net"
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
			log.Printf("conn %s action: pub", addr)
			app.HandlePub(conn)
		case ACTION_SUB:
			log.Printf("conn %s action: sub", addr)
			app.HandleSub(conn)
		default:
			log.Printf("conn %s action: unkwn", addr)
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

// Detect anomaly
func (app *App) Detect(stat *Stat) error {
	key, val, f := stat.Name, stat.Value, app.cfg.Factor
	data, err := app.db.Get([]byte(key), nil)

	if err != nil && err != leveldb.ErrNotFound {
		return err
	}
	var avgOld, stdOld, avgNew, stdNew float64

	if data == nil || err == leveldb.ErrNotFound {
		avgNew = val
		stdNew = 0
	} else {
		n, err := fmt.Sscanf(string(data), "%f %f", &avgOld, &stdOld)
		if err != nil || n != 2 {
			return ErrInvalidDBVal
		}
		avgNew = (1-f)*avgOld + f*val
		stdNew = math.Sqrt((1-f)*stdOld*stdOld + f*(val-avgOld)*(val-avgNew))
	}

	dataNew := []byte(fmt.Sprintf("%.5f %.5f", avgNew, stdNew))
	err = app.db.Put([]byte(key), dataNew, nil)
	if err != nil {
		return err
	}
	stat.Anoma = (val - avgNew) / float64(3.0*stdNew)
	return nil
}
