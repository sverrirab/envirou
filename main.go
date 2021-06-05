package main

import (
	"flag"
	"os"

	"github.com/fatih/color"
	"github.com/sverrirab/envirou/pkg/config"
	"github.com/sverrirab/envirou/pkg/output"
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

	baseEnv := util.NewProfile()
	baseEnv.MergeStrings(os.Environ())

	magenta := color.New(color.FgMagenta).SprintfFunc()
	yellow := color.New(color.FgYellow).SprintfFunc()
	red := color.New(color.FgRed).SprintfFunc()
	// hiMagenta := color.New(color.FgHiMagenta).SprintfFunc()
	green := color.New(color.FgGreen).SprintfFunc()
	if listGroups {
		for _, group := range cfg.Groups.GetAllNames() {
			output.Printf(magenta("# %s\n", group))
		}
	} else if listProfiles {
		for profileName, profile := range cfg.Profiles {
			output.Printf(green("# %s [%v]\n", profileName, profile))
		}
	} else if flag.NArg() > 0 {
		for _, f := range flag.Args() {
			for name, profile := range cfg.Profiles {
				if f == name {
					output.Printf("profile match: %s\n", f)
					added, removed := baseEnv.Diff(&profile)
					for _, add := range added {
						output.Printf(green("%s\n", add))
					}
					for _, remove := range removed {
						output.Printf(red("%s\n", remove))
					}
				}
			}
		}
	} else {
		displayGroup := func(name string, envs util.Envs, baseEnv *util.Profile) {
			if len(envs) > 0 {
				output.Printf(magenta("# %s\n", name))
				for _, env := range envs {
					value, _ := baseEnv.Get(env)
					output.Printf("  %s=%s\n", green(env), yellow(value))
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
