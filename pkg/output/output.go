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

func mapColor(value string) color.Attribute {
	// ; <color> can be one of: green, magenta, red, yellow, cyan, blue, bold, underline
	// case "green", "magenta", "red", "yellow", "blue", "bold", "underline":
	switch value {
	case "green":
		return color.FgGreen
	case "magenta":
		return color.FgMagenta
	case "red":
		return color.FgRed
	case "yellow":
		return color.FgYellow
	case "blue":
		return color.FgBlue
	case "cyan":
		return color.FgHiCyan
	case "bold":
		return color.Bold
	case "underline":
		return color.FgRed
	case "deleted":
		return color.CrossedOut
	case "none":
		return color.Reset
	default:
		return color.FgHiRed
	}
}

func IsValidColor(value string) bool {
	switch value {
	case "green", "magenta", "red", "yellow", "cyan", "blue", "bold", "underline", "deleted", "none":
		return true
	default:
		return false
	}
}

func SetGroupColor(value string) {
	GroupSprintf = color.New(mapColor(value)).SprintfFunc()
}

func SetProfileColor(value string) {
	ProfileSprintf = color.New(mapColor(value)).SprintfFunc()
}

func SetEnvNameColor(value string) {
	EnvNameSprintf = color.New(mapColor(value)).SprintfFunc()
}

func PrintEnv(name, value string) {
	Printf("%s=%s\n", EnvNameSprintf("%s", name), value)
}

func PrintGroup(name string) {
	Printf(GroupSprintf("# %s\n", name))
}

func PrintProfileList(profileNames, mergedNames []string) {
	if len(profileNames)  ==  0 {
		return
	}
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
