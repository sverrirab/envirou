package cmd

import (
	"github.com/spf13/cobra"
	"strings"
)

// setCmd represents the set command
var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Bootstrap current shell",
	Long:  `Run this in your shell initialization script`,
	Run: func(cmd *cobra.Command, args []string) {
		if useBash {
			// Removing the she-bang line from the script
			shellCommands = append(shellCommands, removeFirstLine(bashBootstrap))
		} else if usePowershell {
			shellCommands = append(shellCommands, powershellBootstrap)
		} else {
			shellCommands = append(shellCommands, batBootstrap)
		}
	},
}

var (
	useBash       bool = false
	usePowershell bool = false
	useWindowsBat bool = false
)

func init() {
	addCommand(bootstrapCmd)

	bootstrapCmd.Flags().BoolVar(&useBash, "bash", useBash, "Use bash script")
	bootstrapCmd.Flags().BoolVar(&usePowershell, "powershell", usePowershell, "Use Powershell script")
	bootstrapCmd.Flags().BoolVar(&useWindowsBat, "bat", useWindowsBat, "Use Windows .bat script")
	bootstrapCmd.MarkFlagsOneRequired("bash", "powershell", "bat")
}

func removeFirstLine(s string) string {
	lines := strings.SplitN(s, "\n", 2)
	if len(lines) > 1 {
		return lines[1]
	}
	return s
}
