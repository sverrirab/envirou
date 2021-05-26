package config

import (
	"github.com/sverrirab/envirou/pkg/util"
	"github.com/zieckey/goini"
)

type Configuration struct {
	Quiet     bool
	SortKeys  bool
	PathTilde bool
	Groups    map[string]util.Patterns
}

func ReadConfiguration() (*Configuration, error) {
	configuration := &Configuration{
		Quiet:     true,
		SortKeys:  true,
		PathTilde: true,
		Groups:    make(map[string]util.Patterns),
	}
	ini := goini.New()
	configPath := GetDefaultConfigFilePath()
	err := ini.ParseFile(configPath)
	if err != nil {
		util.Printlnf("Missing configfile? [%s]", configPath, err.Error())
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
	util.Printlnf("groups: %v", groups)
	if ok {
		for k, v := range groups {
			util.Printlnf("group %s -> %s", k, v)
			configuration.Groups[k] = *util.ParsePatterns(v)
		}
	}
	return configuration, nil
}
