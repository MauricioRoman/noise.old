package detector

import "errors"

var (
	ErrStatInputString  = errors.New("invalid stat input string")
	ErrStatOutputString = errors.New("invalid stat output string")
	ErrDBKeyNotFound    = errors.New("key is not found in db")
	ErrInvalidDBData    = errors.New("invalid data in db")
)
