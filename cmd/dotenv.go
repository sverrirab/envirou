package cmd

import (
	"bufio"
	"github.com/spf13/cobra"
	"github.com/sverrirab/envirou/pkg/output"
	"os"
)

// setCmd represents the set command
var dotenvCmd = &cobra.Command{
	Use:   ".env",
	Short: "Read .env file from current directory",
	Run: func(cmd *cobra.Command, args []string) {
		newEnv := baseEnv.Clone()
		file, err := os.Open(dotenvFile)
		if err != nil {
			output.Printf("Local .env file not found\n")
			os.Exit(1)
		}
		defer closeFile(file)

		var lines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			output.Printf("Failed reading from .env file (%s)\n", err.Error())
			os.Exit(1)
		}
		newEnv.MergeStrings(lines)
		shellCommands = append(shellCommands, sh.GetCommands(baseEnv, newEnv)...)
	},
}

var (
	dotenvFile string = ".env"
)

func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		output.Printf("Failed closing file (%s)\n", err.Error())
	}
}

func init() {
	addCommand(dotenvCmd)
	rootCmd.PersistentFlags().StringVarP(
		&dotenvFile, "file", "f", dotenvFile,
		"file name to read (default .env)")
}
