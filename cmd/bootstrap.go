package cmd

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sverrirab/envirou/pkg/shell"
)

// bootstrapCmd outputs the shell integration script for a given shell.
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
	RunE: func(cmd *cobra.Command, args []string) error {
		if bootstrapCompletion && args[0] == "bat" {
			return fmt.Errorf("completion is not supported for the bat bootstrap")
		}
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
		if bootstrapCompletion {
			var completion bytes.Buffer
			if err := writeCompletion(args[0], &completion); err != nil {
				return err
			}
			// Keep the wrapper and generated script in one output item. Inserting
			// them as separate shell commands would put RunCommands' semicolon
			// between a function declaration and the script's leading comment.
			last := len(app.shellCommands) - 1
			app.shellCommands[last] += "\n" + completion.String()
		}
		return nil
	},
}

var addPrompt bool
var bootstrapCompletion bool

func init() {
	addCommand(bootstrapCmd)
	bootstrapCmd.Flags().BoolVarP(&addPrompt, "prompt", "p", addPrompt, "Also modify prompt (PowerShell only)")
	bootstrapCmd.Flags().BoolVarP(&bootstrapCompletion, "completion", "c", bootstrapCompletion, "Also load completion for envirou and ev")
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
