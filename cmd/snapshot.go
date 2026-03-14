package cmd

import (
	"github.com/spf13/cobra"
	"github.com/sverrirab/envirou/pkg/config"
	"github.com/sverrirab/envirou/pkg/output"
)

var snapshotReset bool

var snapshotCmd = &cobra.Command{
	Use:     "snapshot",
	Aliases: []string{"snap"},
	Short:   "Save current environment as a snapshot for later diff",
	GroupID: "configuration",
	Run: func(cmd *cobra.Command, args []string) {
		if snapshotReset {
			err := config.RemoveSnapshot()
			if err != nil {
				output.Printf("Failed to remove snapshot: %v\n", err)
				return
			}
			output.Printf("Snapshot removed\n")
			return
		}
		err := config.SaveSnapshot(app.baseEnv, &app.configuration.Groups, app.caseInsensitive)
		if err != nil {
			output.Printf("Failed to save snapshot: %v\n", err)
			return
		}
		output.Printf("Snapshot saved\n")
	},
}

func init() {
	snapshotCmd.Flags().BoolVarP(&snapshotReset, "reset", "r", false, "Remove saved snapshot")
	addCommand(snapshotCmd)
}
