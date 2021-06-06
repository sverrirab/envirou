package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/sverrirab/envirou/pkg/config"
	"github.com/sverrirab/envirou/pkg/data"
	"github.com/sverrirab/envirou/pkg/output"
	"github.com/sverrirab/envirou/pkg/shell"
)

// Actions:
var actionShowGroup string
var actionListGroups bool
var actionListProfiles bool
var actionActiveProfilesColored bool

// Display modifiers
var showAllGroups bool
var displayRaw bool
var verbose bool
var displayShellCommands bool
var noColor bool

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
	addStrFlag(&actionShowGroup, []string{"g", "group"}, "", "Show a specific group only")
	addBoolFlag(&actionListProfiles, []string{"p", "profiles"}, false, "List profile names")
	addBoolFlag(&actionListGroups, []string{"l", "list"}, false, "List group names")
	addBoolFlag(&actionActiveProfilesColored, []string{"active-profiles-colored"}, false, "List active profiles only")

	addBoolFlag(&showAllGroups, []string{"a", "all"}, false, "Show all (including .hidden) groups")
	addBoolFlag(&displayRaw, []string{"w", "raw"}, false, "Display unformatted env variables")
	addBoolFlag(&verbose, []string{"v", "verbose"}, false, "Increase output verbosity")
	addBoolFlag(&displayShellCommands, []string{"display-shell"}, false, "Display shell commands")
	addBoolFlag(&noColor, []string{"no-color"}, false, "Disable colored output")
}

func displayGroup (cfg *config.Configuration, name string, envs data.Envs, baseEnv *data.Profile) {
	if len(envs) > 0 {
		if showAllGroups || (len(actionShowGroup) > 0 && name == actionShowGroup) || !strings.HasPrefix(name, ".") {
			output.PrintGroup(name)
			for _, env := range envs {
				value, _ := baseEnv.Get(env)
				output.PrintEnv(env, value, cfg.SettingsPath, cfg.SettingsPassword, displayRaw)
			}
		}
	}
}

func main() {
	flag.Parse()

	cfg, err := config.ReadConfiguration(config.GetDefaultConfigFilePath())
	if err != nil {
		output.Printf("Failed to read config file: %v\n", err)
		os.Exit(3)
	}

	output.NoColor(noColor)

	output.SetGroupColor(cfg.FormatGroup)
	output.SetProfileColor(cfg.FormatProfile)
	output.SetEnvNameColor(cfg.FormatEnvName)
	output.SetPathColor(cfg.FormatPath)

	baseEnv := data.NewProfile()
	baseEnv.MergeStrings(os.Environ())
	newEnv := baseEnv.Clone()

	// Figure out what profiles are active. 
	profileNames := make([]string, 0, len(cfg.Profiles))
	mergedNames := make([]string, 0, len(cfg.Profiles))
	for name, profile := range cfg.Profiles {
		profileNames = append(profileNames, name)
		if baseEnv.IsMerged(&profile) {
			mergedNames = append(mergedNames, name)
		}
	}
	sort.Strings(profileNames)
	sort.Strings(mergedNames)

	if actionListGroups {
		for _, group := range cfg.Groups.GetAllNames() {
			output.Printf(output.GroupSprintf("# %s\n", group))
		}
	} else if actionListProfiles {
		for _, profileName := range profileNames {
			output.Printf(output.ProfileSprintf("# %s\n", profileName))
		}
	} else if actionActiveProfilesColored {
		for _, profileName := range mergedNames {
			output.Printf(output.ProfileSprintf("%s ", profileName))
		}
	} else if flag.NArg() > 0 {
		for _, activateName := range flag.Args() {
			var foundProfile *data.Profile = nil
			for _, name := range profileNames {
				if activateName == name {
					profile := cfg.Profiles[name]
					foundProfile = &profile
					break
				}
			}
			if foundProfile == nil {
				output.Printf("Profile %s not found\n", output.DiffSprintf(activateName))
			} else {
				newEnv.Merge(foundProfile)
				output.Printf("Profile %s enabled\n", output.ProfileSprintf(activateName))
			}
		}
	} else {
		matches, remaining := cfg.Groups.MatchAll(baseEnv.SortedNames(false))
		for group, envs := range matches {
			displayGroup(cfg, group, envs, baseEnv)
		}
		displayGroup(cfg, "(no group)", remaining, baseEnv)

		output.PrintProfileList(profileNames, mergedNames)
	}
	commands := shell.GetCommands(baseEnv, newEnv)
	if len(commands) > 0 {
		if displayShellCommands {
			output.Printf("Shell commands to execute: %s\n", shell.RunCommands(commands))
		}
		fmt.Print(shell.RunCommands(commands))	
	}
}
