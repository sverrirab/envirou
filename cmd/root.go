package cmd

import (
	"github.com/sverrirab/envirou/pkg/config"
	"github.com/sverrirab/envirou/pkg/data"
	"github.com/sverrirab/envirou/pkg/output"
	"github.com/sverrirab/envirou/pkg/shell"
	"os"
	"runtime"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

func displayGroup(out *output.Output, name string, envs data.Envs, profile *data.Profile, sh *shell.Shell) bool {
	if len(envs) > 0 {
		out.PrintGroup(name)
		for _, env := range envs {
			value, _ := profile.Get(env)
			out.PrintEnv(sh, env, value)
		}
		return true
	}
	return false
}

func findProfile(out *output.Output, cfg *config.Configuration, name string) (*data.Profile, bool) {
	profile, found := cfg.Profiles.FindProfile(name)
	if !found {
		output.Printf("Profile %s not found\n", out.DiffSprintf(name))
	}
	return profile, found
}

var rootCmd = &cobra.Command{
	Use:   "envirou",
	Short: "Manage and view your shell environment variables",
	Long: `You can view categorized and sorted env variables as well as easily
create custom profiles to quickly identify the local configuration

Using without any command will display the current environment in a shortened format. 
The next step is to use "set" to modify the current environment (this requires "ev" 
shell function to be installed)`,
	Run: func(cmd *cobra.Command, args []string) {
		matches, remaining := configuration.Groups.MatchAll(baseEnv.SortedNames(false))
		if !showAllGroups && len(actionShowGroups) > 0 {
			for _, actionShowGroup := range actionShowGroups {
				if !displayGroup(out, actionShowGroup, matches[actionShowGroup], baseEnv, sh) {
					output.Printf(out.GroupSprintf("# %s (group empty, use -a to show all)\n", actionShowGroup))
				}
			}
		} else {
			notDisplayed := make([]string, 0)
			for _, groupName := range matches.GetAllNames() {
				hideGroup := !showAllGroups && strings.HasPrefix(groupName, ".")
				if hideGroup || !displayGroup(out, groupName, matches[groupName], baseEnv, sh) {
					notDisplayed = append(notDisplayed, groupName)
				}
			}
			displayGroup(out, "(no group)", remaining, baseEnv, sh)
			if len(notDisplayed) > 0 && !configuration.SettingsQuiet {
				sort.Strings(notDisplayed)
				output.Printf(out.GroupSprintf("# Groups not displayed: %s (use -a to show all)\n", strings.Join(notDisplayed, " ")))
			}
		}
		out.PrintProfileList(profileNames, activeProfileNames)
	},
}

var (
	// Initial configuration
	cfgFile       string
	configuration *config.Configuration

	// Global flags
	verbose            bool = false
	noColor            bool = false
	displayUnformatted bool = false
	outputPowerShell   bool = false

	// These are initialized using the current configuration
	sh  *shell.Shell
	out *output.Output

	// Current environment
	baseEnv *data.Profile

	profileNames         []string
	activeProfileNames   []string
	inactiveProfileNames []string
	isActiveProfile      map[string]bool

	// Used by root command
	showAllGroups    bool = false
	actionShowGroups []string
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Flags for root command
	rootCmd.Flags().BoolVarP(&showAllGroups, "all", "a", showAllGroups, "List all groups")
	rootCmd.Flags().StringArrayVarP(&actionShowGroups, "group", "g", nil, "Show individual group")
	rootCmd.MarkFlagsMutuallyExclusive("all", "group")

	// Flags for all commands
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.envirou/config.ini)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", verbose, "Increase output verbosity")
	rootCmd.PersistentFlags().BoolVarP(&displayUnformatted, "unformatted", "u", displayUnformatted, "Display unformatted env variables")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", noColor, "Disable colored output")

	rootCmd.AddGroup(&cobra.Group{ID: "profiles", Title: "Profile commands"})
	rootCmd.AddGroup(&cobra.Group{ID: "groups", Title: "Group commands"})
}

func initConfig() {
	if cfgFile == "" {
		cfgFile = config.GetDefaultConfigFilePath()
	}
	var err error
	configuration, err = config.ReadConfiguration(cfgFile)
	if err != nil {
		output.Printf("Failed to read config file: %v\n", err)
		os.Exit(3)
	}
	output.Printf("Read config file: %s\n", cfgFile)

	// Display modifiers
	output.NoColor(noColor)
	replacePathTilde := ""
	runningOnWindows := runtime.GOOS != "windows"
	if runningOnWindows && configuration.SettingsPathTilde {
		replacePathTilde = os.Getenv("HOME")
	}
	sh = shell.NewShell(outputPowerShell, runningOnWindows && !outputPowerShell)
	out = output.NewOutput(replacePathTilde, configuration.SettingsPath, configuration.SettingsPassword, displayUnformatted, configuration.FormatGroup, configuration.FormatProfile, configuration.FormatEnvName, configuration.FormatPath, configuration.FormatDiff)

	baseEnv = data.NewProfile()
	baseEnv.MergeStrings(os.Environ())

	// Figure out what profiles are active.
	profileNames = make([]string, 0, len(configuration.Profiles))
	activeProfileNames = make([]string, 0, len(configuration.Profiles))
	inactiveProfileNames = make([]string, 0, len(configuration.Profiles))
	isActiveProfile = make(map[string]bool)

	for name, profile := range configuration.Profiles {
		profileNames = append(profileNames, name)
		if baseEnv.IsMerged(&profile) {
			activeProfileNames = append(activeProfileNames, name)
			isActiveProfile[name] = true
		} else {
			inactiveProfileNames = append(inactiveProfileNames, name)
			isActiveProfile[name] = false
		}
	}
	sort.Strings(profileNames)
	sort.Strings(activeProfileNames)
	sort.Strings(inactiveProfileNames)
	// output.Printf("profileNames: %s\n", profileNames)
	// output.Printf("activeProfileNames: %s\n", activeProfileNames)
	// output.Printf("inactiveProfileNames: %s\n", inactiveProfileNames)
}
