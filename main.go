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
var group string
var listGroups bool
var listProfiles bool

func init() {
	const (
		listGroupsDefault       = false
		listGroupsDescription   = "List group names"
		listProfilesDefault     = false
		listProfilesDescription = "List profile names"
		verboseDefault          = false
		verboseDescription      = "Increase output verbosity"
		debugDefault            = false
		debugDescription        = "Output debug information"
	)
	flag.BoolVar(&listGroups, "l", listGroupsDefault, listGroupsDescription)
	flag.BoolVar(&listGroups, "list", listGroupsDefault, listGroupsDescription)
	flag.BoolVar(&listProfiles, "profiles", listProfilesDefault, listProfilesDescription)
	flag.BoolVar(&verbose, "v", verboseDefault, verboseDescription)
	flag.BoolVar(&verbose, "verbose", verboseDefault, verboseDescription)
	flag.BoolVar(&debug, "debug", debugDefault, debugDescription)

	flag.StringVar(&group, "groups", "", "groups!!")
	// flag.StringVar(&listProfiles, "profiles", "", "groups!!")
	// --dry-run for shell code?
}

func main() {

	flag.Parse()

	if debug {
		util.Printf("verbose: %v\n", verbose)
		util.Printf("debug: %v\n", verbose)
		util.Printf("group: %v\n", group)
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
		for _, group := range cfg.GroupsSorted {
			util.Printf(magenta("# %s\n", group))
		}
	} else if listProfiles {
		for profileName, profile := range cfg.Profiles {
			util.Printf(green("# %s [%v]\n", profileName, profile))
		}
	} else {
		sortedEnv := baseEnv.SortedNames(false)
		matched := make(map[string]string, len(sortedEnv))
		for _, group := range cfg.GroupsSorted {
			if group[0] != '.' {
				headerDisplayed := false
				patterns := cfg.Groups[group]
				for _, env := range sortedEnv {
					if util.MatchAny(env, &patterns) {
						if !headerDisplayed {
							util.Printf(magenta("# %s\n", group))
							headerDisplayed = true
						}
						value, _ := baseEnv.Get(env)
						util.Printf("  %s=%s\n", green(env), yellow(value))
						matched[env] = value
					}
				}
			}
		}
		headerDisplayed := false
		for _, env := range sortedEnv {
			_, found := matched[env]
			if ! found {
				if !headerDisplayed {
					util.Printf(hiMagenta("# %s\n", "not matched"))
					headerDisplayed = true
				}
				value, _ := baseEnv.Get(env)
				util.Printf("  %s=%s\n", green(env), yellow(value))
			}
		}
	}
}
