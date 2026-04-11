package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/sverrirab/envirou/pkg/data"
	"github.com/sverrirab/envirou/pkg/output"
)

var (
	findNameOnly   bool
	findValueOnly  bool
	findIgnoreCase bool
)

var findCmd = &cobra.Command{
	Use:     "find PATTERN",
	Aliases: []string{"search"},
	Short:   "Find environment variables matching a pattern",
	Long: `Search environment variable names and values for a glob match.

Patterns support * as a wildcard:
  PATH      substring — matches PATH, CLASSPATH, PATH_INFO, ...
  PATH*     prefix   — matches PATH, PATH_INFO but not CLASSPATH
  *PATH     suffix   — matches PATH, CLASSPATH but not PATH_INFO
  *         matches everything

Quote patterns containing * to prevent shell expansion:
  ev find 'PATH*'     (correct)
  ev find PATH*       (may break — shell expands * against filenames)

By default both names and values are searched. Use --name or --value to restrict.`,
	GroupID: "profiles",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		searchPattern := args[0]
		caseInsensitive := app.caseInsensitive || findIgnoreCase

		// If the pattern contains no wildcards, treat it as a substring match
		pattern := searchPattern
		if !strings.Contains(pattern, "*") {
			pattern = "*" + pattern + "*"
		}
		if caseInsensitive {
			pattern = strings.ToUpper(pattern)
		}
		p := data.Pattern(pattern)

		// Default: search both. --name or --value restricts.
		searchName := !findValueOnly
		searchValue := !findNameOnly

		count := 0
		for _, name := range app.baseEnv.SortedNames(false) {
			value, _ := app.baseEnv.Get(name)
			matched := false
			if searchName && data.Match(name, p, caseInsensitive) {
				matched = true
			}
			if searchValue && data.Match(value, p, caseInsensitive) {
				matched = true
			}
			if matched {
				app.out.PrintEnv(app.sh, name, value)
				count++
			}
		}
		if count == 0 {
			output.Printf("No matches found\n")
		}
	},
}

func init() {
	findCmd.Flags().BoolVar(&findNameOnly, "name", false, "Search names only")
	findCmd.Flags().BoolVar(&findValueOnly, "value", false, "Search values only")
	findCmd.Flags().BoolVarP(&findIgnoreCase, "ignore-case", "i", false, "Force case-insensitive search")
	findCmd.MarkFlagsMutuallyExclusive("name", "value")
	addCommand(findCmd)
}
