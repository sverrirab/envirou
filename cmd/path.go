package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sverrirab/envirou/pkg/output"
)

var pathCheck bool

var pathCmd = &cobra.Command{
	Use:     "path [VAR]",
	Short:   "Display path-like variables with one entry per line",
	Long:    `Show path-like variables split into individual entries. Use --check to flag missing directories and duplicates.`,
	GroupID: "profiles",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var names []string
		if len(args) == 1 {
			// Show specific variable
			name := args[0]
			if _, ok := app.baseEnv.Get(name); !ok {
				output.Printf("Variable %s not found\n", app.out.DiffSprintf(name))
				return
			}
			names = []string{name}
		} else {
			// Show all path-like variables
			for _, name := range app.baseEnv.SortedNames(false) {
				if app.out.IsPathVariable(name) {
					names = append(names, name)
				}
			}
		}

		if len(names) == 0 {
			output.Printf("No path variables found\n")
			return
		}

		sep := string(os.PathListSeparator)
		for _, name := range names {
			value, _ := app.baseEnv.Get(name)
			parts := strings.Split(value, sep)

			output.Printf("%s\n", app.out.EnvNameSprintf("# %s", name))

			seen := make(map[string]bool)
			for i, part := range parts {
				displayPath := app.out.ReplaceHomeTilde(part)

				var annotations []string
				if pathCheck {
					if part == "" {
						annotations = append(annotations, app.out.DiffSprintf("empty"))
					} else {
						if seen[part] {
							annotations = append(annotations, app.out.DiffSprintf("duplicate"))
						}
						if _, err := os.Stat(part); os.IsNotExist(err) {
							annotations = append(annotations, app.out.DiffSprintf("not found"))
						}
					}
				}

				if i%2 == 1 {
					displayPath = app.out.PathSprintf(displayPath)
				}

				if len(annotations) > 0 {
					output.Printf("%s  [%s]\n", displayPath, strings.Join(annotations, ", "))
				} else {
					output.Printf("%s\n", displayPath)
				}
				seen[part] = true
			}
			if pathCheck {
				dupes := len(parts) - len(seen)
				missing := 0
				for part := range seen {
					if part != "" {
						if _, err := os.Stat(part); os.IsNotExist(err) {
							missing++
						}
					}
				}
				if dupes > 0 || missing > 0 {
					var issues []string
					if dupes > 0 {
						word := "duplicates"
						if dupes == 1 {
							word = "duplicate"
						}
						issues = append(issues, fmt.Sprintf("%d %s", dupes, word))
					}
					if missing > 0 {
						issues = append(issues, fmt.Sprintf("%d missing", missing))
					}
					output.Printf("%s %s\n", app.out.GroupSprintf("# %s —", entryCount(len(parts))), app.out.DiffSprintf(strings.Join(issues, ", ")))
				} else {
					output.Printf("%s %s\n", app.out.GroupSprintf("# %s —", entryCount(len(parts))), app.out.ProfileSprintf("all ok"))
				}
			} else {
				output.Printf("%s\n", app.out.GroupSprintf("# %s", entryCount(len(parts))))
			}
		}
	},
}

func entryCount(n int) string {
	if n == 1 {
		return "1 entry"
	}
	return fmt.Sprintf("%d entries", n)
}

func init() {
	pathCmd.Flags().BoolVarP(&pathCheck, "check", "c", false, "Check for missing directories and duplicates")
	addCommand(pathCmd)
}
