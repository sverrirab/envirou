package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sverrirab/envirou/pkg/shell"
)

// setCmd represents the set command
var bootstrapCmd = &cobra.Command{
	Use:       "bootstrap [bash|zsh|powershell|bat]",
	Short:     "Bootstrap current shell",
	Long:      `Run this in your shell initialization script`,
	GroupID:   "configuration",
	ValidArgs: []string{"bash", "zsh", "powershell", "bat"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("only provide one argument: type of shell to bootstrap")
		}
		if !contains(cmd.ValidArgs, args[0]) {
			validArgs := strings.Join(cmd.ValidArgs, ", ")
			return fmt.Errorf("invalid argument \"%s\", must be one of %s", args[0], validArgs)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if args[0] == "powershell" {
			app.sh = shell.NewShell(true, false)
			app.shellCommands = append(app.shellCommands, collapseToOneLine(powershellBootstrap))
			if addPrompt {
				app.shellCommands = append(app.shellCommands, collapseToOneLine(powershellPrompt))
			}
		} else if args[0] == "bat" {
			app.shellCommands = append(app.shellCommands, batBootstrap)
		} else { // bash + zsh
			// Removing the she-bang line from the script
			app.shellCommands = append(app.shellCommands, removeFirstLine(bashBootstrap))
		}
	},
}

var addPrompt bool

func init() {
	addCommand(bootstrapCmd)
	bootstrapCmd.Flags().BoolVarP(&addPrompt, "prompt", "p", addPrompt, "Also modify prompt (PowerShell only)")
}

func removeFirstLine(s string) string {
	lines := strings.SplitN(s, "\n", 2)
	if len(lines) > 1 {
		return lines[1]
	}
	return s
}

// collapseToOneLine converts a multi-line script to a single line
// by replacing newlines with "; " and collapsing extra whitespace.
func collapseToOneLine(s string) string {
	var parts []string
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return strings.Join(parts, "; ")
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
