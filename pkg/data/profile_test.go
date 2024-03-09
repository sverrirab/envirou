package data

import (
	"testing"
)

func verifyValue(t *testing.T, p *Profile, name string, value string) {
	if v, ok := p.Get(name); !ok || value != v {
		t.Errorf("Unexpected value %s != %s", value, v)
	}
}

func verifyNil(t *testing.T, p *Profile, name string, isNil bool) {
	if v := p.GetNil(name); isNil != v {
		t.Errorf("Unexpected %s nil %v != %v", name, isNil, v)
	}
}

func TestProfile(t *testing.T) {
	p1 := NewProfile()

	p1.Set("hello", "world")
	verifyValue(t, p1, "hello", "world")

	p1.SetNil("one")
	verifyNil(t, p1, "one", true)

	verifyNil(t, p1, "two", false)
	p1.SetNil("two")
	verifyNil(t, p1, "two", true)

	p2 := NewProfile()
	p2.Set("world", "oyster")
	p2.Set("one", "1")

	p3 := p1.Clone()
	if p3.IsMerged(p2) {
		t.Error("Unexpected merge!")
	}
	p3.Merge(p2)
	if !p3.IsMerged(p2) {
		t.Error("Should be merged!")
	}
	verifyValue(t, p3, "hello", "world")
	verifyValue(t, p3, "world", "oyster")
	verifyValue(t, p3, "one", "1")
	verifyNil(t, p3, "one", false)
	verifyNil(t, p3, "two", true)
}

func TestProfileSetUnset(t *testing.T) {
	p1 := NewProfile()

	p1.Set("hello", "world")
	verifyValue(t, p1, "hello", "world")

	p2 := NewProfile()
	p2.SetNil("hello")

	p1.Merge(p2)
	verifyNil(t, p1, "hello", true)
}

func TestMergeStrings(t *testing.T) {
	e := []string{"FOO=2", "BAR=FOO=FOOBAR", "SMURF=", "REMOVE"}
	p := NewProfile()
	p.MergeStrings(e)
	verifyValue(t, p, "FOO", "2")
	verifyValue(t, p, "BAR", "FOO=FOOBAR")
	verifyValue(t, p, "SMURF", "")
	verifyNil(t, p, "REMOVE", true)
}

func checkInList(t *testing.T, theList []string, value string) {
	for _, item := range theList {
		if value == item {
			return
		}
	}
	t.Errorf("Did not find %s in %s", value, theList)
}

func matchList(t *testing.T, expected []string, actual []string) {
	if len(expected) != len(actual) {
		t.Errorf("Mismatching length %s does not have same items as %s", expected, actual)
	}
	for _, s := range expected {
		checkInList(t, actual, s)
	}
}

func TestDiff(t *testing.T) {
	p1 := NewProfile()
	p1.MergeStrings([]string{"FOO=2", "BAR=FOO=FOOBAR", "SMURF=", "BAD=true", "REMOVE"})
	p2 := NewProfile()
	p2.MergeStrings([]string{"FOO=3", "BAR=FOO=FOOBAR", "BLURB=yes", "ALSO_REMOVE", "BAD"})
	changed, removed := p1.Diff(p2)
	matchList(t, []string{"FOO", "BLURB"}, changed)
	matchList(t, []string{"BAD"}, removed)

	changed, removed = p2.Diff(p1)
	matchList(t, []string{"FOO", "SMURF", "BAD"}, changed)
	matchList(t, []string{}, removed)

	matchList(t, []string{"BAD", "BAR", "FOO", "SMURF"}, p1.SortedNames(false))
	matchList(t, []string{"ALSO_REMOVE", "BAD", "BAR", "BLURB", "FOO"}, p2.SortedNames(true))
}

func TestFind(t *testing.T) {
	profiles := make(Profiles)
	profiles["foo"] = *NewProfile()
	_, found := profiles.FindProfile("foo")
	if !found {
		t.Error("Failed to find profile")
	}
	_, found = profiles.FindProfile("bar")
	if found {
		t.Error("Should not find profile")
	}
}
