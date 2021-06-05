package output

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

var GroupSprintf = color.New(color.FgMagenta).SprintfFunc()
var ProfileSprintf = color.New(color.FgGreen).SprintfFunc()
var EnvNameSprintf = color.New(color.FgHiCyan).SprintfFunc()
var DiffSprintf = color.New(color.FgRed).SprintfFunc()

// Printf output shown to end user - all output goes to stderr
func Printf(format string, a ...interface{}) {
	_, err := fmt.Fprintf(os.Stderr, format, a...)
	if err != nil {
		panic("Failed to output string")
	}
}
