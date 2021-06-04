package main

import (
	"flag"
	"os"

	"github.com/fatih/color"
	"github.com/sverrirab/envirou/pkg/config"
	"github.com/sverrirab/envirou/pkg/util"
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
		util.Printf("verbose: %v\n", verbose)
		util.Printf("debug: %v\n", verbose)
		util.Printf("tail: %v\n", flag.Args())
	}

	cfg, err := config.ReadConfiguration(config.GetDefaultConfigFilePath())
	if err != nil {
		util.Printf("Failed to read config file: %v\n", err)
		os.Exit(3)
	}
	if debug {
		util.Printf("quiet: %t\n", cfg.Quiet)
		util.Printf("sort_keys: %t\n", cfg.SortKeys)
		util.Printf("path_tilde: %t\n", cfg.PathTilde)
		util.Printf("groups: %s\n", cfg.Groups)
		util.Printf("profiles: %s\n", cfg.Profiles)
	}

	baseEnv := util.NewProfile()
	baseEnv.MergeStrings(os.Environ())

	magenta := color.New(color.FgMagenta).SprintfFunc()
	yellow := color.New(color.FgYellow).SprintfFunc()
	// red := color.New(color.FgRed).SprintfFunc()
	// hiMagenta := color.New(color.FgHiMagenta).SprintfFunc()
	green := color.New(color.FgGreen).SprintfFunc()
	if listGroups {
		for _, group := range cfg.Groups.GetAllNames() {
			util.Printf(magenta("# %s\n", group))
		}
	} else if listProfiles {
		for profileName, profile := range cfg.Profiles {
			util.Printf(green("# %s [%v]\n", profileName, profile))
		}
	} else {
		displayGroup := func (name string, envs util.Envs, baseEnv *util.Profile) {
			if len(envs) > 0 {
				util.Printf(magenta("# %s\n", name))
				for _, env := range envs {
					value, _ := baseEnv.Get(env)
					util.Printf("  %s=%s\n", green(env), yellow(value))
				}
			}
		}
		sortedEnv := baseEnv.SortedNames(false)
		matches, remaining := cfg.Groups.MatchAll(sortedEnv)
		for group, envs := range matches {
			displayGroup(group, envs, baseEnv)
		}
		displayGroup("(no group)", remaining, baseEnv)
	}
}
