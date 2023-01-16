package shell

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/sverrirab/envirou/pkg/data"
)

// TODO: Shell and os specific code.

func needsEscape(value string) bool {
	for i := 0; i < len(value); i++ {
		c := value[i]
		switch {
		case 'a' <= c && c <= 'z':
			continue
		case 'A' <= c && c <= 'Z':
			continue
		case '0' <= c && c <= '9':
			continue
		case c == '/':
			continue
		case c == ':':
			continue
		case c == ';':
			continue
		case c == '_':
			continue
		case c == '-':
			continue
		case c == '+':
			continue
		case c == '.':
			continue
		case c == '?':
			continue
		case c == ',':
			continue
		case c == '!':
			continue
		case c == '#':
			continue
		case c == '=':
			continue
		case c == '*':
			continue
		default:
			return true
		}
	}
	return false
}

func Escape(value string) string {
	if needsEscape(value) {
		return fmt.Sprintf("'%s'", strings.ReplaceAll(value, "'", "'\\''"))
	} else {
		return value
	}
}

func ExportVar(name, value string) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("set %s=%s", name, Escape(value))
	} else {
		return fmt.Sprintf("export %s=%s", name, Escape(value))
	}
}

func UnsetVar(name string) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("set %s=", name)
	} else {
		return fmt.Sprintf("unset %s", name)
	}
}

func RunCommands(commands []string) string {
	if len(commands) == 0 {
		return ""
	} else {
		if runtime.GOOS == "windows" {
			// Windows uses & as seperator
			return fmt.Sprintf("%s\n", strings.Join(commands, " & "))
		} else {
			// Unixes require ; termination
			commands = append(commands, "")
			return fmt.Sprintf("%s\n", strings.Join(commands, ";"))
		}
	}
}

func GetCommands(old, new *data.Profile) (commands []string) {
	added, removed := old.Diff(new)
	for _, add := range added {
		value, _ := new.Get(add)
		commands = append(commands, ExportVar(add, value))
	}
	for _, remove := range removed {
		commands = append(commands, UnsetVar(remove))
	}
	return
}
