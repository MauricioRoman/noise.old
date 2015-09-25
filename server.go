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

type Server struct {
	cfg *Config
	chs map[*net.Conn]chan []byte
}

func NewServer(cfg *Config) *Server {
	server := new(Server)
	server.cfg = cfg
	server.chs = make(map[*net.Conn]chan []byte)
	return server
}

func (server *Server) Serve() {
	addr := fmt.Sprintf("0.0.0.0:%d", server.cfg.Port)
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
		go server.handle(conn)
	}
}

func (server *Server) handle(conn net.Conn) {
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
			server.handlePub(conn)
		case COMM_SUB:
			log.Printf("conn %s action: sub", addr)
			server.handleSub(conn)
		default:
			log.Printf("conn %s action: unkwn", addr)
		}
	}

}

func (server *Server) handlePub(conn net.Conn) {
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			log.Printf("error on recv: %v, closing conn..", err)
			break
		}
		s := scanner.Text()

		for _, ch := range server.chs {
			ch <- []byte(s)
		}
	}
}

func (server *Server) handleSub(conn net.Conn) {
	server.chs[&conn] = make(chan []byte)
	defer delete(server.chs, &conn)
	for {
		bytes := <-server.chs[&conn]
		bytes = append(bytes, '\n')
		_, err := conn.Write(bytes)
		if err != nil {
			break
		}
	}
}
