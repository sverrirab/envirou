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

// IsValidVarName checks whether name is a safe environment variable identifier.
// Only allows POSIX-portable names: [a-zA-Z_][a-zA-Z0-9_]*
func IsValidVarName(name string) bool {
	if len(name) == 0 {
		return false
	}
	c := name[0]
	if !isLetter(c) && c != '_' {
		return false
	}
	for i := 1; i < len(name); i++ {
		c = name[i]
		if !isLetter(c) && !isDigit(c) && c != '_' {
			return false
		}
	}
	return true
}

func isLetter(c byte) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

func isDigit(c byte) bool {
	return '0' <= c && c <= '9'
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

func (shell *Shell) ExportVar(name, value string) string {
	if !IsValidVarName(name) {
		return ""
	}
	if shell.powerShell {
		return fmt.Sprintf("$Env:%s = %s", name, shell.Escape(value))
	} else if shell.bat {
		return fmt.Sprintf("set %s=%s", name, shell.Escape(value))
	} else {
		return fmt.Sprintf("export %s=%s", name, shell.Escape(value))
	}
}

// LocalVar sets a variable visible only to the shell itself, not inherited
// by child processes. POSIX shells get a plain assignment (the leading unset
// strips any export attribute left by older versions); PowerShell gets a
// global PS variable (and any process-level copy is removed). cmd has no
// non-exported variables so bat falls back to a regular set.
func (shell *Shell) LocalVar(name, value string) string {
	if !IsValidVarName(name) {
		return ""
	}
	if shell.powerShell {
		return fmt.Sprintf("Remove-Item Env:%s -ErrorAction SilentlyContinue;$global:%s = %s", name, name, shell.Escape(value))
	} else if shell.bat {
		return fmt.Sprintf("set %s=%s", name, shell.Escape(value))
	} else {
		return fmt.Sprintf("unset %s;%s=%s", name, name, shell.Escape(value))
	}
}

// UnsetLocalVar clears both the shell-local variable and any exported copy.
func (shell *Shell) UnsetLocalVar(name string) string {
	if !IsValidVarName(name) {
		return ""
	}
	if shell.powerShell {
		return fmt.Sprintf("Remove-Item Env:%s -ErrorAction SilentlyContinue;Remove-Variable -Name %s -Scope Global -ErrorAction SilentlyContinue", name, name)
	}
	// POSIX unset clears shell and environment variables alike; bat set= too.
	return shell.UnsetVar(name)
}

func (shell *Shell) UnsetVar(name string) string {
	if !IsValidVarName(name) {
		return ""
	}
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
			// Windows bat file use & as separator
			return fmt.Sprintf("%s\n", strings.Join(commands, " & "))
		} else {
			// Unixes require ; termination (as well as PowerShell)
			// commands = append(commands, "")  // Bash does not like this
			return fmt.Sprintf("%s\n", strings.Join(commands, ";"))
		}
	}
}

// RedactCommandValues returns shell commands suitable for diagnostic output.
// Assignment values may contain secrets, so only variable names and unset
// operations are preserved.
func (shell *Shell) RedactCommandValues(commands []string) []string {
	redacted := make([]string, len(commands))
	for i, command := range commands {
		redacted[i] = command
		if shell.powerShell {
			if idx := strings.Index(command, " = "); idx >= 0 {
				redacted[i] = command[:idx+3] + "<redacted>"
			}
			continue
		}
		if shell.bat {
			if strings.HasPrefix(command, "set ") {
				if idx := strings.Index(command, "="); idx >= 0 && idx < len(command)-1 {
					redacted[i] = command[:idx+1] + "<redacted>"
				}
			}
			continue
		}
		if strings.HasPrefix(command, "export ") {
			if idx := strings.Index(command, "="); idx >= 0 {
				redacted[i] = command[:idx+1] + "<redacted>"
			}
			continue
		}
		redacted[i] = redactPosixAssignment(command)
	}
	return redacted
}

// redactPosixAssignment redacts the value of a plain NAME=value assignment,
// optionally preceded by "unset NAME;" (the LocalVar form). Commands that
// are not assignments are returned unchanged.
func redactPosixAssignment(command string) string {
	prefixLen := 0
	rest := command
	if strings.HasPrefix(rest, "unset ") {
		idx := strings.Index(rest, ";")
		if idx < 0 || !IsValidVarName(rest[len("unset "):idx]) {
			return command
		}
		prefixLen = idx + 1
		rest = rest[prefixLen:]
	}
	idx := strings.Index(rest, "=")
	if idx <= 0 || !IsValidVarName(rest[:idx]) {
		return command
	}
	return command[:prefixLen+idx+1] + "<redacted>"
}

func (shell *Shell) GetCommands(old, new *data.Profile) (commands []string) {
	added, removed := old.Diff(new)
	for _, add := range added {
		value, _ := new.Get(add)
		if cmd := shell.ExportVar(add, value); cmd != "" {
			commands = append(commands, cmd)
		}
	}
	for _, remove := range removed {
		if cmd := shell.UnsetVar(remove); cmd != "" {
			commands = append(commands, cmd)
		}
	}
	return
}
