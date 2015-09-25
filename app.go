// Copyright 2015. Chao Wang <hit9@icloud.com>

package main

import (
	"fmt"
	"log"
	"net"
)

type App struct {
	cfg      *Config
	pubConns []net.Conn
	subConns []net.Conn
}

func NewApp(cfg *Config) *App {
	app := new(App)
	app.cfg = cfg
	return app
}

func (app *App) Serve() {
	addr := fmt.Sprintf("0.0.0.0:%d", app.cfg.Port)
	listener, err := net.Listen("tcp", addr)

	if err != nil {
		log.Fatalf("failed to bind %s", addr)
	} else {
		log.Printf("listening on tcp://%s..", addr)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("failed to accept new conn")
		}
		go app.handle(conn)
	}
}

func (app *App) handle(conn net.Conn) {
	fmt.Printf("new connection")
}
