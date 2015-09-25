// Copyright 2015. Chao Wang <hit9@icloud.com>
// Noise - Metric outliers detection.

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/boltdb/bolt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	COMM_PUB  = "pub"
	COMM_SUB  = "sub"
	DB_BUCKET = "noise"
)

type context struct {
	cfg  *config
	db   *bolt.DB
	outs map[*net.Conn]chan *stat
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

type stat struct {
	name  string
	stamp int
	value float32
	multi float32
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
	ctx, err := newContext(cfg)
	if err != nil {
		log.Fatalf("error to open db %s: %v", cfg.DBFile, err)
	}
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

func newContext(cfg *config) (*context, error) {
	ctx := new(context)
	ctx.cfg = cfg
	ctx.outs = make(map[*net.Conn]chan *stat)
	db, err := bolt.Open(cfg.DBFile, 0600, nil)
	if err != nil {
		return nil, err
	}
	ctx.db = db
	return ctx, nil
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
		st, err := load(s)
		if err != nil {
			log.Printf("invalid input: %s", s)
			continue
		}
		for _, out := range ser.ctx.outs {
			out <- st
		}
	}
}

func (ser *server) handleSub(conn net.Conn) {
	ser.ctx.outs[&conn] = make(chan *stat)
	defer delete(ser.ctx.outs, &conn)
	for {
		st := <-ser.ctx.outs[&conn]
		bytes := []byte(dump(st))
		bytes = append(bytes, '\n')
		_, err := conn.Write(bytes)
		if err != nil {
			break
		}
	}
}

func load(s string) (*stat, error) {
	st := new(stat)
	list := strings.Fields(s)
	name := list[0]
	stamp, err := strconv.Atoi(list[1])
	if err != nil {
		return nil, err
	}
	value, err := strconv.ParseFloat(list[2], 32)
	if err != nil {
		return nil, err
	}
	st.name = name
	st.stamp = stamp
	st.value = float32(value)
	st.multi = 0
	return st, nil
}

func dump(st *stat) string {
	return fmt.Sprintf("%s %d %.3f %.3f", st.name, st.stamp, st.value, st.multi)
}
