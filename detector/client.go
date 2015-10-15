package detector

import (
	"bufio"
	"fmt"
	"net"
)

type Client struct {
	Host  string
	Port  int
	conn  net.Conn
	isPub bool
	isSub bool
}

func NewClientWithDefaults() *Client {
	client := new(Client)
	client.Host = "0.0.0.0"
	client.Port = 9000
	client.conn = nil
	client.isPub = false
	client.isSub = false
	return client
}

func NewClient(host string, port int) *Client {
	client := NewClientWithDefaults()
	client.Host = host
	client.Port = port
	return client
}

func (client *Client) Connect() (err error) {
	addr := fmt.Sprintf("%s:%d", client.Host, client.Port)
	client.conn, err = net.Dial("tcp", addr)
	return
}

func (client *Client) Close() (err error) {
	return client.conn.Close()
}

func (client *Client) Pub(stat *Stat) (err error) {
	if client.isSub {
		panic("Cannot pub in sub mode")
	}
	if client.conn == nil {
		err := client.Connect()
		if err != nil {
			return err
		}
	}
	if !client.isPub {
		client.conn.Write([]byte("pub\n"))
		client.isPub = true
	}
	_, err = client.conn.Write([]byte(stat.InputString()))
	return
}

func (client *Client) Sub(callback func(*Stat, error)) (err error) {
	if client.isPub {
		panic("Cannot sub in pub mode")
	}
	if client.conn == nil {
		err := client.Connect()
		if err != nil {
			return err
		}
	}
	if !client.isSub {
		client.conn.Write([]byte("sub\n"))
		client.isSub = true
	}
	scanner := bufio.NewScanner(client.conn)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return err
		}
		s := scanner.Text()
		stat, err := NewStatWithOutputString(s)
		callback(stat, err)
	}
	return
}
