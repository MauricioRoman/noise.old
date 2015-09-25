package main

import (
	"net"
)

type Context struct {
	cfg   *Config
	chans map[*net.Conn]chan []byte
}

func NewContext(cfg *Config) *Context {
	ctx := new(Context)
	ctx.cfg = cfg
	ctx.chans = make(map[*net.Conn]chan []byte)
	return ctx
}
