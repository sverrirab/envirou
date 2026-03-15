package cmd

import (
	"regexp"

	"github.com/spf13/cobra"
	"github.com/sverrirab/envirou/pkg/output"
)

var (
	findNameOnly   bool
	findValueOnly  bool
	findIgnoreCase bool
	findRegex      bool
)

var findCmd = &cobra.Command{
	Use:     "find PATTERN",
	Aliases: []string{"search"},
	Short:   "Find environment variables matching a pattern",
	Long: `Search environment variable names and values for a substring or regex match.

By default both names and values are searched. Use --name or --value to restrict.`,
	GroupID: "profiles",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		searchPattern := args[0]
		caseInsensitive := app.caseInsensitive || findIgnoreCase

		matcher, err := newMatcher(searchPattern, caseInsensitive, findRegex)
		if err != nil {
			output.Printf("Invalid regex pattern: %v\n", err)
			return
		}

		// Default: search both. --name or --value restricts.
		searchName := !findValueOnly
		searchValue := !findNameOnly

		count := 0
		for _, name := range app.baseEnv.SortedNames(false) {
			value, _ := app.baseEnv.Get(name)
			matched := false
			if searchName && matcher.match(name) {
				matched = true
			}
			if searchValue && matcher.match(value) {
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

type findMatcher struct {
	re *regexp.Regexp
}

func newMatcher(pattern string, caseInsensitive bool, useRegex bool) (*findMatcher, error) {
	var expr string
	if useRegex {
		expr = pattern
	} else {
		expr = regexp.QuoteMeta(pattern)
	}
	if caseInsensitive {
		expr = "(?i)" + expr
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	return &findMatcher{re: re}, nil
}

func (m *findMatcher) match(s string) bool {
	return m.re.MatchString(s)
}

func init() {
	findCmd.Flags().BoolVar(&findNameOnly, "name", false, "Search names only")
	findCmd.Flags().BoolVar(&findValueOnly, "value", false, "Search values only")
	findCmd.Flags().BoolVarP(&findIgnoreCase, "ignore-case", "i", false, "Force case-insensitive search")
	findCmd.Flags().BoolVarP(&findRegex, "regex", "r", false, "Use regex instead of substring match (quote your pattern to avoid shell expansion)")
	findCmd.MarkFlagsMutuallyExclusive("name", "value")
	addCommand(findCmd)
}
