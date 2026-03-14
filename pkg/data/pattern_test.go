package data

import (
	"testing"
)

func TestParse(t *testing.T) {
	p := ParsePatterns("SMURF, FOO*,,  *GLOB* , X,", false)
	if (*p)[0] != Pattern("SMURF") {
		t.Error("0 != SMURF")
	}
	if (*p)[1] != Pattern("FOO*") {
		t.Error("1 != FOO")
	}
	if (*p)[2] != Pattern("*GLOB*") {
		t.Error("2 != GLOB")
	}
	if (*p)[3] != Pattern("X") {
		t.Error("3 != X")
	}
	if len(*p) != 4 {
		t.Error("len != 4")
	}
}

func TestEmptyParse(t *testing.T) {
	p := ParsePatterns("", false)
	if len(*p) != 0 {
		t.Error("Empty length not zero")
	}
}

func TestMatchSimpleMatch(t *testing.T) {
	if !Match("A", "A", false) {
		t.Error("These should match")
	}
}

func TestMatchSimpleNoMatch(t *testing.T) {
	if Match("a", "A", false) {
		t.Error("These should not match")
	}
	if Match("A", "", false) {
		t.Error("These should not match")
	}
}

func TestMatchGlobMatch(t *testing.T) {
	if !Match("PATH", "PA*", false) {
		t.Error("These should match")
	}
	if !Match("PATH", "PATH*", false) {
		t.Error("These should match")
	}
	if !Match("PATH", "*TH*", false) {
		t.Error("These should match")
	}
	if !Match("PATH", "*TH", false) {
		t.Error("These should match")
	}
	if !Match("PATH", "**", false) {
		t.Error("These should match")
	}
}

func TestMatchGlobNoMatch(t *testing.T) {
	if Match("path", "*PATH*", false) {
		t.Error("These should not match")
	}
	if Match("PXTH", "*PATH*", false) {
		t.Error("These should not match")
	}
	if Match("PATHY", "*PATH", false) {
		t.Error("These should not match")
	}
	if Match("XPATH", "PATH*", false) {
		t.Error("These should not match")
	}
}

func TestMatchAny(t *testing.T) {
	if !MatchAny("FOOBAR", ParsePatterns("SMURF,BAR,FOO*", false), false) {
		t.Error("FOO* should match")
	}
	if MatchAny("FOOBAR", ParsePatterns("SMURF,BAR,FOOBA,OOBAR", false), false) {
		t.Error("FOOBAR should not match")
	}
	if !MatchAny("FOOBAR", ParsePatterns("SMURF,*,FOOBA,OOBAR", false), false) {
		t.Error("* should match")
	}
}

func TestMatchCaseInsensitive(t *testing.T) {
	p := ParsePatterns("PATH*", true)
	if (*p)[0] != Pattern("PATH*") {
		t.Error("Pattern should be uppercased")
	}
	if !Match("path", "PATH*", true) {
		t.Error("Case insensitive match should work")
	}
	if !Match("Path_Extra", "PATH*", true) {
		t.Error("Case insensitive prefix match should work")
	}
}
