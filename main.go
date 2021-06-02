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
		util.Printf("quiet: %v\n", cfg.Quiet)
		util.Printf("sort_keys: %v\n", cfg.SortKeys)
		util.Printf("path_tilde: %v\n", cfg.PathTilde)
	}

	baseEnv := util.NewProfile()
	baseEnv.MergeStrings(os.Environ())

	magenta := color.New(color.FgMagenta).SprintfFunc()
	yellow := color.New(color.FgYellow).SprintfFunc()
	// red := color.New(color.FgRed).SprintfFunc()
	//red := color.New(color.FgBlue).SprintfFunc()
	hiMagenta := color.New(color.FgHiMagenta).SprintfFunc()
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
		sortedEnv := baseEnv.SortedNames(false)
		matched := make(map[string]bool, len(sortedEnv))
		for _, group := range cfg.Groups.GetAllNames() {
			headerDisplayed := false
			patterns, found := cfg.Groups.GetPatterns(group)
			if !found {
				continue
			}
			for _, env := range sortedEnv {
				if util.MatchAny(env, patterns) {
					matched[env] = true
					if !showAllGroups && group[0] == '.' {
						continue // Skip display of hidden groups.
					}
					if showGroup != "" && showGroup != group {
						continue
					}
					if !headerDisplayed {
						util.Printf(magenta("# %s\n", group))
						headerDisplayed = true
					}
					value, _ := baseEnv.Get(env)
					util.Printf("  %s=%s\n", green(env), yellow(value))
				}
			}
		}
		headerDisplayed := false
		for _, env := range sortedEnv {
			_, found := matched[env]
			if !found {
				if !headerDisplayed {
					util.Printf(hiMagenta("# %s\n", "(no matching group)"))
					headerDisplayed = true
				}
				value, _ := baseEnv.Get(env)
				util.Printf("  %s=%s\n", green(env), yellow(value))
			}
		}
	}
}
