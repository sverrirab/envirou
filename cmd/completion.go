package cmd

import (
	"fmt"
	"io"
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
		return writeCompletion(args[0], os.Stdout)
	},
}

func writeCompletion(shellName string, out io.Writer) error {
	var err error
	switch shellName {
	case "bash":
		err = rootCmd.GenBashCompletionV2(out, true)
	case "zsh":
		err = rootCmd.GenZshCompletion(out)
	case "fish":
		return rootCmd.GenFishCompletion(out, true)
	case "powershell":
		err = rootCmd.GenPowerShellCompletionWithDesc(out)
	}
	if err != nil {
		return err
	}

	// The shell wrappers pass Cobra's hidden completion protocol through to
	// envirou, so the same generated completer can serve the ev function.
	switch shellName {
	case "bash":
		_, err = fmt.Fprintln(out, "complete -o default -F __start_envirou ev")
	case "zsh":
		_, err = fmt.Fprintln(out, "compdef _envirou ev")
	case "powershell":
		_, err = fmt.Fprintln(out, "Register-ArgumentCompleter -CommandName 'ev' -ScriptBlock ${__envirouCompleterBlock}")
	}
	return err
}

func init() {
	// Registered without addCommand on purpose: the script must go to
	// stdout, not stderr.
	rootCmd.AddCommand(completionCmd)
}
