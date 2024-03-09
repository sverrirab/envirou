package cmd

import (
	"github.com/spf13/cobra"
	"github.com/sverrirab/envirou/pkg/output"
	"runtime"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		output.Printf("Envirou (ev) Version 5.0 on %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	addCommand(versionCmd)
}
