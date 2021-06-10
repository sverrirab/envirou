package shell

import (
	"testing"

	"github.com/sverrirab/envirou/pkg/data"
)

func validateEscaped(t *testing.T, s string) {
	if s == Escape(s) {
		t.Errorf("Should be escaped %s == %s", s, Escape(s))
	}
}

func validateExact(t *testing.T, original, expected string) {
	if expected != Escape(original) {
		t.Errorf("Incorrect escape of %s:\n  EXPECT: %s.\n  ACTUAL: %s.\n", original, expected, Escape(original))
	}
}

func validateUnEscaped(t *testing.T, s string) {
	if s != Escape(s) {
		t.Errorf("Should not be escaped  %s == %s", s, Escape(s))
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

func TestCommands(t *testing.T) {
	e1 := []string{"FOO=2", "BAR=FOO=FOOBAR", "SMURF=", "REMOVE"}
	before := data.NewProfile()
	before.MergeStrings(e1)
	e2 := []string{"SMURF=yes yes", "BOAT", "FOO"}
	after := data.NewProfile()
	after.MergeStrings(e2)

	commands := GetCommands(before, after)
	expected := "foobar"
	if len(commands) != 2 || commands[0] != "export SMURF='yes yes'" || commands[1] != "unset FOO" {
		t.Errorf("Invalid commands:\n  EXPECT: %s.\n  ACTUAL: %s\n", expected, commands)
	}
}

func TestRunCommands(t *testing.T) {
	cmd1 := RunCommands([]string{})
	if cmd1 != "" {
		t.Errorf("Did not expect no command to be: %s.", cmd1)
	}
	cmd2 := RunCommands([]string{"echo hi", "ls -al"})
	if cmd2 != "echo hi;ls -al;\n" {
		t.Errorf("Did not expect commands to be: %s.", cmd2)
	}
}