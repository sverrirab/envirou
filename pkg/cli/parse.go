package cli

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

type Flag struct {
	// Actions:
	actionShowGroup                 string
	actionDiffProfile               string
	actionListGroups                bool
	actionListProfiles              bool
	actionListProfilesActive        bool
	actionListProfilesActiveColored bool
	actionListProfilesInactive      bool
	actionEditConfig                bool

	// Display modifiers
	showAllGroups      bool
	displayUnformatted bool
	verbose            bool
	noColor            bool

	// Bootstrap shell
	bootstrapPowerShell bool
	bootstrapBash       bool

	// Shell overrides
	outputPowerShell bool
}

func ParseCommandLine() {
	app := &cli.App{
		Name:  "envirou",
		Usage: "Manage your shell environment",
		Action: func(*cli.Context) error {
			fmt.Println("boom! I say!")
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
