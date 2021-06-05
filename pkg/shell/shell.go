package shell

import (
	"fmt"
	"strings"
)

// TODO: remove quotes if not needed.
// TODO: escape quotes and other shell special characters such as ;
// TODO: Shell and os specific code.

func ExportVar(name, value string) string {
	return fmt.Sprintf("export %s=\"%s\"", name, value)
}

func UnsetVar(name string) string {
	return fmt.Sprintf("unset %s", name)
}

func RunCommands(commands []string) string {
	if len(commands) > 0 {
		commands := append(commands, "")
		return strings.Join(commands, ";")
	} else {
		return ":"
	}
}
