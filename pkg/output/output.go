package output

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/sverrirab/envirou/pkg/data"
)

const (
	pathListSeperator = string(os.PathListSeparator)
)

var GroupSprintf = color.New(color.FgMagenta).SprintfFunc()
var ProfileSprintf = color.New(color.FgGreen).SprintfFunc()
var EnvNameSprintf = color.New(color.FgHiCyan).SprintfFunc()
var PathSprintf = color.New(color.Underline).SprintfFunc()
var DiffSprintf = color.New(color.FgRed).SprintfFunc()

// Printf output shown to end user - all output goes to stderr
func Printf(format string, a ...interface{}) {
	_, err := fmt.Fprintf(os.Stderr, format, a...)
	if err != nil {
		panic("Failed to output string")
	}
}

func mapColor(value string) (color.Attribute, bool) {
	switch value {
	case "green":
		return color.FgGreen, true
	case "magenta":
		return color.FgMagenta, true
	case "red":
		return color.FgRed, true
	case "yellow":
		return color.FgYellow, true
	case "blue":
		return color.FgBlue, true
	case "cyan":
		return color.FgHiCyan, true
	case "white":
		return color.FgWhite, true
	case "black":
		return color.FgBlack, true
	case "bold":
		return color.Bold, true
	case "underline":
		return color.Underline, true
	case "deleted":
		return color.CrossedOut, true
	case "none":
		return color.CrossedOut, true
	default:
		return color.Reset, false
	}
}

func mapColorDefault(value string) color.Attribute {
	color, _ := mapColor(value)
	return color
}

func IsValidColor(value string) bool {
	_, found := mapColor(value)
	return found
}

func SetGroupColor(value string) {
	GroupSprintf = color.New(mapColorDefault(value)).SprintfFunc()
}

func SetProfileColor(value string) {
	ProfileSprintf = color.New(mapColorDefault(value)).SprintfFunc()
}

func SetEnvNameColor(value string) {
	EnvNameSprintf = color.New(mapColorDefault(value)).SprintfFunc()
}

func SetPathColor(value string) {
	PathSprintf = color.New(mapColorDefault(value)).SprintfFunc()
}

func SprintEnv(name, value string, paths, passwords data.Patterns, displayRaw bool) string {
	outputName := name
	outputValue := value
	if !displayRaw {
		outputName = EnvNameSprintf("%s", name)
		if data.MatchAny(name, &passwords) {
			outputValue = "****--->hidden<---****"
		} else if data.MatchAny(name, &paths) {
			sections := strings.Split(value, pathListSeperator)
			for i, section := range sections {
				if i % 2 == 1 {
					// Color every other path
					sections[i] = PathSprintf(section)
				}
			}
			outputValue = strings.Join(sections, pathListSeperator)
		}
	}
	return fmt.Sprintf("%s=%s\n", outputName, outputValue)
}

func PrintEnv(name, value string, paths, passwords data.Patterns, displayRaw bool) {
	Printf(SprintEnv(name, value, paths, passwords, displayRaw))
}

func PrintGroup(name string) {
	Printf(GroupSprintf("# %s\n", name))
}

func PrintProfileList(profileNames, mergedNames []string) {
	if len(profileNames) == 0 {
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
