package server

import "errors"

var (
	ErrDuplicate = errors.New("duplicate write")
)
