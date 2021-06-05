package main

import (
	"flag"
	"fmt"
	"os"
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

	addBoolFlag(&showAllGroups, []string{"a", "all"}, false, "Show all (including .hidden) groups")
	addBoolFlag(&displayRaw, []string{"w", "raw"}, false, "Display unformatted env variables")
	addBoolFlag(&verbose, []string{"v", "verbose"}, false, "Increase output verbosity")
	addBoolFlag(&displayShellCommands, []string{"display-shell"}, false, "Display shell commands")
	addBoolFlag(&noColor, []string{"no-color"}, false, "Disable colored output")
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

	shellCommands := make([]string, 0)

	if actionListGroups {
		for _, group := range cfg.Groups.GetAllNames() {
			output.Printf(output.GroupSprintf("# %s\n", group))
		}
	} else if actionListProfiles {
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
						value, _ := profile.Get(add)
						shellCommands = append(shellCommands, shell.ExportVar(add, value))
					}
					for _, remove := range removed {
						shellCommands = append(shellCommands, shell.UnsetVar(remove))
					}
				}
			}
		}
	} else {
		displayGroup := func(name string, envs data.Envs, baseEnv *data.Profile) {
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
	if len(shellCommands) > 0 {
		if displayShellCommands {
			output.Printf(output.DiffSprintf("Shell commands executed:\n"))
			for _, cmd := range shellCommands {
				output.Printf("%s\n", output.DiffSprintf(cmd))
			}
			output.Printf("Shortened: %s\n", shell.RunCommands(shellCommands))
		}
		fmt.Print(shell.RunCommands(shellCommands))
	}
}
