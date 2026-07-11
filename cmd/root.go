package cmd

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"

	"github.com/sverrirab/envirou/pkg/config"
	"github.com/sverrirab/envirou/pkg/data"
	"github.com/sverrirab/envirou/pkg/output"
	"github.com/sverrirab/envirou/pkg/shell"

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
		matches, remaining := app.configuration.Groups.MatchAll(app.baseEnv.SortedNames(false), app.caseInsensitive)
		if !showAllGroups && len(actionShowGroups) > 0 {
			for _, actionShowGroup := range actionShowGroups {
				if !displayGroup(app.out, actionShowGroup, matches[actionShowGroup], app.baseEnv, app.sh) {
					output.Printf("%s", app.out.GroupSprintf("# %s (group empty, use -a to show all)\n", actionShowGroup))
				}
			}
		} else {
			notDisplayed := make([]string, 0)
			for _, groupName := range matches.GetAllNames() {
				hideGroup := !showAllGroups && strings.HasPrefix(groupName, ".")
				if hideGroup || !displayGroup(app.out, groupName, matches[groupName], app.baseEnv, app.sh) {
					notDisplayed = append(notDisplayed, groupName)
				}
			}
			displayGroup(app.out, "(no group)", remaining, app.baseEnv, app.sh)
			if len(notDisplayed) > 0 && !app.configuration.SettingsQuiet {
				sort.Strings(notDisplayed)
				output.Printf("%s", app.out.GroupSprintf("# Groups not displayed: %s (use -a to show all)\n", strings.Join(notDisplayed, " ")))
			}
		}
		app.out.PrintProfileList(app.profileNames, app.activeProfileNames)
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Generated completion scripts capture stdout and discard stderr.
		// Normal envirou output stays on stderr so the ev wrapper only
		// evaluates shell mutations, but Cobra's hidden completion protocol
		// must use stdout for every possible target command.
		if cmd.Name() == cobra.ShellCompRequestCmd {
			setCommandOutput(cmd.Root(), os.Stdout)
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if len(app.shellCommands) > 0 {
			commands := app.sh.RunCommands(app.shellCommands)
			if verbose || dryRun {
				diagnosticCommands := app.sh.RunCommands(app.sh.RedactCommandValues(app.shellCommands))
				output.Printf("Shell commands to execute (values redacted):\n>\n> %s>\n", diagnosticCommands)
			}
			if !dryRun {
				fmt.Print(commands)
			}
		}
	},
}

func setCommandOutput(command *cobra.Command, writer io.Writer) {
	command.SetOut(writer)
	for _, child := range command.Commands() {
		setCommandOutput(child, writer)
	}
}

// appState holds runtime state initialized during startup.
type appState struct {
	caseInsensitive      bool
	configuration        *config.Configuration
	sh                   *shell.Shell
	out                  *output.Output
	baseEnv              *data.Profile
	snapshot             *data.Profile
	profileNames         []string
	activeProfileNames   []string
	inactiveProfileNames []string
	isActiveProfile      map[string]bool
	shellCommands        []string
	cryptKey             []byte // verified decryption key cached for this invocation
}

var (
	app appState

	// Initial configuration
	cfgFile             string
	bashBootstrap       string
	bashPromptBootstrap string
	zshPromptBootstrap  string
	powershellBootstrap string
	powershellPrompt    string
	batBootstrap        string

	// Global flags
	verbose            bool
	noColor            bool
	displayUnformatted bool
	outputPowerShell   bool
	dryRun             bool

	// Used by root command
	showAllGroups    bool
	actionShowGroups []string
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(bash, bashPrompt, zshPrompt, powershell, psPrompt, bat string) {
	bashBootstrap = bash
	bashPromptBootstrap = bashPrompt
	zshPromptBootstrap = zshPrompt
	powershellBootstrap = powershell
	powershellPrompt = psPrompt
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.config/envirou/config.ini)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", verbose, "Increase output verbosity")
	rootCmd.PersistentFlags().BoolVarP(&displayUnformatted, "unformatted", "u", displayUnformatted, "Display unformatted env variables")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", noColor, "Disable colored output")
	rootCmd.PersistentFlags().BoolVar(&outputPowerShell, "output-powershell", outputPowerShell, "Enable PowerShell output")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", dryRun, "Only display what would be changed")

	rootCmd.AddGroup(&cobra.Group{ID: "profiles", Title: "Profile commands"})
	rootCmd.AddGroup(&cobra.Group{ID: "encryption", Title: "Encryption commands"})
	rootCmd.AddGroup(&cobra.Group{ID: "groups", Title: "Group commands"})
	rootCmd.AddGroup(&cobra.Group{ID: "configuration", Title: "Configuration commands"})
}

func initConfig() {
	if cfgFile == "" {
		cfgFile = config.GetDefaultConfigFilePath()
	}
	//goland:noinspection ALL
	app.caseInsensitive = runtime.GOOS == "windows"

	var err error
	app.configuration, err = config.ReadConfiguration(cfgFile, app.caseInsensitive)
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
	if app.configuration.SettingsPathTilde {
		replacePathTilde = os.Getenv("HOME")
	}
	if app.caseInsensitive {
		app.sh = shell.NewShell(outputPowerShell, !outputPowerShell)
	} else {
		app.sh = shell.NewShell(false, false)
	}

	app.out = output.NewOutput(replacePathTilde, app.configuration.SettingsPath, app.configuration.SettingsPassword, displayUnformatted, app.caseInsensitive, app.configuration.FormatGroup, app.configuration.FormatProfile, app.configuration.FormatEnvName, app.configuration.FormatPath, app.configuration.FormatDiff)

	app.baseEnv = data.NewProfile(app.caseInsensitive)
	app.baseEnv.MergeStrings(os.Environ())

	app.snapshot, _ = config.LoadSnapshot(app.caseInsensitive)
	if app.snapshot != nil {
		added, changed, _ := data.FullDiff(app.baseEnv, app.snapshot)
		diffNames := make(map[string]bool)
		for _, n := range added {
			diffNames[n] = true
		}
		for _, n := range changed {
			diffNames[n] = true
		}
		app.out.SetDiffNames(diffNames)
	}

	// When unlocked (valid ENVIROU_KEY), decrypt profile values up front so
	// active-profile detection and all commands see plaintext. Locked
	// profiles keep their tokens and simply compare as inactive; commands
	// that apply them prompt via ensureKey.
	decryptProfilesSilently(app.configuration.Profiles)

	// Figure out what profiles are active.
	app.profileNames = make([]string, 0, len(app.configuration.Profiles))
	app.activeProfileNames = make([]string, 0, len(app.configuration.Profiles))
	app.inactiveProfileNames = make([]string, 0, len(app.configuration.Profiles))
	app.isActiveProfile = make(map[string]bool)

	for name, profile := range app.configuration.Profiles {
		app.profileNames = append(app.profileNames, name)
		if app.baseEnv.IsMerged(&profile) {
			app.activeProfileNames = append(app.activeProfileNames, name)
			app.isActiveProfile[name] = true
		} else {
			app.inactiveProfileNames = append(app.inactiveProfileNames, name)
			app.isActiveProfile[name] = false
		}
	}
	sort.Strings(app.profileNames)
	sort.Strings(app.activeProfileNames)
	sort.Strings(app.inactiveProfileNames)

	app.shellCommands = make([]string, 0)
}
