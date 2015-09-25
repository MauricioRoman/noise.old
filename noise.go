// Copyright 2015. Chao Wang <hit9@icloud.com>
// Noise - Metric outliers detection.

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
)

const (
	COMM_PUB  = "pub"
	COMM_SUB  = "sub"
	DB_BUCKET = "noise"
)

type context struct {
	cfg  *config
	outs map[*net.Conn]chan []byte
}

type config struct {
	Port        int     `json:"port"`
	Workers     int     `json:"workers"`
	DBFile      string  `json:"dbfile"`
	Factor      float32 `json:"factor"`
	Strict      bool    `json:"strict"`
	Periodicity int     `json:"periodicity"`
}

type server struct {
	ctx *context
}

func main() {
	fileName := flag.String("c", "config.json", "config file")
	flag.Parse()

	if flag.NFlag() != 1 && flag.NArg() > 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	cfg, err := newConfig(*fileName)

	if err != nil {
		log.Fatalf("error to read %s: %v", *fileName, err)
	}

	ctx := newContext(cfg)
	ser := newServer(ctx)
	ser.serve()
}

func newConfig(fileName string) (*config, error) {
	cfg := new(config)
	cfg.Port = 9000
	cfg.Workers = 1
	cfg.DBFile = "noise.db"
	cfg.Factor = 0.06
	cfg.Strict = true
	cfg.Periodicity = 24 * 3600
	log.Printf("reading config from %s..", fileName)
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func newContext(cfg *config) *context {
	ctx := new(context)
	ctx.cfg = cfg
	ctx.outs = make(map[*net.Conn]chan []byte)
	return ctx
}

func newServer(ctx *context) *server {
	ser := new(server)
	ser.ctx = ctx
	return ser
}

func (ser *server) serve() {
	addr := fmt.Sprintf("0.0.0.0:%d", ser.ctx.cfg.Port)
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
		go ser.handle(conn)
	}
}

func (ser *server) handle(conn net.Conn) {
	addr := conn.RemoteAddr()
	log.Printf("conn %s established", addr)
	defer func() {
		conn.Close()
		log.Printf("conn %s disconnected", addr)
	}()
	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		if err := scanner.Err(); err != nil {
			log.Printf("error recv: %v, closing conn..", err)
			return
		}
		switch scanner.Text() {
		case COMM_PUB:
			log.Printf("conn %s action: pub", addr)
			ser.handlePub(conn)
		case COMM_SUB:
			log.Printf("conn %s action: sub", addr)
			ser.handleSub(conn)
		default:
			log.Printf("conn %s action: unkwn", addr)
		}
	}
}

func (ser *server) handlePub(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			log.Printf("error recv: %v, closing conn..", err)
			break
		}
		s := scanner.Text()
		for _, out := range ser.ctx.outs {
			out <- []byte(s)
		}
	}
}

func (ser *server) handleSub(conn net.Conn) {
	ser.ctx.outs[&conn] = make(chan []byte)
	defer delete(ser.ctx.outs, &conn)
	for {
		bytes := <-ser.ctx.outs[&conn]
		bytes = append(bytes, '\n')
		_, err := conn.Write(bytes)
		if err != nil {
			break
		}
	}
}
