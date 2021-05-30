package config

import (
	"sort"

	"github.com/sverrirab/envirou/pkg/util"
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
	ini, err := util.NewIni(configPath)
	if err != nil {
		err := WriteDefaultConfigFile(configPath)
		if err != nil {
			util.Printlnf("Writing config file failed [%s]", err.Error())
			return configuration, err
		}
		// Read again now that we have written the default file
		ini, err = util.NewIni(configPath)
		if err != nil {
			util.Printlnf("Parsing of %s failed [%s]", configPath, err.Error())
			return configuration, err
		}
	}
	configuration.Quiet = ini.GetBool("settings", "quiet", false)
	configuration.SortKeys = ini.GetBool("settings", "sort_keys", true)
	configuration.PathTilde = ini.GetBool("settings", "path_tilde", true)

	// Groups
	groups:= ini.GetAllVariables("groups")
	for _, k := range groups {
		configuration.Groups[k] = *util.ParsePatterns(ini.GetString("groups", k, ""))
	}
	custom := ini.GetAllVariables("custom")
	for _, k := range custom {
		configuration.Groups[k] = *util.ParsePatterns(ini.GetString("custom", k, ""))
	}
	// Create sorted list of groups
	for k := range configuration.Groups {
		configuration.GroupsSorted = append(configuration.GroupsSorted, k)
	}
	sort.Strings(configuration.GroupsSorted)

	return configuration, nil
}
