package config

import (
	"sort"

	"github.com/sverrirab/envirou/pkg/util"
	"github.com/zieckey/goini"
)

type Configuration struct {
	Quiet        bool
	SortKeys     bool
	PathTilde    bool
	Groups       map[string]util.Patterns
	GroupsSorted []string
}

func ReadConfiguration(configPath string) (*Configuration, error) {
	configuration := &Configuration{
		Quiet:        false,
		SortKeys:     false,
		PathTilde:    false,
		Groups:       make(map[string]util.Patterns),
		GroupsSorted: make([]string, 0, 10),
	}
	ini := goini.New()
	err := ini.ParseFile(configPath)
	if err != nil {
		err := WriteDefaultConfigFile(configPath)
		if err != nil {
			util.Printlnf("Writing config file failed [%s]", err.Error())
			return configuration, err
		}
		// Read again now that we have written the default file
		err = ini.ParseFile(configPath)
		if err != nil {
			util.Printlnf("Parsing of %s failed [%s]", configPath, err.Error())
			return configuration, err
		}
	}
	configuration.Quiet, _ = ini.SectionGetBool("settings", "quiet")
	configuration.SortKeys, _ = ini.SectionGetBool("settings", "sort_keys")
	configuration.PathTilde, _ = ini.SectionGetBool("settings", "path_tilde")

	// Groups
	groups, ok := ini.GetKvmap("groups")
	if ok {
		for k, v := range groups {
			configuration.Groups[k] = *util.ParsePatterns(v)
		}
	}
	custom, ok := ini.GetKvmap("custom")
	if ok {
		for k, v := range custom {
			configuration.Groups[k] = *util.ParsePatterns(v)
		}
	}
	// Create sorted list of groups
	for k, _ := range configuration.Groups {
		configuration.GroupsSorted = append(configuration.GroupsSorted, k)
	}
	sort.Strings(configuration.GroupsSorted)

	return configuration, nil
}
