package main

import (
	"flag"
	"os"

	"github.com/fatih/color"
	"github.com/sverrirab/envirou/pkg/config"
	"github.com/sverrirab/envirou/pkg/state"
	"github.com/sverrirab/envirou/pkg/util"
)

var verbose bool
var debug bool
var group string
var listGroups bool

func init() {
	const (
		listGroupsDefault     = false
		listGroupsDescription = "List group names"
		verboseDefault        = false
		verboseDescription    = "Increase output verbosity"
		debugDefault          = false
		debugDescription      = "Output debug information"
	)
	flag.BoolVar(&listGroups, "l", listGroupsDefault, listGroupsDescription)
	flag.BoolVar(&listGroups, "list", listGroupsDefault, listGroupsDescription)
	flag.BoolVar(&verbose, "v", verboseDefault, verboseDescription)
	flag.BoolVar(&verbose, "verbose", verboseDefault, verboseDescription)
	flag.BoolVar(&debug, "debug", debugDefault, debugDescription)

	flag.StringVar(&group, "group", "", "groups!!")
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

	env := state.NewEnvirou(os.Environ())

	//displayGroups := 
	magenta := color.New(color.FgHiMagenta).SprintfFunc()
	yellow := color.New(color.FgYellow).SprintfFunc()
	red := color.New(color.FgRed).SprintfFunc()
	if listGroups {
		for _, group := range cfg.GroupsSorted {
			util.Printf(magenta("# %s\n", group))
		}
	} else {
		for _, key := range env.SortedKeys {
			if len(flag.Args()) == 1 {
				pattern := flag.Args()[0]
				if util.Match(key, util.Pattern(pattern)) {
					util.Printf("Matched %s vs %s\n", pattern, key)
				} else {
					util.Printf("No match %s vs %s\n", pattern, key)
				}
			} else {
				util.Printf("%s -> %s\t\n", yellow(key), red(env.Env[key]))
				matchAny := false
				for group, patterns := range cfg.Groups {
					if util.MatchAny(key, &patterns) {
						util.Printf("%s,\t", yellow(group))
						matchAny = true
					}
				}
				if !matchAny {
					util.Printf(" - NO MATCH - ")
				}
				util.Printf("\n")
			}
		}
	}
}
