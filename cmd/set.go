package cmd

import (
	"github.com/sverrirab/envirou/pkg/config"
	"github.com/sverrirab/envirou/pkg/data"
	"github.com/sverrirab/envirou/pkg/output"

	"github.com/spf13/cobra"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:     "set PROFILE1 [PROFILE2] ...",
	Aliases: []string{"."},
	Short:   "Update current environment using profiles",
	Long: `Each profile will be merged with your current environment

To change profiles edit the config file (see "config" command)`,
	// ValidArgs: []string{"one", "two", "three"},
	GroupID: "profiles",
	Args:    cobra.MatchAll(cobra.MinimumNArgs(1)),
	Run: func(cmd *cobra.Command, args []string) {
		newEnv := app.baseEnv.Clone()
		for _, activateName := range args {
			profile, found := findProfile(app.out, app.configuration, activateName)
			if found {
				newEnv.Merge(profile)
				output.Printf("Profile %s enabled\n", app.out.ProfileSprintf(activateName))
			}
		}
		app.shellCommands = append(app.shellCommands, app.sh.GetCommands(app.baseEnv, newEnv)...)
	},
}

func findProfile(out *output.Output, cfg *config.Configuration, name string) (*data.Profile, bool) {
	profile, found := cfg.Profiles.FindProfile(name)
	if !found {
		output.Printf("Profile %s not found\n", out.DiffSprintf(name))
	}
	return profile, found
}

func init() {
	addCommand(setCmd)
}
