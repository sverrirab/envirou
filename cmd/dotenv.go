package cmd

import (
	"bufio"
	"github.com/spf13/cobra"
	"github.com/sverrirab/envirou/pkg/output"
	"os"
	"strings"
)

var dotenvCmd = &cobra.Command{
	Use:     ".env",
	Short:   "Read .env file from current directory",
	GroupID: "profiles",
	Run: func(cmd *cobra.Command, args []string) {
		newEnv := app.baseEnv.Clone()
		file, err := os.Open(dotenvFile)
		if err != nil {
			output.Printf("Local .env file not found\n")
			os.Exit(1)
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
			output.Printf("Failed reading from .env file (%s)\n", err.Error())
			os.Exit(1)
		}
		newEnv.MergeStrings(lines)
		app.shellCommands = append(app.shellCommands, app.sh.GetCommands(app.baseEnv, newEnv)...)
	},
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

var dotenvFile = ".env"

func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		output.Printf("Failed closing file (%s)\n", err.Error())
	}
}

func init() {
	addCommand(dotenvCmd)
	dotenvCmd.Flags().StringVarP(
		&dotenvFile, "file", "f", dotenvFile,
		"file name to read (default .env)")
}
