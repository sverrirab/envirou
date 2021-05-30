package util

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
	p3.Merge(p2)
	verifyValue(t, p3, "hello", "world")
	verifyValue(t, p3, "world", "oyster")
	verifyValue(t, p3, "one", "1")
	verifyNil(t, p3, "one", false)
	verifyNil(t, p3, "two", true)
}

func TestMergeStrings(t *testing.T) {
	e := [...]string{"FOO=2", "BAR=FOO=FOOBAR", "SMURF=", "REMOVE"}
	p := NewProfile()
	p.MergeStrings(e[:])
	verifyValue(t, p, "FOO", "2")
	verifyValue(t, p, "BAR", "FOO=FOOBAR")
	verifyValue(t, p, "SMURF", "")
	verifyNil(t, p, "REMOVE", true)
}
