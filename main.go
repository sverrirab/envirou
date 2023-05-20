package main

import (
	_ "embed"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"

	"github.com/sverrirab/envirou/pkg/cli"
	"github.com/sverrirab/envirou/pkg/config"
	"github.com/sverrirab/envirou/pkg/data"
	"github.com/sverrirab/envirou/pkg/output"
	"github.com/sverrirab/envirou/pkg/shell"
)

// These variables contain embedded scripts
//
//go:embed powershell/ev.ps1
var embeddedBootstrapPowerShell string

//go:embed bash/ev.sh
var embeddedBootstrapBash string

func displayGroup(out *output.Output, name string, envs data.Envs, profile *data.Profile, sh *shell.Shell) bool {
	if len(envs) > 0 {
		out.PrintGroup(name)
		for _, env := range envs {
			value, _ := profile.Get(env)
			out.PrintEnv(sh, env, value)
		}
		return true
	}
	return false
}

func findProfile(out *output.Output, cfg *config.Configuration, name string) (*data.Profile, bool) {
	profile, found := cfg.Profiles.FindProfile(name)
	if !found {
		output.Printf("Profile %s not found\n", out.DiffSprintf(name))
	}
	return profile, found
}

func main() {
	cfg, err := config.ReadConfiguration(config.GetDefaultConfigFilePath())
	if err != nil {
		output.Printf("Failed to read config file: %v\n", err)
		os.Exit(3)
	}

	flags := cli.ParseCommandLine(cfg)

	output.NoColor(flags.NoColor)

	replacePathTilde := ""
	runningOnWindows := runtime.GOOS != "windows"
	if runningOnWindows && cfg.SettingsPathTilde {
		replacePathTilde = os.Getenv("HOME")
	}
	sh := shell.NewShell(flags.OutputPowerShell, flags.OutputBat)
	out := output.NewOutput(replacePathTilde, cfg.SettingsPath, cfg.SettingsPassword, flags.DisplayUnformatted, cfg.FormatGroup, cfg.FormatProfile, cfg.FormatEnvName, cfg.FormatPath, cfg.FormatDiff)

	baseEnv := data.NewProfile()
	baseEnv.MergeStrings(os.Environ())

	// Figure out what profiles are active.
	profileNames := cfg.GetAllProfileNames()
	activeProfileNames := make([]string, 0, len(cfg.Profiles))
	inactiveProfileNames := make([]string, 0, len(cfg.Profiles))
	for name, profile := range cfg.Profiles {
		if baseEnv.IsMerged(&profile) {
			activeProfileNames = append(activeProfileNames, name)
		} else {
			inactiveProfileNames = append(inactiveProfileNames, name)
		}
	}
	sort.Strings(activeProfileNames)
	sort.Strings(inactiveProfileNames)
	shellCommands := make([]string, 0)

	switch {
	case flags.BootstrapPowerShell:
		shellCommands = append(shellCommands, embeddedBootstrapPowerShell)
	case flags.BootstrapBash:
		shellCommands = append(shellCommands, embeddedBootstrapBash)
	case flags.ActionListGroups:
		for _, group := range cfg.Groups.GetAllNames() {
			output.Printf(out.GroupSprintf("# %s\n", group))
		}
	case flags.ActionListProfiles:
		for _, profileName := range profileNames {
			output.Printf(out.ProfileSprintf("# %s\n", profileName))
		}
	case flags.ActionListProfilesActive:
		for _, profileName := range activeProfileNames {
			output.Printf("%s ", profileName)
		}
		output.Printf("\n")
	case flags.ActionListProfilesActiveColored:
		for _, profileName := range activeProfileNames {
			output.Printf(out.ProfileSprintf("%s ", profileName))
		}
	case flags.ActionListProfilesInactive:
		for _, profileName := range inactiveProfileNames {
			output.Printf("%s ", profileName)
		}
		output.Printf("\n")
	case flags.ActionEditConfig:
		editor, found := baseEnv.Get("EDITOR")
		if !found {
			output.Printf("You need to set the EDITOR environment to point to your editor first\n")
			output.Printf("Configuration file location: %s\n", config.GetDefaultConfigFilePath())
			os.Exit(3)
		}
		shellCommands = append(shellCommands, fmt.Sprintf("%s \"%s\"", editor, config.GetDefaultConfigFilePath()))
	case len(flags.ActionDiffProfile) > 0:
		profile, found := findProfile(out, cfg, flags.ActionDiffProfile)
		if found {
			//diffEnv := baseEnv.Clone()
			//diffEnv.Merge(profile)
			changed, removed := profile.Diff(baseEnv)
			displayGroup(out, "Changed", changed, baseEnv, sh)
			displayGroup(out, "Removed", removed, baseEnv, sh)
			//output.Printf("Profile %s enabled\n", out.ProfileSprintf(activateName))
		}
	case len(flags.ActionActivateProfile) > 0:
		newEnv := baseEnv.Clone()
		profile, found := findProfile(out, cfg, flags.ActionActivateProfile)
		if found {
			newEnv.Merge(profile)
			output.Printf("Profile %s enabled\n", out.ProfileSprintf(flags.ActionActivateProfile))
		}
		shellCommands = append(shellCommands, sh.GetCommands(baseEnv, newEnv)...)
	default:
		matches, remaining := cfg.Groups.MatchAll(baseEnv.SortedNames(false))
		if !flags.ShowAllGroups && len(flags.ActionShowGroup) > 0 {
			if !displayGroup(out, flags.ActionShowGroup, matches[flags.ActionShowGroup], baseEnv, sh) {
				output.Printf(out.GroupSprintf("# %s (group empty, use -a to show all)\n", flags.ActionShowGroup))
			}
		} else {
			notDisplayed := make([]string, 0)
			for _, groupName := range matches.GetAllNames() {
				hideGroup := !flags.ShowAllGroups && strings.HasPrefix(groupName, ".")
				if hideGroup || !displayGroup(out, groupName, matches[groupName], baseEnv, sh) {
					notDisplayed = append(notDisplayed, groupName)
				}
			}
			displayGroup(out, "(no group)", remaining, baseEnv, sh)
			if len(notDisplayed) > 0 && !cfg.SettingsQuiet {
				sort.Strings(notDisplayed)
				output.Printf(out.GroupSprintf("# Groups not displayed: %s (use -a to show all)\n", strings.Join(notDisplayed, " ")))
			}
		}

		out.PrintProfileList(profileNames, activeProfileNames)
	}
	if len(shellCommands) > 0 {
		commands := sh.RunCommands(shellCommands)
		if flags.Verbose {
			output.Printf("Shell commands to execute: %s\n", commands)
		}
		fmt.Print(commands)
	}
}
