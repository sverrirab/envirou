package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sverrirab/envirou/pkg/output"
)

type shellInfo struct {
	name        string
	profilePath string
	bootstrapFn func(bool) string
}

func getBootstrapLine(shellName string, prompt bool) string {
	switch shellName {
	case "bash":
		return `eval "$(envirou bootstrap bash)"`
	case "zsh":
		return `eval "$(envirou bootstrap zsh)"`
	case "powershell":
		if prompt {
			return "Invoke-Expression (& envirou bootstrap powershell --prompt)"
		}
		return "Invoke-Expression (& envirou bootstrap powershell)"
	default:
		return ""
	}
}

func detectShell() *shellInfo {
	if runtime.GOOS == "windows" {
		profilePath := os.Getenv("USERPROFILE")
		if profilePath == "" {
			return nil
		}
		psProfile := filepath.Join(profilePath, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
		// Check if the newer PowerShell 7+ profile dir exists, otherwise fall back to WindowsPowerShell
		if _, err := os.Stat(filepath.Dir(psProfile)); os.IsNotExist(err) {
			psProfile = filepath.Join(profilePath, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
		}
		return &shellInfo{name: "powershell", profilePath: psProfile}
	}

	shellEnv := os.Getenv("SHELL")
	home := os.Getenv("HOME")
	if home == "" {
		return nil
	}

	if strings.HasSuffix(shellEnv, "/zsh") {
		return &shellInfo{name: "zsh", profilePath: filepath.Join(home, ".zshrc")}
	}
	if strings.HasSuffix(shellEnv, "/bash") {
		// macOS uses .bash_profile, Linux uses .bashrc
		if runtime.GOOS == "darwin" {
			return &shellInfo{name: "bash", profilePath: filepath.Join(home, ".bash_profile")}
		}
		return &shellInfo{name: "bash", profilePath: filepath.Join(home, ".bashrc")}
	}
	return nil
}

func fileContainsLine(path string, line string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == strings.TrimSpace(line) {
			return true, nil
		}
	}
	return false, scanner.Err()
}

func appendToFile(path string, line string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Check if file exists and if it ends with a newline
	needsNewline := false
	if info, err := os.Stat(path); err == nil && info.Size() > 0 {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		buf := make([]byte, 1)
		_, err = f.ReadAt(buf, info.Size()-1)
		f.Close()
		if err == nil && buf[0] != '\n' {
			needsNewline = true
		}
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	prefix := ""
	if needsNewline {
		prefix = "\n"
	}
	_, err = fmt.Fprintf(f, "%s%s\n", prefix, line)
	return err
}

func removeLine(path string, line string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != strings.TrimSpace(line) {
			lines = append(lines, scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644)
}

var installCmd = &cobra.Command{
	Use:   "install [bash|zsh|powershell]",
	Short: "Install ev shell function into your shell profile",
	Long: `Automatically adds the envirou bootstrap line to your shell profile file.
If no shell is specified, the current shell is auto-detected.`,
	GroupID:   "configuration",
	ValidArgs: []string{"bash", "zsh", "powershell"},
	Args:      cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var si *shellInfo
		if len(args) == 1 {
			shellName := args[0]
			if !contains(bootstrapCmd.ValidArgs, shellName) || shellName == "bat" {
				output.Printf("Unsupported shell: %s (supported: bash, zsh, powershell)\n", shellName)
				return
			}
			si = detectShell()
			if si == nil {
				si = &shellInfo{name: shellName}
			} else {
				si.name = shellName
			}
			// Override profile path for explicit shell selection
			home := os.Getenv("HOME")
			if runtime.GOOS == "windows" {
				home = os.Getenv("USERPROFILE")
			}
			switch shellName {
			case "bash":
				if runtime.GOOS == "darwin" {
					si.profilePath = filepath.Join(home, ".bash_profile")
				} else {
					si.profilePath = filepath.Join(home, ".bashrc")
				}
			case "zsh":
				si.profilePath = filepath.Join(home, ".zshrc")
			case "powershell":
				if runtime.GOOS == "windows" {
					psProfile := filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
					if _, err := os.Stat(filepath.Dir(psProfile)); os.IsNotExist(err) {
						psProfile = filepath.Join(home, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
					}
					si.profilePath = psProfile
				} else {
					si.profilePath = filepath.Join(home, ".config", "powershell", "Microsoft.PowerShell_profile.ps1")
				}
			}
		} else {
			si = detectShell()
			if si == nil {
				output.Printf("Could not detect shell. Please specify: envirou install [bash|zsh|powershell]\n")
				return
			}
		}

		bootstrapLine := getBootstrapLine(si.name, installPrompt)

		if uninstall {
			found, err := fileContainsLine(si.profilePath, bootstrapLine)
			if err != nil {
				output.Printf("Error reading %s: %v\n", si.profilePath, err)
				return
			}
			if !found {
				// Also check the prompt variant
				promptLine := getBootstrapLine(si.name, true)
				found, _ = fileContainsLine(si.profilePath, promptLine)
				if found {
					bootstrapLine = promptLine
				}
			}
			if !found {
				output.Printf("envirou bootstrap not found in %s\n", si.profilePath)
				return
			}
			if dryRun {
				output.Printf("Would remove from %s:\n  %s\n", si.profilePath, bootstrapLine)
				return
			}
			if err := removeLine(si.profilePath, bootstrapLine); err != nil {
				output.Printf("Error updating %s: %v\n", si.profilePath, err)
				return
			}
			output.Printf("Removed from %s:\n  %s\n", si.profilePath, bootstrapLine)
			output.Printf("Restart your shell for changes to take effect.\n")
			return
		}

		// Check if already installed (either variant)
		found, err := fileContainsLine(si.profilePath, bootstrapLine)
		if err != nil {
			output.Printf("Error reading %s: %v\n", si.profilePath, err)
			return
		}
		if found {
			output.Printf("Already installed in %s\n", si.profilePath)
			return
		}
		// Check for the other prompt variant
		otherLine := getBootstrapLine(si.name, !installPrompt)
		otherFound, _ := fileContainsLine(si.profilePath, otherLine)
		if otherFound {
			output.Printf("Already installed in %s (use --uninstall first to change prompt setting)\n", si.profilePath)
			return
		}

		if dryRun {
			output.Printf("Would add to %s:\n  %s\n", si.profilePath, bootstrapLine)
			return
		}

		if err := appendToFile(si.profilePath, bootstrapLine); err != nil {
			output.Printf("Error writing to %s: %v\n", si.profilePath, err)
			return
		}
		output.Printf("Added to %s:\n  %s\n", si.profilePath, bootstrapLine)
		output.Printf("Restart your shell or run the following to activate now:\n  %s\n", bootstrapLine)
	},
}

var (
	installPrompt bool
	uninstall     bool
)

func init() {
	addCommand(installCmd)
	installCmd.Flags().BoolVarP(&installPrompt, "prompt", "p", false, "Also install prompt customization (PowerShell only)")
	installCmd.Flags().BoolVar(&uninstall, "uninstall", false, "Remove envirou from shell profile")
}
