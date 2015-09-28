package main

import "errors"

var (
	// ErrInvalidInput is returned when input string format
	// is invalid
	ErrInvalidInput = errors.New("invalid input")
)
