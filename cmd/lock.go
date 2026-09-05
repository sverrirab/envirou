package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/sverrirab/envirou/pkg/config"
	"github.com/sverrirab/envirou/pkg/crypt"
	"github.com/sverrirab/envirou/pkg/output"
)

var unlockPrintKey bool

var unlockCmd = &cobra.Command{
	Use:   "unlock",
	Short: "Unlock encrypted values for this shell session",
	Long: `Prompt for the passphrase once and store the derived key in the ENVIROU_KEY
shell variable so encrypted profiles can be applied without further prompts
(requires the "ev" wrapper). The variable is not exported: programs started
from the shell do not inherit it; the wrapper passes it to envirou only.
Run "ev lock" to clear the key again.`,
	GroupID: "encryption",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		material, err := crypt.LoadMaterial(config.GetDefaultConfigFileFolder())
		if err != nil {
			if errors.Is(err, crypt.ErrNotInitialized) {
				err = notInitializedError()
			}
			exitOnError(err)
		}
		passphrase, err := crypt.ReadPassphrase("Envirou passphrase: ")
		exitOnError(err)
		key, err := crypt.DeriveKey(passphrase, material.Salt, material.Iterations)
		exitOnError(err)
		if err := material.VerifyKey(key); err != nil {
			exitOnError(errors.New("incorrect passphrase"))
		}
		if unlockPrintKey {
			fmt.Println(crypt.EncodeKey(key))
			return
		}
		app.shellCommands = append(app.shellCommands, app.sh.LocalVar("ENVIROU_KEY", crypt.EncodeKey(key)))
		output.Printf("Unlocked - encrypted values now decrypt automatically. Run 'ev lock' to clear.\n")
	},
}

var lockCmd = &cobra.Command{
	Use:     "lock",
	Short:   "Clear the encryption key from this shell session",
	Long:    `Clear the ENVIROU_KEY shell variable so encrypted profiles prompt for the passphrase again.`,
	GroupID: "encryption",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if os.Getenv("ENVIROU_KEY") == "" {
			output.Printf("Already locked\n")
			return
		}
		app.shellCommands = append(app.shellCommands, app.sh.UnsetLocalVar("ENVIROU_KEY"))
		output.Printf("Locked\n")
	},
}

func init() {
	unlockCmd.Flags().BoolVar(&unlockPrintKey, "print-key", false, "Print the derived key to stdout instead of exporting it (for CI secret stores)")
	addCommand(unlockCmd)
	addCommand(lockCmd)
}
