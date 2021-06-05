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
var debug bool
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
	addBoolFlag(&verbose, []string{"v", "verbose"}, false, "Increase output verbosity")
	addBoolFlag(&debug, []string{"debug"}, false, "Output debug information")
	addStrFlag(&showGroup, []string{"g", "group"}, "", "Show a specific group only")
	// --dry-run for shell code?
}

func main() {
	flag.Parse()

	if debug {
		output.Printf("verbose: %v\n", verbose)
		output.Printf("debug: %v\n", verbose)
		output.Printf("tail: %v\n", flag.Args())
	}

	cfg, err := config.ReadConfiguration(config.GetDefaultConfigFilePath())
	if err != nil {
		output.Printf("Failed to read config file: %v\n", err)
		os.Exit(3)
	}
	if debug {
		output.Printf("quiet: %t\n", cfg.Quiet)
		output.Printf("sort_keys: %t\n", cfg.SortKeys)
		output.Printf("path_tilde: %t\n", cfg.PathTilde)
		output.Printf("groups: %s\n", cfg.Groups)
		output.Printf("profiles: %s\n", cfg.Profiles)
	}

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
					output.Group(name)
					for _, env := range envs {
						value, _ := baseEnv.Get(env)
						output.Env(env, value)
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
		output.ProfileList(profileNames, mergedNames)
	}
}
