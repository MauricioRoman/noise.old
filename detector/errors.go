package detector

import "errors"

var (
	ErrInvalidStatInput = errors.New("invalid stat input")
	ErrDBKeyNotFound    = errors.New("key is not found in db")
	ErrInvalidDBData    = errors.New("invalid data in db")
)
