package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sverrirab/envirou/pkg/data"
	"github.com/sverrirab/envirou/pkg/output"
)

var dotenvCmd = &cobra.Command{
	Use:     "dotenv [files...]",
	Aliases: []string{".env"},
	Short:   "Load environment from .env files",
	Long: `Read one or more .env files and apply the variables to the current environment.
If no files are specified, reads .env from the current directory.
When multiple files are given, they are loaded in order and later values override earlier ones.`,
	GroupID: "profiles",
	Run: func(cmd *cobra.Command, args []string) {
		files := args
		if len(files) == 0 {
			files = []string{".env"}
		}

		newEnv := app.baseEnv.Clone()
		for _, filename := range files {
			if err := loadDotenvFile(filename, newEnv); err != nil {
				output.Printf("%s: %s\n", filename, err.Error())
				os.Exit(1)
			}
		}
		app.shellCommands = append(app.shellCommands, app.sh.GetCommands(app.baseEnv, newEnv)...)
	},
}

func loadDotenvFile(filename string, env *data.Profile) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("file not found")
	}
	defer closeFile(file)

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := parseDotenvLine(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed reading file (%s)", err.Error())
	}
	env.MergeStrings(lines)
	return nil
}

// parseDotenvLine parses a single .env line, returning a clean KEY=value string.
// Returns empty string for lines that should be skipped (comments, blank lines).
func parseDotenvLine(line string) string {
	line = strings.TrimSpace(line)

	// Skip empty lines and comments
	if line == "" || strings.HasPrefix(line, "#") {
		return ""
	}

	// Strip optional "export " prefix
	line = strings.TrimPrefix(line, "export ")

	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return ""
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	// Strip matching quotes
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'') {
			value = value[1 : len(value)-1]
		}
	}

	return key + "=" + value
}

func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		output.Printf("Failed closing file (%s)\n", err.Error())
	}
}

func init() {
	addCommand(dotenvCmd)
}
