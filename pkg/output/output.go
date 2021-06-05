package output

import (
	"fmt"
	"os"
	"sort"
	"strings"

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

func Env(name, value string) {
	Printf("%s=%s\n", EnvNameSprintf("%s", name), value)
}

func Group(name string) {
	Printf(GroupSprintf("# %s\n", name))
}

func ProfileList(profileNames, mergedNames []string) {
	sort.Strings(profileNames)
	output := make([]string, 0, len(profileNames))
	for _, name := range profileNames {
		isMerged := false
		for _, mergeName := range mergedNames {
			if mergeName == name {
				isMerged = true
			}
		}
		s := name
		if isMerged {
			s = ProfileSprintf(name)
		}
		output = append(output, s)
	}
	Printf("%s: %s\n", ProfileSprintf("# Profiles"), strings.Join(output, ", "))
}
