package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"
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

		for _, arg := range args {
			if !contains(cmd.ValidArgs, arg) {
				validArgs := strings.Join(cmd.ValidArgs, ",")
				return fmt.Errorf("invalid argument \"%s\", must be one of %s", arg, validArgs)
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if args[0] == "powershell" {
			shellCommands = append(shellCommands, powershellBootstrap)
		} else if args[0] == "bat" {
			shellCommands = append(shellCommands, batBootstrap)
		} else { // bash + zsh
			// Removing the she-bang line from the script
			shellCommands = append(shellCommands, removeFirstLine(bashBootstrap))
		}
	},
}

func init() {
	addCommand(bootstrapCmd)
}

func removeFirstLine(s string) string {
	lines := strings.SplitN(s, "\n", 2)
	if len(lines) > 1 {
		return lines[1]
	}
	return s
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
