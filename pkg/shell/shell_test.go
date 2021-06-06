package shell

import (
	"testing"
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