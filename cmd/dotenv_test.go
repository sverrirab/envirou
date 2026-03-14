package cmd

import "testing"

func TestParseDotenvLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "FOO=bar", "FOO=bar"},
		{"with spaces in value", "FOO=hello world", "FOO=hello world"},
		{"empty value", "FOO=", "FOO="},
		{"value with equals", "FOO=bar=baz", "FOO=bar=baz"},

		// Quoting
		{"double quoted", `FOO="bar"`, "FOO=bar"},
		{"single quoted", "FOO='bar'", "FOO=bar"},
		{"quoted with spaces", `FOO="hello world"`, "FOO=hello world"},
		{"mismatched quotes kept", `FOO="bar'`, `FOO="bar'`},
		{"single char value not stripped", `FOO="`, `FOO="`},

		// Comments and blank lines
		{"comment", "# this is a comment", ""},
		{"comment with leading space", "  # comment", ""},
		{"empty line", "", ""},
		{"whitespace only", "   ", ""},

		// Export prefix
		{"export prefix", "export FOO=bar", "FOO=bar"},
		{"export quoted", `export FOO="bar"`, "FOO=bar"},

		// Whitespace handling
		{"leading whitespace", "  FOO=bar", "FOO=bar"},
		{"trailing whitespace", "FOO=bar  ", "FOO=bar"},
		{"spaces around equals", " FOO = bar ", "FOO=bar"},

		// Lines without equals
		{"no equals", "FOOBAR", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDotenvLine(tt.input)
			if got != tt.expected {
				t.Errorf("parseDotenvLine(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestRemoveFirstLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"with shebang", "#!/bin/bash\necho hello", "echo hello"},
		{"no newline", "single line", "single line"},
		{"empty first line", "\nrest", "rest"},
		{"multiple lines", "first\nsecond\nthird", "second\nthird"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removeFirstLine(tt.input)
			if got != tt.expected {
				t.Errorf("removeFirstLine(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestContains(t *testing.T) {
	slice := []string{"bash", "zsh", "powershell", "bat"}
	if !contains(slice, "bash") {
		t.Error("should find bash")
	}
	if !contains(slice, "powershell") {
		t.Error("should find powershell")
	}
	if contains(slice, "fish") {
		t.Error("should not find fish")
	}
	if contains(nil, "bash") {
		t.Error("should not find in nil slice")
	}
}
