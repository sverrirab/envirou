package cmd

import (
	"fmt"
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
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if len(shellCommands) > 0 {
			commands := sh.RunCommands(shellCommands)
			if verbose || dryRun {
				output.Printf("Shell commands to execute:\n>\n> %s>\n", commands)
			}
			if !dryRun {
				fmt.Print(commands)
			}
		}
	},
}

var (
	// Initial configuration
	cfgFile             string
	configuration       *config.Configuration
	bashBootstrap       string
	powershellBootstrap string
	batBootstrap        string

	// Global flags
	verbose            bool = false
	noColor            bool = false
	displayUnformatted bool = false
	outputPowerShell   bool = false
	dryRun             bool = false

	// These are initialized using the current configuration
	sh  *shell.Shell
	out *output.Output

	// Current environment
	baseEnv *data.Profile

	profileNames         []string
	activeProfileNames   []string
	inactiveProfileNames []string
	isActiveProfile      map[string]bool

	// Here is what will be executed
	shellCommands []string

	// Used by root command
	showAllGroups    bool = false
	actionShowGroups []string
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(bash, powershell, bat string) {
	bashBootstrap = bash
	powershellBootstrap = powershell
	batBootstrap = bat

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func addCommand(command *cobra.Command) {
	// We need to redirect all output to stderr for all commands so this does not conflict with
	// the outputting of shell execution of stderr
	command.SetOut(os.Stderr)
	rootCmd.AddCommand(command)
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.SetOut(os.Stderr)

	// Flags for root command
	rootCmd.Flags().BoolVarP(&showAllGroups, "all", "a", showAllGroups, "List all groups")
	rootCmd.Flags().StringArrayVarP(&actionShowGroups, "group", "g", nil, "Show individual group")
	rootCmd.MarkFlagsMutuallyExclusive("all", "group")

	// Flags for all commands
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.envirou/config.ini)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", verbose, "Increase output verbosity")
	rootCmd.PersistentFlags().BoolVarP(&displayUnformatted, "unformatted", "u", displayUnformatted, "Display unformatted env variables")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", noColor, "Disable colored output")
	rootCmd.PersistentFlags().BoolVar(&outputPowerShell, "output-powershell", outputPowerShell, "Enable PowerShell output")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", dryRun, "Only display what would be changed")

	rootCmd.AddGroup(&cobra.Group{ID: "profiles", Title: "Profile commands"})
	rootCmd.AddGroup(&cobra.Group{ID: "groups", Title: "Group commands"})
	rootCmd.AddGroup(&cobra.Group{ID: "configuration", Title: "Configuration commands"})
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
	if verbose {
		output.Printf("Read config file: %s\n", cfgFile)
	}

	// Display modifiers
	output.NoColor(noColor)
	replacePathTilde := ""
	//goland:noinspection ALL
	if runtime.GOOS == "windows" {
		data.SetCaseInsensitive()
		if configuration.SettingsPathTilde {
			replacePathTilde = os.Getenv("HOME")
		}
		sh = shell.NewShell(outputPowerShell, !outputPowerShell)
	} else {
		sh = shell.NewShell(false, false)
	}

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

	shellCommands = make([]string, 0)
}
