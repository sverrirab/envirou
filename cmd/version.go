package cmd

import (
	"github.com/spf13/cobra"
	"github.com/sverrirab/envirou/pkg/output"
	"runtime"
)

// Version is set at build time via -ldflags "-X github.com/sverrirab/envirou/cmd.Version=..."
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		output.Printf("Envirou (ev) Version %s on %s/%s\n", Version, runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	addCommand(versionCmd)
}
