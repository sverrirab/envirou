package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// completionCmd replaces cobra's auto-generated completion command.
// The auto-generated one inherits the root command's stderr output, so
// "envirou completion bash > file" would produce an empty file. This
// version writes the script to stdout for redirection.
var completionCmd = &cobra.Command{
	Use:       "completion [bash|zsh|fish|powershell]",
	Short:     "Output shell completion script",
	Long:      `Output a tab completion script for the specified shell to stdout. See docs/completion.md for installation instructions.`,
	GroupID:   "configuration",
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("only provide one argument: type of shell")
		}
		if !contains(cmd.ValidArgs, args[0]) {
			validArgs := strings.Join(cmd.ValidArgs, ", ")
			return fmt.Errorf("invalid argument \"%s\", must be one of %s", args[0], validArgs)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// os.Stdout, not cmd.OutOrStdout(): the latter inherits the
		// root command's stderr redirection.
		out := os.Stdout
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletionV2(out, true)
		case "zsh":
			return rootCmd.GenZshCompletion(out)
		case "fish":
			return rootCmd.GenFishCompletion(out, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(out)
		}
		return nil
	},
}

func init() {
	// Registered without addCommand on purpose: the script must go to
	// stdout, not stderr.
	rootCmd.AddCommand(completionCmd)
}
