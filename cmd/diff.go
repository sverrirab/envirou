package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.com/sverrirab/envirou/pkg/config"
	"github.com/sverrirab/envirou/pkg/data"
	"github.com/sverrirab/envirou/pkg/output"
)

var diffSaveProfile string

var diffCmd = &cobra.Command{
	Use:     "diff",
	Short:   "Show environment changes since last snapshot",
	GroupID: "configuration",
	Run: func(cmd *cobra.Command, args []string) {
		snapshot, err := config.LoadSnapshot(app.caseInsensitive)
		if err != nil {
			output.Printf("Failed to load snapshot: %v\n", err)
			return
		}
		if snapshot == nil {
			output.Printf("No snapshot found. Run %s first.\n", app.out.ProfileSprintf("snapshot"))
			return
		}

		added, changed, removed := data.FullDiff(app.baseEnv, snapshot)

		// Filter out ignored vars
		added = filterIgnored(added, &app.configuration.Groups, app.caseInsensitive)
		changed = filterIgnored(changed, &app.configuration.Groups, app.caseInsensitive)
		removed = filterIgnored(removed, &app.configuration.Groups, app.caseInsensitive)

		if len(added) == 0 && len(changed) == 0 && len(removed) == 0 {
			output.Printf("No changes since snapshot\n")
			return
		}

		for _, name := range added {
			value, _ := app.baseEnv.Get(name)
			output.Printf("%s %s=%s\n", app.out.DiffSprintf("+"), app.out.EnvNameSprintf("%s", name), value)
		}
		for _, name := range changed {
			value, _ := app.baseEnv.Get(name)
			output.Printf("%s %s=%s\n", app.out.DiffSprintf("~"), app.out.EnvNameSprintf("%s", name), value)
		}
		for _, name := range removed {
			output.Printf("%s %s\n", app.out.DiffSprintf("-"), app.out.EnvNameSprintf("%s", name))
		}

		if diffSaveProfile != "" {
			_, exists := app.configuration.Profiles.FindProfile(diffSaveProfile)
			if exists {
				output.Printf("Profile %s already exists\n", app.out.DiffSprintf(diffSaveProfile))
				return
			}

			f, err := os.OpenFile(cfgFile, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				output.Printf("Failed to open config file: %v\n", err)
				return
			}
			defer f.Close()

			fmt.Fprintf(f, "\n[profile:%s]\n", diffSaveProfile)
			// Combine and sort all entries
			type entry struct {
				name  string
				value string
				isNil bool
			}
			entries := make([]entry, 0, len(added)+len(changed)+len(removed))
			for _, name := range added {
				value, _ := app.baseEnv.Get(name)
				entries = append(entries, entry{name, value, false})
			}
			for _, name := range changed {
				value, _ := app.baseEnv.Get(name)
				entries = append(entries, entry{name, value, false})
			}
			for _, name := range removed {
				entries = append(entries, entry{name, "", true})
			}
			sort.Slice(entries, func(i, j int) bool { return entries[i].name < entries[j].name })
			for _, e := range entries {
				if e.isNil {
					fmt.Fprintf(f, "%s\n", e.name)
				} else {
					fmt.Fprintf(f, "%s=%s\n", e.name, e.value)
				}
			}
			output.Printf("Saved profile %s\n", app.out.ProfileSprintf(diffSaveProfile))
		}
	},
}

func filterIgnored(names []string, groups *data.Groups, caseInsensitive bool) []string {
	result := make([]string, 0, len(names))
	for _, name := range names {
		if !groups.IsIgnored(name, caseInsensitive) {
			result = append(result, name)
		}
	}
	return result
}

func init() {
	diffCmd.Flags().StringVarP(&diffSaveProfile, "save", "s", "", "Save diff as a new profile")
	addCommand(diffCmd)
}
