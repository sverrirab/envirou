package cli

import (
	"github.com/sverrirab/envirou/pkg/config"
	"github.com/sverrirab/envirou/pkg/output"
	"github.com/urfave/cli/v2"
	"io"
	"os"
	"runtime"
)

type Flags struct {
	// Actions:
	ActionShowGroup                 string
	ActionDiffProfile               string
	ActionListGroups                bool
	ActionListProfiles              bool
	ActionListProfilesActive        bool
	ActionListProfilesActiveColored bool
	ActionListProfilesInactive      bool
	ActionEditConfig                bool
	ActionActivateProfile           string

	// Display modifiers
	ShowAllGroups      bool
	DisplayUnformatted bool
	Verbose            bool
	NoColor            bool

	// Bootstrap shell
	BootstrapPowerShell bool
	BootstrapBash       bool

	// Shell overrides
	OutputBat        bool
	OutputBash       bool
	OutputPowerShell bool
}

func getBoolText(b bool) string {
	if b {
		return "true"
	} else {
		return "false"
	}
}

func ParseCommandLine(cfg *config.Configuration) Flags {
	flags := Flags{}
	helpOnly := true

	// Guess which shell is in use
	if runtime.GOOS == "windows" {
		flags.OutputBat = true
	} else {
		flags.OutputBash = true
	}

	cli.HelpPrinter = func(w io.Writer, templ string, data interface{}) {
		// Redirect help output to stderr
		cli.HelpPrinterCustom(os.Stderr, templ, data, nil)
	}

	allProfileNames := cfg.GetAllProfileNames()
	commands := make([]*cli.Command, 0, len(allProfileNames))
	for _, profileName := range allProfileNames {
		commands = append(commands, &cli.Command{
			Name:     profileName,
			Category: "profiles",
			Action: func(c *cli.Context) error {
				helpOnly = false
				flags.ActionActivateProfile = c.Command.Name
				return nil
			},
		})
	}

	app := &cli.App{
		Name:  "envirou",
		Usage: "Manage your shell environment",
		Flags: []cli.Flag{
			// Profile Category
			&cli.BoolFlag{
				Name:        "profiles",
				Aliases:     []string{"p"},
				Usage:       "List profile names",
				Category:    "Profiles:",
				Destination: &flags.ActionListProfiles,
			},
			&cli.BoolFlag{
				Name:        "active-profiles",
				Usage:       "List active profiles only",
				Category:    "Profiles:",
				Destination: &flags.ActionListProfilesActive,
				Hidden:      true,
			},
			&cli.BoolFlag{
				Name:        "active-profiles-colored",
				Usage:       "List active profiles only (w/color)",
				Category:    "Profiles:",
				Destination: &flags.ActionListProfilesActiveColored,
				Hidden:      true,
			},
			&cli.BoolFlag{
				Name:        "inactive-profiles",
				Usage:       "List inactive profiles only",
				Category:    "Profiles:",
				Destination: &flags.ActionListProfiles,
				Hidden:      true,
			},
			&cli.StringFlag{
				Name:     "diff",
				Aliases:  []string{"d"},
				Usage:    "Show changes in current env from specific profile",
				Category: "Profiles:",
			},

			// Groups Category
			&cli.StringFlag{
				Name:     "group",
				Aliases:  []string{"g"},
				Usage:    "Show env variables from `GROUP` only",
				Category: "Groups:",
			},
			&cli.BoolFlag{
				Name:        "all",
				Aliases:     []string{"a"},
				Usage:       "Show all (including .hidden) groups",
				Category:    "Groups:",
				Destination: &flags.ShowAllGroups,
			},
			&cli.BoolFlag{
				Name:        "list",
				Aliases:     []string{"l"},
				Usage:       "List group names",
				Category:    "Groups:",
				Destination: &flags.ActionListGroups,
			},

			// Configuration Category
			&cli.BoolFlag{
				Name:        "edit",
				Usage:       "Edit configuration",
				Category:    "Configuration:",
				Destination: &flags.ActionEditConfig,
			},

			// Display Category
			&cli.BoolFlag{
				Name:        "no-color",
				Usage:       "disable colored output",
				Category:    "Display:",
				Destination: &flags.NoColor,
			},
			&cli.BoolFlag{
				Name:        "unformatted",
				Aliases:     []string{"u"},
				Usage:       "Display unformatted/raw env variables",
				Category:    "Display:",
				Destination: &flags.DisplayUnformatted,
			},
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Usage:       "Increase output verbosity",
				Category:    "Display:",
				Destination: &flags.Verbose,
			},

			// Shell Category
			&cli.BoolFlag{
				Name:        "bootstrap-powershell",
				Usage:       "Enable PowerShell support (ev function)",
				Category:    "Shell:",
				Destination: &flags.BootstrapPowerShell,
			},
			&cli.BoolFlag{
				Name:        "bootstrap-bash",
				Usage:       "Enable Bash support (ev function)",
				Category:    "Shell:",
				Destination: &flags.BootstrapBash,
			},
			&cli.BoolFlag{
				Name:        "output-bat",
				Usage:       "Enable windows bat output",
				DefaultText: getBoolText(flags.OutputBat),
				Category:    "Shell:",
				Action: func(ctx *cli.Context, b bool) error {
					if b {
						flags.OutputBat = true
						flags.OutputBash = false
						flags.OutputPowerShell = false
					}
					return nil
				},
			},
			&cli.BoolFlag{
				Name:        "output-bash",
				Usage:       "Enable Bash/zsh output",
				Aliases:     []string{"sh"},
				DefaultText: getBoolText(flags.OutputBash),
				Category:    "Shell:",
				Destination: &flags.OutputBash,
				Action: func(ctx *cli.Context, b bool) error {
					if b {
						flags.OutputBat = false
						flags.OutputBash = true
						flags.OutputPowerShell = false
					}
					return nil
				},
			},
			&cli.BoolFlag{
				Name:        "output-powershell",
				Usage:       "Enable PowerShell output",
				Aliases:     []string{"ps1"},
				DefaultText: getBoolText(flags.OutputPowerShell),
				Category:    "Shell:",
				Destination: &flags.OutputPowerShell,
				Action: func(ctx *cli.Context, b bool) error {
					if b {
						flags.OutputBat = false
						flags.OutputBash = false
						flags.OutputPowerShell = true
					}
					return nil
				},
			},
		},
		Commands: commands,
		Action: func(*cli.Context) error {
			helpOnly = false
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		output.Printf("%v\n", err)
		os.Exit(3)
	}
	if helpOnly {
		os.Exit(0)
	}
	return flags
}
