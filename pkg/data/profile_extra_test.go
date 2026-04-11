package data

import (
	"testing"
)

func TestProfileString(t *testing.T) {
	p := NewProfile(false)
	p.Set("B", "2")
	p.Set("A", "1")
	p.SetNil("C")

	s := p.String()
	if s != "A,B,C" {
		t.Errorf("expected 'A,B,C', got %q", s)
	}
}

func TestProfileStringEmpty(t *testing.T) {
	p := NewProfile(false)
	s := p.String()
	if s != "" {
		t.Errorf("expected empty string, got %q", s)
	}
}

func TestGetMergeModeDefault(t *testing.T) {
	p := NewProfile(false)
	p.Set("FOO", "bar")

	if mode := p.GetMergeMode("FOO"); mode != MergeReplace {
		t.Errorf("expected MergeReplace, got %d", mode)
	}
	// Missing key should also return MergeReplace
	if mode := p.GetMergeMode("MISSING"); mode != MergeReplace {
		t.Errorf("expected MergeReplace for missing key, got %d", mode)
	}
}

func TestGetMergeModeExplicit(t *testing.T) {
	p := NewProfile(false)
	p.SetWithMode("PATH", "/new", MergePrepend)
	p.SetWithMode("LIB", "/lib", MergeAppend)

	if mode := p.GetMergeMode("PATH"); mode != MergePrepend {
		t.Errorf("expected MergePrepend, got %d", mode)
	}
	if mode := p.GetMergeMode("LIB"); mode != MergeAppend {
		t.Errorf("expected MergeAppend, got %d", mode)
	}
}

func TestIsMergedNilBranch(t *testing.T) {
	env := NewProfile(false)
	env.Set("FOO", "bar")

	profile := NewProfile(false)
	profile.SetNil("FOO")

	// FOO exists in env but profile wants it nil — not merged
	if env.IsMerged(profile) {
		t.Error("should not be merged: FOO exists but profile wants it nil")
	}

	// After removing FOO, it should be merged
	env.SetNil("FOO")
	if !env.IsMerged(profile) {
		t.Error("should be merged: FOO is nil in both")
	}
}

func TestIsMergedMissingVar(t *testing.T) {
	env := NewProfile(false)
	// env has no variables

	profile := NewProfile(false)
	profile.Set("FOO", "bar")

	if env.IsMerged(profile) {
		t.Error("should not be merged: FOO missing from env")
	}
}

func TestClonePreservesMergeMode(t *testing.T) {
	p := NewProfile(false)
	p.SetWithMode("PATH", "/a", MergePrepend)
	p.SetWithMode("LIB", "/b", MergeAppend)
	p.Set("FOO", "bar")
	p.SetNil("GONE")

	c := p.Clone()

	if mode := c.GetMergeMode("PATH"); mode != MergePrepend {
		t.Errorf("clone should preserve MergePrepend, got %d", mode)
	}
	if mode := c.GetMergeMode("LIB"); mode != MergeAppend {
		t.Errorf("clone should preserve MergeAppend, got %d", mode)
	}
	if mode := c.GetMergeMode("FOO"); mode != MergeReplace {
		t.Errorf("clone should preserve MergeReplace, got %d", mode)
	}
	if !c.GetNil("GONE") {
		t.Error("clone should preserve nil entries")
	}
}

func TestCloneCaseInsensitive(t *testing.T) {
	p := NewProfile(true)
	p.Set("Hello", "world")

	c := p.Clone()
	verifyValue(t, c, "HELLO", "world")
	verifyValue(t, c, "hello", "world")
}

func TestMergeResultPathSkipped(t *testing.T) {
	env := NewProfile(false)
	env.Set("PATH", p("/a", "/b"))

	profile := NewProfile(false)
	profile.SetWithMode("PATH", "/a", MergePrepend)

	result := env.Merge(profile)
	if len(result.PathSkipped) != 1 || result.PathSkipped[0] != "PATH" {
		t.Errorf("expected PathSkipped=[PATH], got %v", result.PathSkipped)
	}
}

func TestMergeResultNoSkip(t *testing.T) {
	env := NewProfile(false)
	env.Set("PATH", p("/a", "/b"))

	profile := NewProfile(false)
	profile.SetWithMode("PATH", "/c", MergePrepend)

	result := env.Merge(profile)
	if len(result.PathSkipped) != 0 {
		t.Errorf("expected no PathSkipped, got %v", result.PathSkipped)
	}
}
