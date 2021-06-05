package main

import (
	"flag"
	"os"
	"strings"

	"github.com/sverrirab/envirou/pkg/config"
	"github.com/sverrirab/envirou/pkg/data"
	"github.com/sverrirab/envirou/pkg/output"
)

var verbose bool
var displayRaw bool
var listGroups bool
var listProfiles bool
var showAllGroups bool
var showGroup string

func addBoolFlag(p *bool, names []string, value bool, usage string) {
	for _, name := range names {
		flag.BoolVar(p, name, value, usage)
	}
}

func addStrFlag(p *string, names []string, value string, usage string) {
	for _, name := range names {
		flag.StringVar(p, name, value, usage)
	}
}

func init() {
	addBoolFlag(&showAllGroups, []string{"a", "all"}, false, "Show all (including .hidden) groups")
	addBoolFlag(&listProfiles, []string{"p", "profiles"}, false, "List profile names")
	addBoolFlag(&listGroups, []string{"l", "list"}, false, "List group names")
	addBoolFlag(&displayRaw, []string{"w", "raw"}, false, "Display unformatted env variables")
	addBoolFlag(&verbose, []string{"v", "verbose"}, false, "Increase output verbosity")
	addStrFlag(&showGroup, []string{"g", "group"}, "", "Show a specific group only")
}

func main() {
	flag.Parse()

	cfg, err := config.ReadConfiguration(config.GetDefaultConfigFilePath())
	if err != nil {
		output.Printf("Failed to read config file: %v\n", err)
		os.Exit(3)
	}

	output.SetGroupColor(cfg.FormatGroup)
	output.SetProfileColor(cfg.FormatProfile)
	output.SetEnvNameColor(cfg.FormatEnvName)
	output.SetPathColor(cfg.FormatPath)

	baseEnv := data.NewProfile()
	baseEnv.MergeStrings(os.Environ())

	if listGroups {
		for _, group := range cfg.Groups.GetAllNames() {
			output.Printf(output.GroupSprintf("# %s\n", group))
		}
	} else if listProfiles {
		for profileName, profile := range cfg.Profiles {
			output.Printf(output.ProfileSprintf("# %s [%v]\n", profileName, profile))
		}
	} else if flag.NArg() > 0 {
		for _, f := range flag.Args() {
			for name, profile := range cfg.Profiles {
				if f == name {
					output.Printf("profile match: %s\n", f)
					added, removed := baseEnv.Diff(&profile)
					for _, add := range added {
						output.Printf("%s\n", add)
					}
					for _, remove := range removed {
						output.Printf("REMOVE: %s\n", remove)
					}
				}
			}
		}
	} else {
		displayGroup := func(name string, envs data.Envs, baseEnv *data.Profile) {
			if len(envs) > 0 {
				if showAllGroups || (len(showGroup) > 0 && name == showGroup) || !strings.HasPrefix(name, ".") {
					output.PrintGroup(name)
					for _, env := range envs {
						value, _ := baseEnv.Get(env)
						output.PrintEnv(env, value, cfg.SettingsPath, cfg.SettingsPassword, displayRaw)
					}
				}
			}
		}
		matches, remaining := cfg.Groups.MatchAll(baseEnv.SortedNames(false))
		for group, envs := range matches {
			displayGroup(group, envs, baseEnv)
		}
		displayGroup("(no group)", remaining, baseEnv)

		profileNames := make([]string, 0, len(cfg.Profiles))
		mergedNames := make([]string, 0, len(cfg.Profiles))
		for name, profile := range cfg.Profiles {
			profileNames = append(profileNames, name)
			if baseEnv.IsMerged(&profile) {
				mergedNames = append(mergedNames, name)
			}
		}
		output.PrintProfileList(profileNames, mergedNames)
	}
}
