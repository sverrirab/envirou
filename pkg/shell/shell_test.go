package shell

import (
	"testing"

	"github.com/sverrirab/envirou/pkg/data"
)

func validateEscaped(t *testing.T, s string) {
	if s == escape(s) {
		t.Errorf("Should be escaped %s == %s", s, escape(s))
	}
}

func validateExact(t *testing.T, original, expected string) {
	if expected != escape(original) {
		t.Errorf("Incorrect escape of %s:\n  EXPECT: %s.\n  ACTUAL: %s.\n", original, expected, escape(original))
	}
}

func validateUnEscaped(t *testing.T, s string) {
	if s != escape(s) {
		t.Errorf("Should not be escaped  %s == %s", s, escape(s))
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