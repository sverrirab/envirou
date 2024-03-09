package cmd

import (
	"github.com/sverrirab/envirou/pkg/output"

	"github.com/spf13/cobra"
)

var groupsCmd = &cobra.Command{
	Use:     "groups",
	Aliases: []string{"group", "g"},
	Short:   "List all groups",
	Long:    `List all the groups defined in the config file`,
	GroupID: "groups",
	Run: func(cmd *cobra.Command, args []string) {
		for _, group := range configuration.Groups.GetAllNames() {
			output.Printf(out.GroupSprintf("# %s\n", group))
		}
	},
}

func init() {
	addCommand(groupsCmd)
}
