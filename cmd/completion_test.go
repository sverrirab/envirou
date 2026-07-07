package cmd

import (
	"strings"
	"testing"
)

// Completion scripts must arrive on stdout so that
// "envirou completion SHELL > file" works.
func TestCompletionZsh(t *testing.T) {
	out := executeCommand(t, "completion", "zsh")
	if !strings.HasPrefix(out, "#compdef envirou") {
		t.Errorf("Expected zsh completion script on stdout, got: %.80s", out)
	}
	if !strings.Contains(out, "_envirou()") {
		t.Error("Expected _envirou function in zsh completion script")
	}
}

func TestCompletionBash(t *testing.T) {
	out := executeCommand(t, "completion", "bash")
	if !strings.Contains(out, "__start_envirou") {
		t.Errorf("Expected bash completion script on stdout, got: %.80s", out)
	}
}

func TestCompletionFish(t *testing.T) {
	out := executeCommand(t, "completion", "fish")
	if !strings.Contains(out, "complete -c envirou") {
		t.Errorf("Expected fish completion script on stdout, got: %.80s", out)
	}
}

func TestCompletionPowershell(t *testing.T) {
	out := executeCommand(t, "completion", "powershell")
	if !strings.Contains(out, "Register-ArgumentCompleter") {
		t.Errorf("Expected PowerShell completion script on stdout, got: %.80s", out)
	}
}

func TestCompletionInvalidArg(t *testing.T) {
	resetState(t)
	rootCmd.SetArgs([]string{"completion", "fish-food"})
	if err := rootCmd.Execute(); err == nil {
		t.Error("Expected error for invalid shell type")
	}
}

func TestCompletionNoArg(t *testing.T) {
	resetState(t)
	rootCmd.SetArgs([]string{"completion"})
	if err := rootCmd.Execute(); err == nil {
		t.Error("Expected error when no shell type provided")
	}
}
