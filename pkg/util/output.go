package util

import (
	"fmt"
	"os"
)

// Printf Print output shown to end user - all output goes to stderr
func Printf(format string, a ...interface{}) {
	_, err := fmt.Fprintf(os.Stderr, format, a...)
	if err != nil {
		panic("Failed to output string")
	}
}

// Printlnf Print output shown to end user - all output goes to stderr
func Printlnf(format string, a ...interface{}) {
	Printf(format+"\n", a...)
}