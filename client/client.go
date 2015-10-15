package client

import (
	"bufio"
	"fmt"
	"net"
)

type Noise struct {
	Host  string
	Port  int
	conn  net.Conn
	isPub bool
	isSub bool
}

// Create noise client with default values.
func NewNoiseClientWithDefaults() *Noise {
	noise := new(Noise)
	noise.Host = "0.0.0.0"
	noise.Port = 9000
	noise.conn = nil
	noise.isPub = false
	noise.isSub = false
	return noise
}

// Create noise client with host and port as arguments.
func NewNoiseClient(host string, port int) *Noise {
	noise := NewNoiseClientWithDefaults()
	noise.Host = host
	noise.Port = port
	return noise
}

// Connect to noise server.
func (noise *Noise) Connect() (err error) {
	addr := fmt.Sprintf("%s:%d", noise.Host, noise.Port)
	noise.conn, err = net.Dial("tcp", addr)
	return
}

// Close the connection.
func (noise *Noise) Close() (err error) {
	return noise.conn.Close()
}

// Publish stats to noise.
func (noise *Noise) Pub(name string, stamp int, value float64) (err error) {
	if noise.isSub {
		panic("Cannot pub in sub mode")
	}
	if noise.conn == nil {
		err := noise.Connect()
		if err != nil {
			return err
		}
	}
	if !noise.isPub {
		noise.conn.Write([]byte("pub\n"))
		noise.isPub = true
	}
	s := fmt.Sprintf("%s %d %.5f\n", name, stamp, value)
	_, err = noise.conn.Write([]byte(s))
	return
}

// Subscribe anomalies from noise.
func (noise *Noise) Sub(onAnomaly func(string, int, float64, float64, float64, float64)) (err error) {
	if noise.isPub {
		panic("Cannot sub in pub mode")
	}
	if noise.conn == nil {
		err := noise.Connect()
		if err != nil {
			return err
		}
	}
	if !noise.isSub {
		noise.conn.Write([]byte("sub\n"))
		noise.isSub = true
	}
	var name string
	var stamp int
	var value float64
	var anoma float64
	var avgOld float64
	var avgNew float64
	scanner := bufio.NewScanner(noise.conn)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return err
		}
		s := scanner.Text()
		n, err := fmt.Sscanf(s, "%s %d %f %f %f %f ", &name, &stamp, &value,
			&anoma, &avgOld, &avgNew)
		if err != nil || n != 6 {
			return err
		}
		onAnomaly(name, stamp, value, anoma, avgOld, avgNew)
	}
	return
}
