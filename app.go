package main

import (
	"bufio"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"net"
	"strings"
)

type App struct {
	cfg  *Config                  // app config
	db   *bolt.DB                 // bold db
	outs map[*net.Conn]chan *Stat // output channels map
}

// Create app by config.
func NewApp(cfg *Config) *App {
	db, err := bolt.Open(cfg.DBFile, 0600, nil)
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
		go app.handle(conn)
	}
}

// Handle connection request
func (app *App) handle(conn net.Conn) {
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
			app.handlePub(conn)
		case ACTION_SUB:
			log.Printf("conn %s action: sub", addr)
			app.handleSub(conn)
		default:
			log.Printf("conn %s action: unkwn", addr)
		}
	}
}

// Handle connection request for pub
func (app *App) handlePub(conn net.Conn) {
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
		for _, out := range app.outs {
			out <- stat
		}
	}
}

// Handle connection request for sub
func (app *App) handleSub(conn net.Conn) {
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
