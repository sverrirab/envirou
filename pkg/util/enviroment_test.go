package util

import (
	"testing"
)

func TestParseEnvironment(t *testing.T) {
	e := [...]string{"FOO=2", "BAR=FOO=FOOBAR", "SMURF="}
	env, keys := ParseEnvironment(e[:])
	if env["FOO"] != "2" {
		t.Errorf("Unexpected value %s", env["FOO"])
	}
	if env["BAR"] != "FOO=FOOBAR" {
		t.Errorf("Unexpected value %s", env["BAR"])
	}
	if env["SMURF"] != "" {
		t.Errorf("Unexpected value %s", env["SMURF"])
	}
	if keys[0] != "BAR" {
		t.Error("Sorting BAR failed?")
	}
	if keys[1] != "FOO" {
		t.Error("Sorting FOO failed?")
	}
	if keys[2] != "SMURF" {
		t.Error("Sorting SMURF failed?")
	}
}
