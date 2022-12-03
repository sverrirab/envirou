package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/sverrirab/envirou/pkg/data"
	"github.com/sverrirab/envirou/pkg/shell"
)

const (
	pathListSeperator = string(os.PathListSeparator)
)

type ColorPrintFunc func(format string, a ...interface{}) string

type Output struct {
	replacePathTilde string
	paths            data.Patterns
	passwords        data.Patterns
	displayRaw       bool

	groupSprintf   ColorPrintFunc
	profileSprintf ColorPrintFunc
	envNameSprintf ColorPrintFunc
	pathSprintf    ColorPrintFunc
	diffSprintf    ColorPrintFunc
}

func NewOutput(replacePathTilde string, paths, passwords data.Patterns, displayRaw bool,
	groupColor, profileColor, envNameColor, pathColor, diffColor string) *Output {
	return &Output{
		replacePathTilde: replacePathTilde,
		paths:            paths,
		passwords:        passwords,
		displayRaw:       displayRaw,
		groupSprintf:     color.New(mapColorDefault(groupColor, "magenta")).SprintfFunc(),
		profileSprintf:   color.New(mapColorDefault(profileColor, "green")).SprintfFunc(),
		envNameSprintf:   color.New(mapColorDefault(envNameColor, "cyan")).SprintfFunc(),
		pathSprintf:      color.New(mapColorDefault(pathColor, "underline")).SprintfFunc(),
		diffSprintf:      color.New(mapColorDefault(diffColor, "red")).SprintfFunc(),
	}
}

// Printf output shown to end user - all output goes to stderr
func Printf(format string, a ...interface{}) {
	_, err := fmt.Fprintf(os.Stderr, format, a...)
	if err != nil {
		panic("Failed to output string")
	}
}

// NoColor forces colored output
func NoColor(noColor bool) {
	color.NoColor = noColor
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
	case "reverse":
		return color.ReverseVideo, true
	case "deleted":
		return color.CrossedOut, true
	case "none":
		return color.Reset, true
	default:
		return color.Reset, false
	}
}

func mapColorDefault(value, defaultColor string) color.Attribute {
	color, found := mapColor(value)
	if !found {
		color, _ = mapColor(defaultColor)
	}
	return color
}

func IsValidColor(value string) bool {
	_, found := mapColor(value)
	return found
}

func (out *Output) GroupSprintf(format string, a ...interface{}) string {
	return out.groupSprintf(format, a...)
}

func (out *Output) ProfileSprintf(format string, a ...interface{}) string {
	return out.profileSprintf(format, a...)
}

func (out *Output) EnvNameSprintf(format string, a ...interface{}) string {
	return out.envNameSprintf(format, a...)
}

func (out *Output) PathSprintf(format string, a ...interface{}) string {
	return out.pathSprintf(format, a...)
}

func (out *Output) DiffSprintf(format string, a ...interface{}) string {
	return out.diffSprintf(format, a...)
}

func (out *Output) SprintEnv(name, value string) string {
	outputName := name
	outputValue := value
	if out.displayRaw {
		outputValue = shell.Escape(value)
	} else {
		outputName = out.EnvNameSprintf("%s", name)
		if data.MatchAny(name, &out.passwords) {
			outputValue = "****--->hidden<---****"
		} else if data.MatchAny(name, &out.paths) {
			sections := strings.Split(value, pathListSeperator)
			for i := range sections {
				if len(out.replacePathTilde) > 0 {
					if strings.HasPrefix(sections[i], out.replacePathTilde) {
						sections[i] = strings.Replace(sections[i], out.replacePathTilde, "~", 1)
					}
				}
				if i%2 == 1 {
					// Color every other path
					sections[i] = out.PathSprintf(sections[i])
				}
			}
			outputValue = strings.Join(sections, pathListSeperator)
		}
	}
	return fmt.Sprintf("%s=%s\n", outputName, outputValue)
}

func (out *Output) PrintEnv(name, value string) {
	Printf(out.SprintEnv(name, value))
}

func (out *Output) PrintGroup(name string) {
	Printf(out.GroupSprintf("# %s\n", name))
}

func (out *Output) PrintProfileList(profileNames, mergedNames []string) {
	Printf(out.SPrintProfileList(profileNames, mergedNames))
}

func (out *Output) SPrintProfileList(profileNames, mergedNames []string) string {
	if len(profileNames) == 0 {
		return ""
	}
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
			s = out.ProfileSprintf(name)
		}
		output = append(output, s)
	}
	return fmt.Sprintf("%s: %s\n", out.ProfileSprintf("# Profiles"), strings.Join(output, ", "))
}
