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

func TestMatchAll(t *testing.T) {
	g := NewGroups()
	g.ParseAndAdd("zero", "SMURF, DURF")
	g.ParseAndAdd("foo", "*FOO*")
	g.ParseAndAdd("bar", "FOO*, SMURF")
	g.ParseAndAdd("joe", "whatever*")

	m, r := g.MatchAll([]string{"SMURF", "FOOBAR", "bob"})
	if len(m) != 3 {
		t.Error("Expected three matching groups")
	}
	if len(m["bar"]) != 2 {
		t.Errorf("Expected bar to contain two items: %s", m["bar"])
	}
	if len(r) != 1 || r[0] != "bob" {
		t.Error("Where is bob?")
	}
}
