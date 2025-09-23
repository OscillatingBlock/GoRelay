package http_errors

import "errors"

var (
	ErrNoHealthyBackend = errors.New("no healthy backends")
	ErrInvalidConfig    = errors.New("invalid config")
)
