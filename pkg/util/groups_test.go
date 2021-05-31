package util

import (
	"testing"
)

func TestGroups(t *testing.T) {
	g := NewGroups()

	if len(g.GetAllNames()) != 0 {
		t.Error("len should be 0")
	}
	g.ParseAndAdd("smurf", "one, two*")
	g.ParseAndAdd("foobar", "*FOO*")
	names := g.GetAllNames()	
	if len(names) != 2 {
		t.Error("len should be 2")
	}
	if names[0] != "foobar" {
		t.Error("incorrect order")
	}
	_, found := g.GetPatterns("one")
	if found {
		t.Error("should not find one")
	}

	patterns, found := g.GetPatterns("smurf")
	if !found {
		t.Error("should find smurf")
	}
	if ! MatchAny("twofold", patterns) {
		t.Error("pattern not correct")
	}
}
