package config

import (
	"strings"

	"github.com/sverrirab/envirou/pkg/data"
	"github.com/sverrirab/envirou/pkg/ini"
	"github.com/sverrirab/envirou/pkg/output"
)

type Configuration struct {
	SettingsQuiet     bool
	SettingsSortKeys  bool
	SettingsPathTilde bool
	SettingsPassword  data.Patterns
	SettingsPath      data.Patterns

	FormatGroup   string
	FormatProfile string
	FormatEnvName string
	FormatPath    string
	FormatDiff    string

	Groups   data.Groups
	Profiles data.Profiles
}

func readFormat(config *ini.IniFile, name, defaultValue string) string {
	value := config.GetString("format", name, defaultValue)
	if output.IsValidColor(value) {
		return value
	} else {
		return defaultValue
	}
}

func ReadConfiguration(configPath string, caseInsensitive bool) (*Configuration, error) {
	configuration := &Configuration{
		SettingsQuiet:     false,
		SettingsSortKeys:  false,
		SettingsPathTilde: false,
		Groups:            make(data.Groups),
		Profiles:          make(data.Profiles),
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
	configuration.SettingsQuiet = config.GetBool("settings", "quiet", false)
	configuration.SettingsSortKeys = config.GetBool("settings", "sort_keys", true)
	configuration.SettingsPathTilde = config.GetBool("settings", "path_tilde", true)
	configuration.SettingsPassword = *data.ParsePatterns(config.GetString("settings", "password", ""), caseInsensitive)
	configuration.SettingsPath = *data.ParsePatterns(config.GetString("settings", "path", ""), caseInsensitive)

	configuration.FormatGroup = readFormat(config, "group", "magenta")
	configuration.FormatProfile = readFormat(config, "profile", "green")
	configuration.FormatEnvName = readFormat(config, "env_name", "cyan")
	configuration.FormatPath = readFormat(config, "path", "reverse")
	configuration.FormatDiff = readFormat(config, "diff", "red")

	// Groups
	groups := config.GetAllVariables("groups")
	for _, k := range groups {
		configuration.Groups.ParseAndAdd(k, config.GetString("groups", k, ""), caseInsensitive)
	}
	custom := config.GetAllVariables("custom")
	for _, k := range custom {
		configuration.Groups.ParseAndAdd(k, config.GetString("custom", k, ""), caseInsensitive)
	}

	sections := config.GetAllSections()
	for _, section := range sections {
		split := strings.SplitN(section, ":", 2)
		if len(split) == 2 && strings.TrimSpace(strings.ToLower(split[0])) == "profile" {
			profileName := strings.TrimSpace(split[1])
			profile := data.NewProfile(caseInsensitive)
			for _, entry := range config.GetAllVariables(section) {
				if config.IsNil(section, entry) {
					profile.SetNil(entry)
				} else {
					op := config.GetOperator(section, entry)
					mode := data.MergeReplace
					switch op {
					case ini.OpPrepend:
						mode = data.MergePrepend
					case ini.OpAppend:
						mode = data.MergeAppend
					}
					profile.SetWithMode(entry, config.GetString(section, entry, ""), mode)
				}
			}
			configuration.Profiles[profileName] = *profile
		}
	}

	return configuration, nil
}
