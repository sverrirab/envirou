package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"

	"github.com/sverrirab/envirou/pkg/config"
	"github.com/sverrirab/envirou/pkg/data"
	"github.com/sverrirab/envirou/pkg/output"
	"github.com/sverrirab/envirou/pkg/shell"
)

// Actions:
var actionShowGroup string
var actionDiffProfile string
var actionListGroups bool
var actionListProfiles bool
var actionListProfilesActive bool
var actionListProfilesActiveColored bool
var actionListProfilesInactive bool
var actionEditConfig bool

// Display modifiers
var showAllGroups bool
var displayUnformatted bool
var verbose bool
var noColor bool

// Bootstrap shell
var bootstrapPowerShell bool
var bootstrapBash bool

// Shell overrides
var outputPowerShell bool

// These variables contain embedded scripts
//
//go:embed powershell/ev.ps1
var embeddedBootstrapPowerShell string

//go:embed bash/ev.sh
var embeddedBootstrapBash string

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
	addStrFlag(&actionDiffProfile, []string{"d", "diff"}, "", "Show changes in current env from specfic profile")
	addBoolFlag(&actionListProfiles, []string{"p", "profiles"}, false, "List profile names")
	addBoolFlag(&actionListGroups, []string{"l", "list"}, false, "List group names")
	addBoolFlag(&actionListProfilesActive, []string{"active-profiles"}, false, "List active profiles only")
	addBoolFlag(&actionListProfilesActiveColored, []string{"active-profiles-colored"}, false, "List active profiles only (w/color)")
	addBoolFlag(&actionListProfilesInactive, []string{"inactive-profiles"}, false, "List inactive profiles only")
	addBoolFlag(&actionEditConfig, []string{"edit"}, false, "Edit configuration")

	addBoolFlag(&showAllGroups, []string{"a", "all"}, false, "Show all (including .hidden) groups")
	addBoolFlag(&displayUnformatted, []string{"u", "unformatted"}, false, "Display unformatted env variables")
	addBoolFlag(&verbose, []string{"v", "verbose"}, false, "Increase output verbosity")
	addBoolFlag(&noColor, []string{"no-color"}, false, "Disable colored output")

	addBoolFlag(&bootstrapPowerShell, []string{"bootstrap-powershell"}, false, "Enable PowerShell support (ev function)")
	addBoolFlag(&bootstrapBash, []string{"bootstrap-bash"}, false, "Enable Bash support (ev function)")

	addBoolFlag(&outputPowerShell, []string{"ps1", "output-powershell"}, false, "Enable PowerShell output")
}

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
	flag.Parse()

	cfg, err := config.ReadConfiguration(config.GetDefaultConfigFilePath())
	if err != nil {
		output.Printf("Failed to read config file: %v\n", err)
		os.Exit(3)
	}

	output.NoColor(noColor)
	replacePathTilde := ""
	runningOnWindows := runtime.GOOS != "windows"
	if runningOnWindows && cfg.SettingsPathTilde {
		replacePathTilde = os.Getenv("HOME")
	}
	sh := shell.NewShell(outputPowerShell, runningOnWindows && !outputPowerShell)
	out := output.NewOutput(replacePathTilde, cfg.SettingsPath, cfg.SettingsPassword, displayUnformatted, cfg.FormatGroup, cfg.FormatProfile, cfg.FormatEnvName, cfg.FormatPath, cfg.FormatDiff)

	baseEnv := data.NewProfile()
	baseEnv.MergeStrings(os.Environ())

	// Figure out what profiles are active.
	profileNames := make([]string, 0, len(cfg.Profiles))
	activeProfileNames := make([]string, 0, len(cfg.Profiles))
	inactiveProfileNames := make([]string, 0, len(cfg.Profiles))
	for name, profile := range cfg.Profiles {
		profileNames = append(profileNames, name)
		if baseEnv.IsMerged(&profile) {
			activeProfileNames = append(activeProfileNames, name)
		} else {
			inactiveProfileNames = append(inactiveProfileNames, name)
		}
	}
	sort.Strings(profileNames)
	sort.Strings(activeProfileNames)
	sort.Strings(inactiveProfileNames)
	shellCommands := make([]string, 0)

	switch {
	// For shell bootstrap we are outputting to stdout to simplify calling
	case bootstrapPowerShell:
		fmt.Print(embeddedBootstrapPowerShell)
	case bootstrapBash:
		fmt.Print(embeddedBootstrapBash)
	case actionListGroups:
		for _, group := range cfg.Groups.GetAllNames() {
			output.Printf(out.GroupSprintf("# %s\n", group))
		}
	case actionListProfiles:
		for _, profileName := range profileNames {
			output.Printf(out.ProfileSprintf("# %s\n", profileName))
		}
	case actionListProfilesActive:
		for _, profileName := range activeProfileNames {
			output.Printf("%s ", profileName)
		}
		output.Printf("\n")
	case actionListProfilesActiveColored:
		for _, profileName := range activeProfileNames {
			output.Printf(out.ProfileSprintf("%s ", profileName))
		}
	case actionListProfilesInactive:
		for _, profileName := range inactiveProfileNames {
			output.Printf("%s ", profileName)
		}
		output.Printf("\n")
	case actionEditConfig:
		editor, found := baseEnv.Get("EDITOR")
		if !found {
			output.Printf("You need to set the EDITOR environment to point to your editor first\n")
			output.Printf("Configuration file location: %s\n", config.GetDefaultConfigFilePath())
			os.Exit(3)
		}
		shellCommands = append(shellCommands, fmt.Sprintf("%s \"%s\"", editor, config.GetDefaultConfigFilePath()))
	case len(actionDiffProfile) > 0:
		profile, found := findProfile(out, cfg, actionDiffProfile)
		if found {
			//diffEnv := baseEnv.Clone()
			//diffEnv.Merge(profile)
			changed, removed := profile.Diff(baseEnv)
			displayGroup(out, "Changed", changed, baseEnv, sh)
			displayGroup(out, "Removed", removed, baseEnv, sh)
			//output.Printf("Profile %s enabled\n", out.ProfileSprintf(activateName))
		}
	case flag.NArg() > 0:
		newEnv := baseEnv.Clone()
		for _, activateName := range flag.Args() {
			profile, found := findProfile(out, cfg, activateName)
			if found {
				newEnv.Merge(profile)
				output.Printf("Profile %s enabled\n", out.ProfileSprintf(activateName))
			}
		}
		shellCommands = append(shellCommands, sh.GetCommands(baseEnv, newEnv)...)
	default:
		matches, remaining := cfg.Groups.MatchAll(baseEnv.SortedNames(false))
		if !showAllGroups && len(actionShowGroup) > 0 {
			if !displayGroup(out, actionShowGroup, matches[actionShowGroup], baseEnv, sh) {
				output.Printf(out.GroupSprintf("# %s (group empty, use -a to show all)\n", actionShowGroup))
			}
		} else {
			notDisplayed := make([]string, 0)
			for _, groupName := range matches.GetAllNames() {
				hideGroup := !showAllGroups && strings.HasPrefix(groupName, ".")
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
		if verbose {
			output.Printf("Shell commands to execute: %s\n", commands)
		}
		fmt.Print(commands)
	}
}
