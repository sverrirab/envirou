package cmd

import (
	"github.com/spf13/cobra"
	"github.com/sverrirab/envirou/pkg/output"
)

var profilesCmd = &cobra.Command{
	Use:     "profiles",
	Aliases: []string{"profile", "p"},
	Short:   "List profiles",
	GroupID: "profiles",
	Run: func(cmd *cobra.Command, args []string) {
		for _, profileName := range profileNames {
			active := isActiveProfile[profileName]
			if active && !showInactiveProfilesOnly {
				output.Printf(out.ProfileSprintf("%s ", profileName))
			} else if !active && !showActiveProfilesOnly {
				output.Printf("%s ", profileName)
			}
		}
		output.Printf("\n")
	},
}

var (
	showActiveProfilesOnly   bool = false
	showInactiveProfilesOnly bool = false
)

func init() {
	addCommand(profilesCmd)

	profilesCmd.Flags().BoolVarP(&showActiveProfilesOnly, "active", "a", showActiveProfilesOnly, "Show active profiles only")
	profilesCmd.Flags().BoolVarP(&showInactiveProfilesOnly, "inactive", "i", showInactiveProfilesOnly, "Show inactive profiles only")
	profilesCmd.MarkFlagsMutuallyExclusive("active", "inactive")
}
