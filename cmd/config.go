package cmd

import (
	"fmt"
	"github.com/sverrirab/envirou/pkg/config"
	"github.com/sverrirab/envirou/pkg/output"
	"os"

	"github.com/spf13/cobra"
)

// configCmd opens the config file in the user's editor.
var configCmd = &cobra.Command{
	Use:     "config",
	Short:   "Configure settings",
	Long:    `By default this will run an editor with the current config file`,
	GroupID: "configuration",
	Run: func(cmd *cobra.Command, args []string) {
		configFile := config.GetDefaultConfigFilePath()
		editor, found := app.baseEnv.Get("EDITOR")
		if !found {
			output.Printf("EDITOR is not set. To open the config file, either:\n")
			output.Printf("  export EDITOR=nano   # (or vim, code, etc.)\n")
			output.Printf("  ev config\n")
			output.Printf("Or edit directly:\n")
			output.Printf("  %s\n", configFile)
			os.Exit(1)
		}
		output.Printf("Launching EDITOR ...\n")
		app.shellCommands = append(app.shellCommands, fmt.Sprintf("%s \"%s\"", editor, configFile))
	},
}

func init() {
	addCommand(configCmd)
}
