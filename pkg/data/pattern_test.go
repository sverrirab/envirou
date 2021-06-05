package data

import (
	"testing"
)

func TestParse(t *testing.T) {
	p := ParsePatterns("SMURF, FOO*,,  *GLOB* , X,")
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
	p := ParsePatterns("")
	if len(*p) != 0 {
		t.Error("Empty length not zero")
	}
}

func TestMatchSimpleMatch(t *testing.T) {
	if !Match("A", "A") {
		t.Error("These should match")
	}
}

func TestMatchSimpleNoMatch(t *testing.T) {
	if Match("a", "A") {
		t.Error("These should not match")
	}
	if Match("A", "") {
		t.Error("These should not match")
	}
}

func TestMatchGlobMatch(t *testing.T) {
	if !Match("PATH", "PA*") {
		t.Error("These should match")
	}
	if !Match("PATH", "PATH*") {
		t.Error("These should match")
	}
	if !Match("PATH", "*TH*") {
		t.Error("These should match")
	}
	if !Match("PATH", "*TH") {
		t.Error("These should match")
	}
	if !Match("PATH", "**") {
		t.Error("These should match")
	}
}

func TestMatchGlobNoMatch(t *testing.T) {
	if Match("path", "*PATH*") {
		t.Error("These should not match")
	}
	if Match("PXTH", "*PATH*") {
		t.Error("These should not match")
	}
	if Match("PATHY", "*PATH") {
		t.Error("These should not match")
	}
	if Match("XPATH", "PATH*") {
		t.Error("These should not match")
	}
}

func TestMatchAny(t *testing.T) {
	if !MatchAny("FOOBAR", ParsePatterns("SMURF,BAR,FOO*")) {
		t.Error("FOO* should match")
	}
	if MatchAny("FOOBAR", ParsePatterns("SMURF,BAR,FOOBA,OOBAR")) {
		t.Error("FOOBAR should not match")
	}
	if !MatchAny("FOOBAR", ParsePatterns("SMURF,*,FOOBA,OOBAR")) {
		t.Error("* should match")
	}
}
