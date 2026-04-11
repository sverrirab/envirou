package data

import (
	"testing"
)

func TestIsIgnored(t *testing.T) {
	g := NewGroups()
	g.ParseAndAdd("visible", "FOO, BAR*", false)
	g.ParseAndAdd("..ignore", "PWD, OLDPWD, SHLVL", false)
	g.ParseAndAdd("..transient", "_", false)

	if !g.IsIgnored("PWD", false) {
		t.Error("PWD should be ignored")
	}
	if !g.IsIgnored("SHLVL", false) {
		t.Error("SHLVL should be ignored")
	}
	if !g.IsIgnored("_", false) {
		t.Error("_ should be ignored")
	}
	if g.IsIgnored("FOO", false) {
		t.Error("FOO should not be ignored")
	}
	if g.IsIgnored("RANDOM_VAR", false) {
		t.Error("RANDOM_VAR should not be ignored")
	}
}

func TestIsIgnoredCaseInsensitive(t *testing.T) {
	g := NewGroups()
	g.ParseAndAdd("..ignore", "PWD", true)

	if !g.IsIgnored("pwd", true) {
		t.Error("pwd should be ignored case-insensitively")
	}
}

func TestIsIgnoredNoDoublePrefix(t *testing.T) {
	g := NewGroups()
	g.ParseAndAdd(".hidden", "SECRET*", false)

	// Single-dot prefix is hidden, not ignored
	if g.IsIgnored("SECRET_KEY", false) {
		t.Error(".hidden group should not count as ignored (needs .. prefix)")
	}
}

func TestGroupNameToEnvsGetAllNames(t *testing.T) {
	g := NewGroups()
	g.ParseAndAdd("beta", "B*", false)
	g.ParseAndAdd("alpha", "A*", false)

	envs := Envs{"ALPHA_ONE", "BETA_ONE", "CHARLIE"}
	matched, _ := g.MatchAll(envs, false)

	names := matched.GetAllNames()
	if len(names) != 2 {
		t.Fatalf("expected 2 group names, got %d", len(names))
	}
	if names[0] != "alpha" || names[1] != "beta" {
		t.Errorf("expected [alpha, beta], got %v", names)
	}
}

func TestGroupNameToEnvsGetAllNamesEmpty(t *testing.T) {
	m := make(GroupNameToEnvs)
	names := m.GetAllNames()
	if len(names) != 0 {
		t.Errorf("expected 0 names, got %d", len(names))
	}
}
