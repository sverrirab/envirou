package shell

import (
	"testing"

	"github.com/sverrirab/envirou/pkg/data"
)

func validateEscaped(t *testing.T, s string) {
	sh := NewShell(false, false)
	if s == sh.Escape(s) {
		t.Errorf("Should be escaped %s == %s", s, sh.Escape(s))
	}
}

func validateExact(t *testing.T, original, expected string) {
	sh := NewShell(false, false)
	if expected != sh.Escape(original) {
		t.Errorf("Incorrect escape of %s:\n  EXPECT: %s.\n  ACTUAL: %s.\n", original, expected, sh.Escape(original))
	}
}

func validateUnEscaped(t *testing.T, s string) {
	sh := NewShell(false, false)
	if s != sh.Escape(s) {
		t.Errorf("Should not be escaped  %s == %s", s, sh.Escape(s))
	}
}

func TestIsValidVarName(t *testing.T) {
	valid := []string{"FOO", "BAR_BAZ", "_private", "a", "A1", "PATH", "__", "_0"}
	for _, name := range valid {
		if !IsValidVarName(name) {
			t.Errorf("expected %q to be valid", name)
		}
	}
	invalid := []string{
		"",
		"0FOO",
		"FOO BAR",
		"FOO;BAR",
		"FOO;rm -rf ~",
		"$(cmd)",
		"`cmd`",
		"FOO=BAR",
		"FOO\nBAR",
		"a-b",
		"a.b",
		"FOO$HOME",
	}
	for _, name := range invalid {
		if IsValidVarName(name) {
			t.Errorf("expected %q to be invalid", name)
		}
	}
}

func TestExportVarRejectsInvalidName(t *testing.T) {
	for _, ps := range []bool{false, true} {
		sh := NewShell(ps, false)
		if cmd := sh.ExportVar("VALID", "value"); cmd == "" {
			t.Error("expected command for valid name")
		}
		if cmd := sh.ExportVar("BAD;rm -rf /", "value"); cmd != "" {
			t.Errorf("expected empty for injected name, got %q", cmd)
		}
		if cmd := sh.ExportVar("$(evil)", "value"); cmd != "" {
			t.Errorf("expected empty for subshell name, got %q", cmd)
		}
	}
}

func TestUnsetVarRejectsInvalidName(t *testing.T) {
	for _, ps := range []bool{false, true} {
		sh := NewShell(ps, false)
		if cmd := sh.UnsetVar("VALID"); cmd == "" {
			t.Error("expected command for valid name")
		}
		if cmd := sh.UnsetVar("BAD;rm -rf /"); cmd != "" {
			t.Errorf("expected empty for injected name, got %q", cmd)
		}
	}
}

func TestEscape(t *testing.T) {
	validateEscaped(t, "hello world")
	validateEscaped(t, "$hi")
	validateEscaped(t, "what~up")
	validateEscaped(t, "what`up")

	validateExact(t, "hi", "hi")
	validateExact(t, "'why don't you do that to me?'", "''\\''why don'\\''t you do that to me?'\\'''")

	validateUnEscaped(t, "hello")
	validateUnEscaped(t, "He.l0981o-World!")
	validateUnEscaped(t, "How*Tri+cks?")
	validateUnEscaped(t, "/:;_,-!#=*")
}

func TestCommandsBash(t *testing.T) {
	e1 := []string{"FOO=2", "BAR=FOO=FOOBAR", "SMURF=", "REMOVE"}
	before := data.NewProfile(false)
	before.MergeStrings(e1)
	e2 := []string{"SMURF=yes yes", "BOAT", "FOO"}
	after := data.NewProfile(false)
	after.MergeStrings(e2)

	sh := NewShell(false, false)
	commands := sh.GetCommands(before, after)
	expected := "foobar"
	if len(commands) != 2 || commands[0] != "export SMURF='yes yes'" || commands[1] != "unset FOO" {
		t.Errorf("Invalid commands:\n  EXPECT: %s.\n  ACTUAL: %s\n", expected, commands)
	}
}

func TestCommandsBat(t *testing.T) {
	e1 := []string{"FOO=2", "BAR=FOO=FOOBAR", "SMURF=", "REMOVE"}
	before := data.NewProfile(false)
	before.MergeStrings(e1)
	e2 := []string{"SMURF=yes yes", "BOAT", "FOO"}
	after := data.NewProfile(false)
	after.MergeStrings(e2)

	sh := NewShell(false, true)
	// shellBash := NewShell(false, false)
	commands := sh.GetCommands(before, after)
	expected := "foobar"
	if len(commands) != 2 || commands[0] != "set SMURF='yes yes'" || commands[1] != "set FOO=" {
		t.Errorf("Invalid commands:\n  EXPECT: %s.\n  ACTUAL: %s\n", expected, commands)
	}
}

func TestRunCommandsBash(t *testing.T) {
	sh := NewShell(false, false)
	cmd1 := sh.RunCommands([]string{})
	if cmd1 != "" {
		t.Errorf("Did not expect no command to be: %s.", cmd1)
	}
	cmd2 := sh.RunCommands([]string{"echo hi", "ls -al"})
	if cmd2 != "echo hi;ls -al\n" {
		t.Errorf("Did not expect commands to be: %s.", cmd2)
	}
}

func TestRunCommandsBat(t *testing.T) {
	sh := NewShell(false, true)
	cmd1 := sh.RunCommands([]string{})
	if cmd1 != "" {
		t.Errorf("Did not expect no command to be: %s.", cmd1)
	}
	cmd2 := sh.RunCommands([]string{"echo hi", "ls -al"})
	if cmd2 != "echo hi & ls -al\n" {
		t.Errorf("Did not expect commands to be: %s.", cmd2)
	}
}
