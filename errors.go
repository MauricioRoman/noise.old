package main

import "errors"

var (
	ErrInvalidInput     = errors.New("invalid input")
	ErrInvalidDBVal     = errors.New("invalid value in db")
	ErrInvalidCfgFactor = errors.New("invalid factor in config")
)
