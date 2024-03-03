package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:     "set PROFILE1 [PROFILE2] ...",
	Aliases: []string{"."},
	Short:   "Modify current environment with profile",
	Long: `Each profile will be merged with your current environment

To change profiles edit the config file (see "config" command)`,
	// ValidArgs: []string{"one", "two", "three"},
	GroupID: "profiles",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("set command! %s\n", cmd.Flag("config").Value)
	},
}

func init() {
	// profilesCmd.AddCommand(setCmd)
	// Add the new 'example' Group to the root command
	rootCmd.AddCommand(setCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
