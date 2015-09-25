// Copyright 2015. Chao Wang <hit9@icloud.com>

package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

const (
	COMM_PUB = "pub"
	COMM_SUB = "sub"
)

type App struct {
	cfg *Config
	chs map[*net.Conn]chan []byte
}

func NewApp(cfg *Config) *App {
	app := new(App)
	app.cfg = cfg
	app.chs = make(map[*net.Conn]chan []byte)
	return app
}

func (app *App) Serve() {
	addr := fmt.Sprintf("0.0.0.0:%d", app.cfg.Port)
	listener, err := net.Listen("tcp", addr)

	if err != nil {
		log.Fatalf("error to bind %s", addr)
	}

	log.Printf("listening on %s..", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("error to accept new conn")
		}
		go app.handle(conn)
	}
}

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
			log.Printf("error on recv: %v, closing conn..", err)
			return
		}

		switch scanner.Text() {
		case COMM_PUB:
			log.Printf("conn %s action: pub", addr)
			app.handlePub(conn)
		case COMM_SUB:
			log.Printf("conn %s action: sub", addr)
			app.handleSub(conn)
		default:
			log.Printf("conn %s action: unkwn", addr)
		}
	}

}

func (app *App) handlePub(conn net.Conn) {
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			log.Printf("error on recv: %v, closing conn..", err)
			break
		}
		s := scanner.Text()

		for _, ch := range app.chs {
			ch <- []byte(s)
		}
	}
}

func (app *App) handleSub(conn net.Conn) {
	app.chs[&conn] = make(chan []byte)
	defer delete(app.chs, &conn)
	for {
		bytes := <-app.chs[&conn]
		bytes = append(bytes, '\n')
		_, err := conn.Write(bytes)
		if err != nil {
			break
		}
	}
}
