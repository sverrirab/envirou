package output

import (
	"os"
	"strings"
	"testing"

	"github.com/sverrirab/envirou/pkg/data"
)

func validateDifferent(t *testing.T, before, after string) {
	if before == after {
		t.Errorf("Unexpected equality %s == %s", before, after)
	}
}

func validateSame(t *testing.T, before, after string) {
	if before != after {
		t.Errorf("Unexpected inequality %s != %s", before, after)
	}
}

func TestColorChange(t *testing.T) {
	NoColor(false) // Test need to force color.

	const pathListSeperator = string(os.PathListSeparator)
	profileNames := []string{"p1", "p2", "p3"}
	activeNames := []string{"p3"}

	twoPath := strings.Join([]string{"FOO", "BAR"}, pathListSeperator)

	out1 := NewOutput("", *data.ParsePatterns("foo"), *data.ParsePatterns("bar"), false, "red", "blue", "cyan", "green", "white")
	out2 := NewOutput("", *data.ParsePatterns("foo"), *data.ParsePatterns("bar"), false, "magenta", "yellow", "cyan", "black", "bold")
	out3 := NewOutput("/X", *data.ParsePatterns("foo"), *data.ParsePatterns(""), false, "red", "blue", "cyan", "green", "white")

	beforeGroup := out1.GroupSprintf("HELLO")
	beforeProfile := out1.ProfileSprintf("HELLO")
	beforeProfileList := out1.SPrintProfileList(profileNames, activeNames)
	beforeDiff := out1.DiffSprintf("FOOBAR")
	beforeEnvOne := out1.SprintEnv("foo", "/path")
	beforeEnvTwo := out1.SprintEnv("foo", twoPath)

	validateDifferent(t, beforeGroup, beforeProfile)
	validateDifferent(t, beforeGroup, beforeDiff)

	afterEnvOne := out2.SprintEnv("foo", "/path")
	afterEnvTwo := out2.SprintEnv("foo", twoPath)
	
	validateSame(t, beforeEnvOne, afterEnvOne)
	validateDifferent(t, beforeEnvTwo, afterEnvTwo)

	// Test password hiding 
	beforeEnvThree := out3.SprintEnv("bar", "smurfy")
	afterEnvThree := out1.SprintEnv("bar", "smurfy")
	validateDifferent(t, beforeEnvThree, afterEnvThree)

	// Test tilde replacement
	tildeEnv := out3.SprintEnv("foo", "/X/Y")
	if ! strings.Contains(tildeEnv, "~/Y") {
		t.Errorf("Did not find path replacement: %s", tildeEnv)
	}

	afterProfileList := out2.SPrintProfileList(profileNames, activeNames)
	validateDifferent(t, beforeProfileList, afterProfileList)
}

func TestColorMap(t *testing.T) {
	color, found := mapColor("smurf")
	if found {
		t.Errorf("Unexpected color: %v", color)
	}

	c1 := mapColorDefault("reset", "foo")
	c2 := mapColorDefault("foo", "reset")
	c3 := mapColorDefault("foo", "bar")
	if c1 != c2 || c2 != c3 {
		t.Errorf("Default color failure %v - %v - %v", c1, c2, c3)
	}

	c4 := mapColorDefault("reverse", "smurf")
	c5 := mapColorDefault("x", "underline")
	if c4 == c5 {
		t.Errorf("Color match error %v - %v", c4, c5)
	}
}