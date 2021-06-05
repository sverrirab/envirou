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
	twoPath := strings.Join([]string{"FOO", "BAR"}, pathListSeperator)

	SetGroupColor("red")
	SetProfileColor("blue")
	SetEnvNameColor("yellow")
	SetPathColor("green")

	beforeGroup := GroupSprintf("HELLO")
	beforeProfile := ProfileSprintf("HELLO")
	beforeDiff := DiffSprintf("FOOBAR")
	beforeEnvOne := SprintEnv("foo", "/path", *data.ParsePatterns("foo"), *data.ParsePatterns(""), false)
	beforeEnvTwo := SprintEnv("foo", twoPath, *data.ParsePatterns("foo"), *data.ParsePatterns(""), false)

	validateDifferent(t, beforeGroup, beforeProfile)
	validateDifferent(t, beforeGroup, beforeDiff)

	SetPathColor("cyan")
	afterEnvOne := SprintEnv("foo", "/path", *data.ParsePatterns("foo"), *data.ParsePatterns(""), false)
	afterEnvTwo := SprintEnv("foo", twoPath, *data.ParsePatterns("foo"), *data.ParsePatterns(""), false)
	validateSame(t, beforeEnvOne, afterEnvOne)
	validateDifferent(t, beforeEnvTwo, afterEnvTwo)

	SetGroupColor("blue")
	afterGroup := GroupSprintf("HELLO")
	validateDifferent(t, beforeGroup, afterGroup)
	validateSame(t, afterGroup, beforeProfile)
}
