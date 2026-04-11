package shell

import (
	"testing"
)

func TestEscapePowerShell(t *testing.T) {
	sh := NewShell(true, false)

	// PowerShell always wraps in single quotes
	if got := sh.Escape("hello"); got != "'hello'" {
		t.Errorf("expected 'hello', got %q", got)
	}

	// PowerShell doubles single quotes for escaping
	if got := sh.Escape("it's"); got != "'it''s'" {
		t.Errorf("expected 'it''s', got %q", got)
	}

	// Simple value
	if got := sh.Escape("simple"); got != "'simple'" {
		t.Errorf("expected 'simple', got %q", got)
	}
}

func TestExportVarPowerShell(t *testing.T) {
	sh := NewShell(true, false)
	got := sh.ExportVar("FOO", "bar")
	if got != "$Env:FOO = 'bar'" {
		t.Errorf("expected PowerShell export, got %q", got)
	}
}

func TestExportVarBat(t *testing.T) {
	sh := NewShell(false, true)
	got := sh.ExportVar("FOO", "bar")
	if got != "set FOO=bar" {
		t.Errorf("expected bat export, got %q", got)
	}
}

func TestExportVarBash(t *testing.T) {
	sh := NewShell(false, false)
	got := sh.ExportVar("FOO", "bar")
	if got != "export FOO=bar" {
		t.Errorf("expected bash export, got %q", got)
	}
}

func TestUnsetVarPowerShell(t *testing.T) {
	sh := NewShell(true, false)
	got := sh.UnsetVar("FOO")
	if got != "Remove-Item Env:FOO" {
		t.Errorf("expected PowerShell unset, got %q", got)
	}
}

func TestUnsetVarBat(t *testing.T) {
	sh := NewShell(false, true)
	got := sh.UnsetVar("FOO")
	if got != "set FOO=" {
		t.Errorf("expected bat unset, got %q", got)
	}
}

func TestUnsetVarBash(t *testing.T) {
	sh := NewShell(false, false)
	got := sh.UnsetVar("FOO")
	if got != "unset FOO" {
		t.Errorf("expected bash unset, got %q", got)
	}
}

func TestExportVarWithSpacesPowerShell(t *testing.T) {
	sh := NewShell(true, false)
	got := sh.ExportVar("FOO", "hello world")
	if got != "$Env:FOO = 'hello world'" {
		t.Errorf("expected PowerShell export with spaces, got %q", got)
	}
}

func TestExportVarWithSpacesBat(t *testing.T) {
	sh := NewShell(false, true)
	got := sh.ExportVar("FOO", "hello world")
	if got != "set FOO='hello world'" {
		t.Errorf("expected bat export with escaped value, got %q", got)
	}
}

func TestRunCommandsPowerShell(t *testing.T) {
	sh := NewShell(true, false)

	// Empty commands
	if got := sh.RunCommands(nil); got != "" {
		t.Errorf("expected empty for nil commands, got %q", got)
	}

	// PowerShell uses ; separator (same as unix path)
	got := sh.RunCommands([]string{"cmd1", "cmd2"})
	if got != "cmd1;cmd2\n" {
		t.Errorf("expected semicolon-separated, got %q", got)
	}
}

func TestNeedsEscapeEdgeCases(t *testing.T) {
	sh := NewShell(false, false)

	// Empty string should not be escaped
	if got := sh.Escape(""); got != "" {
		t.Errorf("empty string should pass through, got %q", got)
	}

	// All safe characters
	if got := sh.Escape("a-z_A-Z.0/9:;+?,-!#=*"); got != "a-z_A-Z.0/9:;+?,-!#=*" {
		t.Errorf("safe chars should not be escaped, got %q", got)
	}

	// Dollar sign needs escape
	got := sh.Escape("$HOME")
	if got == "$HOME" {
		t.Error("$ should trigger escaping")
	}

	// Backtick needs escape
	got = sh.Escape("back`tick")
	if got == "back`tick" {
		t.Error("backtick should trigger escaping")
	}
}
