package main

import (
	"flag"
	"os"

	"github.com/fatih/color"
	"github.com/sverrirab/envirou/config"
	"github.com/sverrirab/envirou/util"
)

var verbose bool
var debug bool

func init() {
	const (
		verboseDefault     = false
		verboseDescription = "Increase output verbosity"
		debugDefault       = false
		debugDescription   = "Output debug information"
	)
	flag.BoolVar(&verbose, "verbose", verboseDefault, verboseDescription)
	flag.BoolVar(&verbose, "v", verboseDefault, verboseDescription)
	flag.BoolVar(&debug, "debug", debugDefault, debugDescription)
}

func main() {

	//var svar string
	//flag.StringVar(&svar, "svar", "bar", "a string var")

	flag.Parse()

	if debug {
		util.Printlnf("verbose: %v", verbose)
		util.Printlnf("debug: %v", verbose)
		util.Printlnf("tail: %v", flag.Args())
	}

	env, keys := util.ParseEnvironment(os.Environ())

	cfg, err := config.ReadConfiguration()
	if err != nil {
		util.Printlnf("Failed to read config file: %v", err)
		os.Exit(3)
	}
	if debug {
		util.Printlnf("quiet: %v", cfg.Quiet)
		util.Printlnf("sort_keys: %v", cfg.SortKeys)
		util.Printlnf("path_tilde: %v", cfg.PathTilde)
	}

	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	for _, key := range keys {
		if len(flag.Args()) == 1 {
			pattern := flag.Args()[0]
			if util.Match(key, util.Pattern(pattern)) {
				util.Printlnf("Matched %s vs %s", pattern, key)
			} else {
				util.Printlnf("No match %s vs %s", pattern, key)
			}
		} else {
			util.Printf("%s -> %s\t", yellow(key), red(env[key]))
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
			util.Printlnf("")
		}
	}
}
