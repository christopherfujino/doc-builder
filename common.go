package main

import (
	"fmt"
	"os"
)

// TODO configure via CLI args
var debug = true

var ensureMode = false

func Trace(format string, args ...any) {
	if debug {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}
