package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetBootstrapLine(t *testing.T) {
	tests := []struct {
		shell    string
		prompt   bool
		contains string
	}{
		{"bash", false, `eval "$(envirou bootstrap bash)"`},
		{"zsh", false, `eval "$(envirou bootstrap zsh)"`},
		{"powershell", false, "Invoke-Expression (& envirou bootstrap powershell)"},
		{"powershell", true, "Invoke-Expression (& envirou bootstrap powershell --prompt)"},
		{"unknown", false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			got := getBootstrapLine(tt.shell, tt.prompt)
			if got != tt.contains {
				t.Errorf("getBootstrapLine(%q, %v) = %q, want %q", tt.shell, tt.prompt, got, tt.contains)
			}
		})
	}
}

func TestFileContainsLine(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "install-test")
	if err != nil {
		t.Fatal(err)
	}
	name := tmpFile.Name()
	t.Cleanup(func() { os.Remove(name) })

	tmpFile.WriteString("first line\nsecond line\nthird line\n")
	tmpFile.Close()

	found, err := fileContainsLine(name, "second line")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Error("expected to find 'second line'")
	}

	found, err = fileContainsLine(name, "missing line")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("should not find 'missing line'")
	}

	// Test with whitespace trimming
	found, err = fileContainsLine(name, "  second line  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Error("expected to find with whitespace trimming")
	}
}

func TestFileContainsLineNoFile(t *testing.T) {
	found, err := fileContainsLine("/nonexistent/path/file", "anything")
	if err != nil {
		t.Fatalf("should not error for missing file: %v", err)
	}
	if found {
		t.Error("should not find in missing file")
	}
}

func TestAppendToFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "profile")

	// Append to new file
	err := appendToFile(path, "line one")
	if err != nil {
		t.Fatalf("appendToFile failed: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "line one\n" {
		t.Errorf("expected 'line one\\n', got %q", content)
	}

	// Append to existing file (already ends with newline)
	err = appendToFile(path, "line two")
	if err != nil {
		t.Fatalf("appendToFile failed: %v", err)
	}

	content, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "line one\nline two\n" {
		t.Errorf("expected two lines, got %q", content)
	}
}

func TestAppendToFileNoTrailingNewline(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "profile")

	// Write file without trailing newline
	os.WriteFile(path, []byte("existing"), 0644)

	err := appendToFile(path, "new line")
	if err != nil {
		t.Fatalf("appendToFile failed: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "existing\nnew line\n" {
		t.Errorf("expected newline inserted, got %q", content)
	}
}

func TestAppendToFileCreatesDir(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "subdir", "profile")

	err := appendToFile(path, "hello")
	if err != nil {
		t.Fatalf("appendToFile should create parent dirs: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "hello\n" {
		t.Errorf("expected 'hello\\n', got %q", content)
	}
}

func TestRemoveLine(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "install-test")
	if err != nil {
		t.Fatal(err)
	}
	name := tmpFile.Name()
	t.Cleanup(func() { os.Remove(name) })

	tmpFile.WriteString("keep this\nremove this\nalso keep\n")
	tmpFile.Close()

	err = removeLine(name, "remove this")
	if err != nil {
		t.Fatalf("removeLine failed: %v", err)
	}

	content, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "keep this\nalso keep\n" {
		t.Errorf("expected line removed, got %q", content)
	}
}

func TestRemoveLineWithWhitespace(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "install-test")
	if err != nil {
		t.Fatal(err)
	}
	name := tmpFile.Name()
	t.Cleanup(func() { os.Remove(name) })

	tmpFile.WriteString("keep\n  target  \nend\n")
	tmpFile.Close()

	err = removeLine(name, "target")
	if err != nil {
		t.Fatalf("removeLine failed: %v", err)
	}

	content, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "keep\nend\n" {
		t.Errorf("expected whitespace-trimmed removal, got %q", content)
	}
}

func TestInstallDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("SHELL", "/bin/bash")
	_ = executeCommand(t, "install", "--dry-run")
}

func TestInstallExplicitShellDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("SHELL", "/bin/bash")
	_ = executeCommand(t, "install", "bash", "--dry-run")
	_ = executeCommand(t, "install", "zsh", "--dry-run")
}

func TestInstallUninstallNotInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("SHELL", "/bin/bash")
	_ = executeCommand(t, "install", "--uninstall")
}

func TestCollapseToOneLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"multiline", "line one\nline two\nline three", "line one; line two; line three"},
		{"with blanks", "a\n\nb\n\nc", "a; b; c"},
		{"single", "only", "only"},
		{"empty", "", ""},
		{"whitespace lines", "  a  \n  \n  b  ", "a; b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collapseToOneLine(tt.input)
			if got != tt.expected {
				t.Errorf("collapseToOneLine(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
