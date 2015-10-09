// Go client implemention for github.com/eleme/noise.
package noise

import (
	"bufio"
	"fmt"
	"net"
)

type Noise struct {
	Host  string   // noise server host, default: '0.0.0.0'
	Port  int      // noise server port, default: 9000
	conn  net.Conn // connection to remote server
	isPub bool     // is in pub mode?
	isSub bool     // is in sub mode?
}

// Create noise client with default values.
func NewNoiseClientWithDefaults() *Noise {
	noise = new(Noise)
	noise.Host = "0.0.0.0"
	noise.Port = 9000
	noise.conn = nil
	noise.isPub = false
	noise.isSub = false
	return noise
}

// Create noise client with host and port as arguments.
func NewNoiseClient(host string, port int) *Noise {
	noise = NewStatWithDefaults()
	noise.host = host
	noise.port = port
}

// Connect to noise server.
func (noise *Noise) Connect() (err error) {
	addr := fmt.Sprintf("%s:%d", noise.host, noise.port)
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
			return
		}
	}
	if !noise.isPub {
		noise.conn.Write("pub\n")
		noise.isPub = true
	}
	s := fmt.Sprintf("%s %d %.5f\n", name, stamp, value)
	_, err = noise.conn.Write(s)
	return
}

// Subscribe anomalies from noise.
func (noise *Noise) Sub(onAnomaly func(string, int, float64, float64)) (err error) {
	if noise.isPub {
		panic("Cannot sub in pub mode")
	}
	if noise.conn == nil {
		err := noise.Connect()
		if err != nil {
			return
		}
	}
	if !noise.isSub {
		noise.conn.Write("sub\n")
		noise.isSub = true
	}
	var name string
	var stamp int
	var value float64
	var anoma float64
	scanner := bufio.NewScanner(noise.conn)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return
		}
		s := scanner.Text()
		n, err := fmt.Sscanf(s, "%s %d %f %f", &name, &stamp, &value, &anoma)
		if err != nil || n != 4 {
			return
		}
		onAnomaly(name, stamp, value, anoma)
	}
	return
}
