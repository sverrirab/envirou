package cmd

import (
	"github.com/spf13/cobra"
	"github.com/sverrirab/envirou/pkg/output"
)

// setCmd represents the set command
var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Bootstrap current shell",
	Long:  `Run this in your shell initialization script`,
	Run: func(cmd *cobra.Command, args []string) {
		if useBash {
			output.Printf(bashBootstrap)
		} else if usePowershell {
			output.Printf(powershellBootstrap)
		} else {
			output.Printf(batBootstrap)
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
