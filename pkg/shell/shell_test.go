package shell

import (
	"reflect"
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

func TestLocalVar(t *testing.T) {
	sh := NewShell(false, false)
	if got, want := sh.LocalVar("FOO", "a b"), "unset FOO;FOO='a b'"; got != want {
		t.Errorf("posix LocalVar = %q, want %q", got, want)
	}
	ps := NewShell(true, false)
	if got, want := ps.LocalVar("FOO", "a b"), "Remove-Item Env:FOO -ErrorAction SilentlyContinue;$global:FOO = 'a b'"; got != want {
		t.Errorf("powershell LocalVar = %q, want %q", got, want)
	}
	bat := NewShell(false, true)
	if got, want := bat.LocalVar("FOO", "a b"), "set FOO='a b'"; got != want {
		t.Errorf("bat LocalVar = %q, want %q", got, want)
	}
}

func TestLocalVarRejectsInvalidName(t *testing.T) {
	for _, ps := range []bool{false, true} {
		sh := NewShell(ps, false)
		if cmd := sh.LocalVar("BAD;rm -rf /", "value"); cmd != "" {
			t.Errorf("expected empty for injected name, got %q", cmd)
		}
		if cmd := sh.UnsetLocalVar("$(evil)"); cmd != "" {
			t.Errorf("expected empty for subshell name, got %q", cmd)
		}
	}
}

func TestUnsetLocalVar(t *testing.T) {
	sh := NewShell(false, false)
	if got, want := sh.UnsetLocalVar("FOO"), "unset FOO"; got != want {
		t.Errorf("posix UnsetLocalVar = %q, want %q", got, want)
	}
	ps := NewShell(true, false)
	want := "Remove-Item Env:FOO -ErrorAction SilentlyContinue;Remove-Variable -Name FOO -Scope Global -ErrorAction SilentlyContinue"
	if got := ps.UnsetLocalVar("FOO"); got != want {
		t.Errorf("powershell UnsetLocalVar = %q, want %q", got, want)
	}
	bat := NewShell(false, true)
	if got, want := bat.UnsetLocalVar("FOO"), "set FOO="; got != want {
		t.Errorf("bat UnsetLocalVar = %q, want %q", got, want)
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

func TestRedactCommandValues(t *testing.T) {
	tests := []struct {
		name     string
		shell    *Shell
		commands []string
		want     []string
	}{
		{
			name:     "bash",
			shell:    NewShell(false, false),
			commands: []string{"export ENVIROU_KEY=secret", "export EMPTY=", "unset OLD"},
			want:     []string{"export ENVIROU_KEY=<redacted>", "export EMPTY=<redacted>", "unset OLD"},
		},
		{
			name:     "bash local assignment",
			shell:    NewShell(false, false),
			commands: []string{"unset ENVIROU_KEY;ENVIROU_KEY='secret'", "FOO=bar", "echo hi=there", "unset notaname!;X=1"},
			want:     []string{"unset ENVIROU_KEY;ENVIROU_KEY=<redacted>", "FOO=<redacted>", "echo hi=there", "unset notaname!;X=1"},
		},
		{
			name:     "powershell",
			shell:    NewShell(true, false),
			commands: []string{"$Env:SECRET = 'value'", "Remove-Item Env:OLD"},
			want:     []string{"$Env:SECRET = <redacted>", "Remove-Item Env:OLD"},
		},
		{
			name:     "powershell local assignment",
			shell:    NewShell(true, false),
			commands: []string{"Remove-Item Env:ENVIROU_KEY -ErrorAction SilentlyContinue;$global:ENVIROU_KEY = 'secret'"},
			want:     []string{"Remove-Item Env:ENVIROU_KEY -ErrorAction SilentlyContinue;$global:ENVIROU_KEY = <redacted>"},
		},
		{
			name:     "bat",
			shell:    NewShell(false, true),
			commands: []string{"set SECRET=value", "set OLD="},
			want:     []string{"set SECRET=<redacted>", "set OLD="},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.shell.RedactCommandValues(tt.commands)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("RedactCommandValues() = %q, want %q", got, tt.want)
			}
		})
	}
}
