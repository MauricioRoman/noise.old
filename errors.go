package main

import "errors"

var (
	ErrInvalidInput     = errors.New("invalid input from pub end")
	ErrInvalidDBVal     = errors.New("invalid value is found in db")
	ErrInvalidCfgFactor = errors.New("invalid factor in config (require 0~1)")
)
