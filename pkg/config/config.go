package config

import (
	"strings"

	"github.com/sverrirab/envirou/pkg/ini"
	"github.com/sverrirab/envirou/pkg/data"
)

type Configuration struct {
	Quiet        bool
	SortKeys     bool
	PathTilde    bool
	Groups       data.Groups
	Profiles     map[string]data.Profile
}

func ReadConfiguration(configPath string) (*Configuration, error) {
	configuration := &Configuration{
		Quiet:        false,
		SortKeys:     false,
		PathTilde:    false,
		Groups:       make(data.Groups),
		Profiles:     make(map[string]data.Profile),
	}
	config, err := ini.NewIni(configPath)
	if err != nil {
		err := WriteDefaultConfigFile(configPath)
		if err != nil {
			return configuration, err
		}
		// Read again now that we have written the default file
		config, err = ini.NewIni(configPath)
		if err != nil {
			return configuration, err
		}
	}
	configuration.Quiet = config.GetBool("settings", "quiet", false)
	configuration.SortKeys = config.GetBool("settings", "sort_keys", true)
	configuration.PathTilde = config.GetBool("settings", "path_tilde", true)

	// Groups
	groups := config.GetAllVariables("groups")
	for _, k := range groups {
		configuration.Groups.ParseAndAdd(k, config.GetString("groups", k, ""))
	}
	custom := config.GetAllVariables("custom")
	for _, k := range custom {
		configuration.Groups.ParseAndAdd(k, config.GetString("custom", k, ""))
	}

	sections := config.GetAllSections()
	for _, section := range sections {
		split := strings.SplitN(section, ":", 2)
		if len(split) == 2 && strings.TrimSpace(strings.ToLower(split[0])) == "profile" {
			profileName := strings.TrimSpace(split[1])
			profile := data.NewProfile()
			for _, entry := range config.GetAllVariables(section) {
				if config.IsNil(section, entry) {
					profile.SetNil(entry)
				} else {
					profile.Set(entry, config.GetString(section, entry, ""))
				}
			}
			configuration.Profiles[profileName] = *profile
		}
	}

	return configuration, nil
}
