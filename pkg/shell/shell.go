package shell

import (
	"fmt"
	"strings"

	"github.com/sverrirab/envirou/pkg/data"
)

type Shell struct {
	bat        bool
	powerShell bool
}

func NewShell(powerShell bool, bat bool) *Shell {
	return &Shell{
		powerShell: powerShell,
		bat:        bat,
	}
}

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

func (shell *Shell) Escape(value string) string {
	if shell.powerShell {
		return fmt.Sprintf("'%s'", strings.ReplaceAll(value, "'", "''"))
	} else if needsEscape(value) {
		return fmt.Sprintf("'%s'", strings.ReplaceAll(value, "'", "'\\''"))
	} else {
		return value
	}
}

func (shell *Shell) EscapePowerShell(value string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(value, "'", "''"))
}

func (shell *Shell) ExportVar(name, value string) string {
	if shell.powerShell {
		return fmt.Sprintf("$Env:%s = %s", name, shell.Escape(value))
	} else if shell.bat {
		return fmt.Sprintf("set %s=%s", name, shell.Escape(value))
	} else {
		return fmt.Sprintf("export %s=%s", name, shell.Escape(value))
	}
}

func (shell *Shell) UnsetVar(name string) string {
	if shell.powerShell {
		return fmt.Sprintf("Remove-Item Env:%s", name)
	} else if shell.bat {
		return fmt.Sprintf("set %s=", name)
	} else {
		return fmt.Sprintf("unset %s", name)
	}
}

func (shell *Shell) RunCommands(commands []string) string {
	if len(commands) == 0 {
		return ""
	} else {
		if shell.bat {
			// Windows bat file use & as seperator
			return fmt.Sprintf("%s\n", strings.Join(commands, " & "))
		} else {
			// Unixes require ; termination (as well as PowerShell)
			commands = append(commands, "")
			return fmt.Sprintf("%s\n", strings.Join(commands, ";"))
		}
	}
}

func (shell *Shell) GetCommands(old, new *data.Profile) (commands []string) {
	added, removed := old.Diff(new)
	for _, add := range added {
		value, _ := new.Get(add)
		commands = append(commands, shell.ExportVar(add, value))
	}
	for _, remove := range removed {
		commands = append(commands, shell.UnsetVar(remove))
	}
	return
}
