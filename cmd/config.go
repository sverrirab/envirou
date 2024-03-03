package cmd

import (
	"fmt"
	"github.com/sverrirab/envirou/pkg/config"
	"github.com/sverrirab/envirou/pkg/output"
	"os"

	"github.com/spf13/cobra"
)

// setCmd represents the set command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure settings",
	Long:  `By default this will run an editor with the current config file`,
	Run: func(cmd *cobra.Command, args []string) {
		editor, found := baseEnv.Get("EDITOR")
		if !found {
			output.Printf("You need to set the EDITOR environment to point to your editor first\n")
			output.Printf("Configuration file location: %s\n", config.GetDefaultConfigFilePath())
			os.Exit(3)
		}
		output.Printf("Launching EDITOR ...\n")
		configFile := config.GetDefaultConfigFilePath()
		shellCommands = append(shellCommands, fmt.Sprintf("%s \"%s\"", editor, configFile))
	},
}

func init() {
	addCommand(configCmd)
}
