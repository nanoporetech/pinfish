package main

import (
	"log"
	"os"
)

var L *log.Logger

// Create new logger.
func NewLogger(prefix string, flag int) *log.Logger {
	return log.New(os.Stderr, prefix, flag)
}
