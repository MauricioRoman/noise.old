package detector

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"net"
	"path/filepath"
	"strings"
	"time"

	"github.com/eleme/noise/config"
	"github.com/syndtr/goleveldb/leveldb"
)

type Detector struct {
	cfg  *config.SectionDetector
	db   *leveldb.DB
	outs map[*net.Conn]chan *Stat
}

func New(cfg *config.SectionDetector) *Detector {
	db, err := leveldb.OpenFile(cfg.DBFile, nil)
	if err != nil {
		log.Fatalf("failed to open %s: %v", cfg.DBFile, err)
	}
	detector := new(Detector)
	detector.cfg = cfg
	detector.db = db
	detector.outs = make(map[*net.Conn]chan *Stat)
	return detector
}

func (detector *Detector) Start() {
	addr := fmt.Sprintf("0.0.0.0:%d", detector.cfg.Port)
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
		go detector.Handle(conn)
	}
}

func (detector *Detector) Handle(conn net.Conn) {
	addr := conn.RemoteAddr()
	defer func() {
		conn.Close()
		log.Printf("conn %s disconnected", addr)
	}()
	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		if err := scanner.Err(); err != nil {
			log.Printf("failed to read conn: %v, closing it..", err)
			return
		}
		s := scanner.Text()
		s = strings.TrimSpace(s)
		switch strings.ToLower(s) {
		case "pub":
			log.Printf("conn %s established: pub", addr)
			detector.HandlePub(conn, scanner)
		case "sub":
			log.Printf("conn %s established: sub", addr)
			detector.HandleSub(conn)
		}
	}
}

func (detector *Detector) HandleSub(conn net.Conn) {
	detector.outs[&conn] = make(chan *Stat)
	defer delete(detector.outs, &conn)
	for {
		stat := <-detector.outs[&conn]
		bytes := []byte(stat.String())
		bytes = append(bytes, '\n')
		_, err := conn.Write(bytes)
		if err != nil {
			log.Printf("failed to write conn: %v", err)
			break
		}
	}
}

func (detector *Detector) HandlePub(conn net.Conn, scanner *bufio.Scanner) {
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			log.Printf("failed to read con: %v, closing it..", err)
			break
		}
		s := scanner.Text()
		stat, err := NewStatWithString(s)
		if err != nil {
			log.Printf("invalid stat input: %s, skipping..", s)
			continue
		}
		startAt := time.Now()
		if !detector.Match(stat) {
			continue
		}
		err = detector.Detect(stat)
		if err != nil {
			log.Printf("failed to detect %s: %v, skipping..", stat.Name, err)
			continue
		}
		elapsed := time.Since(startAt)
		ms := float64(elapsed.Nanoseconds()) / float64(1000*1000)
		log.Printf("%.2fms %s", ms, stat.String())
		for _, out := range detector.outs {
			if math.Abs(stat.Anoma) >= 1.0 {
				out <- stat
			}
		}
	}
}

func (detector *Detector) Match(stat *Stat) bool {
	for _, pattern := range detector.cfg.BlackList {
		matched, err := filepath.Match(pattern, stat.Name)
		if err != nil {
			log.Printf("invalid pattern in blackList: %s, %v, skipping..", pattern, err)
			continue
		}
		if matched {
			return false
		}
	}
	for _, pattern := range detector.cfg.WhiteList {
		matched, err := filepath.Match(pattern, stat.Name)
		if err != nil {
			log.Printf("invalid pattern in whitelist: %s, %v, skipping..", pattern, err)
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

func (detector *Detector) Detect(stat *Stat) error {
	f := detector.cfg.Factor
	strict := detector.cfg.Strict
	startSize := detector.cfg.StartSize
	v := stat.Value
	key := detector.GetDBKey(stat)
	var (
		avgNew float64 = 0
		stdNew float64 = 0
		result float64 = 0
		numNew int     = 0
	)
	avgOld, stdOld, numOld, err := detector.GetDBData(key)
	if err != nil && err != ErrDBKeyNotFound {
		return err
	}
	if err != ErrDBKeyNotFound {
		if !strict {
			v = (v + avgOld) / float64(2)
		}
		avgNew = (1-f)*avgOld + f*v
		stdNew = math.Sqrt((1-f)*stdOld*stdOld + f*(v-avgOld)*(v-avgNew))
		if numOld < startSize {
			numNew = numOld + 1
			result = 0
		} else {
			numNew = numOld
			result = (v - avgNew) / (3 * stdNew)
			if math.IsNaN(result) {
				result = 0
			}
		}
	} else {
		avgOld = 0
		stdOld = 0
	}
	if err = detector.PutDBData(key, avgNew, stdNew, numNew); err != nil {
		return err
	}
	stat.Anoma = result
	stat.AvgOld = avgOld
	stat.AvgNew = avgNew
	return nil
}

func (detector *Detector) GetDBKey(stat *Stat) string {
	grid := detector.cfg.Periodicity[0]
	numGrid := detector.cfg.Periodicity[1]
	periodicity := grid * numGrid
	gridNo := (stat.Stamp % periodicity) / grid
	return fmt.Sprintf("stat-%dx%d-%d:%s", grid, numGrid, gridNo, stat.Name)
}

func (detector *Detector) GetDBData(key string) (avg float64, std float64, num int, err error) {
	data, err := detector.db.Get([]byte(key), nil)
	if err != leveldb.ErrNotFound || data == nil {
		err = ErrDBKeyNotFound
		return
	}
	if err != nil {
		return
	}
	n, err := fmt.Sscanf(string(data), "%f %f %d", &avg, &std, &num)
	if err != nil || n != 3 {
		err = ErrInvalidDBData
		return
	}
	return
}

func (detector *Detector) PutDBData(key string, avg float64, std float64, num int) error {
	s := fmt.Sprintf("%.5f %.5f %d", avg, std, num)
	return detector.db.Put([]byte(key), []byte(s), nil)
}
