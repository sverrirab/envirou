package cmd

import (
	"os"
	"strings"

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
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		used := make(map[string]bool, len(args))
		for _, name := range args {
			used[name] = true
		}

		matches := make([]string, 0, len(app.profileNames))
		for _, name := range app.profileNames {
			if used[name] || !strings.HasPrefix(name, toComplete) {
				continue
			}
			status := "inactive profile"
			if app.isActiveProfile[name] {
				status = "active profile"
			}
			matches = append(matches, name+"\t"+status)
		}
		return matches, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		newEnv := app.baseEnv.Clone()
		var notFound []string
		var alreadyActive []string
		for _, activateName := range args {
			profile, found := findProfile(app.out, app.configuration, activateName)
			if !found {
				notFound = append(notFound, activateName)
				continue
			}
			if hasEncryptedValues(profile) {
				// Locked profile: get a key (ENVIROU_KEY or passphrase
				// prompt, at most once) and decrypt a copy before merging.
				key, err := ensureKey()
				if err != nil {
					output.Printf("%s\n", err.Error())
					os.Exit(1)
				}
				decrypted := profile.Clone()
				if err := decryptProfileInPlace(decrypted, key); err != nil {
					output.Printf("%s\n", err.Error())
					os.Exit(1)
				}
				profile = decrypted
			}
			wasActive := app.baseEnv.IsMerged(profile)
			result := newEnv.Merge(profile)
			if wasActive {
				alreadyActive = append(alreadyActive, activateName)
			} else {
				suffix := ""
				if len(result.PathSkipped) > 0 {
					suffix = " (* " + strings.Join(result.PathSkipped, ", ") + " already in path)"
				}
				output.Printf("Profile %s enabled%s\n", app.out.ProfileSprintf(activateName), suffix)
			}
		}
		if len(alreadyActive) > 0 {
			colored := make([]string, len(alreadyActive))
			for i, name := range alreadyActive {
				colored[i] = app.out.ProfileSprintf(name)
			}
			output.Printf("Already active: %s\n", strings.Join(colored, ", "))
		}
		if len(notFound) > 0 {
			output.Printf("Warning: profiles not found: %s\n", strings.Join(notFound, ", "))
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
