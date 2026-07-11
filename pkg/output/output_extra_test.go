package output

import (
	"os"
	"strings"
	"testing"

	"github.com/sverrirab/envirou/pkg/data"
	"github.com/sverrirab/envirou/pkg/shell"
)

func TestIsValidColor(t *testing.T) {
	valid := []string{"green", "magenta", "red", "yellow", "blue", "cyan", "white", "black", "bold", "underline", "reverse", "deleted", "none"}
	for _, c := range valid {
		if !IsValidColor(c) {
			t.Errorf("%q should be a valid color", c)
		}
	}

	invalid := []string{"smurf", "rainbow", "", "GREEN", "Red"}
	for _, c := range invalid {
		if IsValidColor(c) {
			t.Errorf("%q should not be a valid color", c)
		}
	}
}

func TestSetDiffNames(t *testing.T) {
	NoColor(true)
	sh := shell.NewShell(false, false)
	out := NewOutput("", *data.ParsePatterns("", false), *data.ParsePatterns("", false), false, false, "cyan", "green", "cyan", "reverse", "red")

	// Without diff names, variable should use env_name color
	before := out.SprintEnv(sh, "FOO", "bar")

	// Set diff names
	out.SetDiffNames(map[string]bool{"FOO": true})
	after := out.SprintEnv(sh, "FOO", "bar")

	// The output text is the same but with no-color both are plain.
	// With color enabled they'd differ. Just verify it doesn't crash
	// and the name is present.
	if !strings.Contains(before, "FOO") || !strings.Contains(after, "FOO") {
		t.Error("expected FOO in output")
	}
}

func TestSetDiffNamesWithColor(t *testing.T) {
	forceColor(t)

	sh := shell.NewShell(false, false)
	out := NewOutput("", *data.ParsePatterns("", false), *data.ParsePatterns("", false), false, false, "cyan", "green", "cyan", "reverse", "red")

	normal := out.SprintEnv(sh, "FOO", "bar")

	out.SetDiffNames(map[string]bool{"FOO": true})
	diffed := out.SprintEnv(sh, "FOO", "bar")

	// The diff-highlighted version should use red (diff color) instead of cyan (env_name)
	if normal == diffed {
		t.Error("diff-highlighted output should differ from normal output")
	}
}

func TestIsPathVariable(t *testing.T) {
	out := NewOutput("", *data.ParsePatterns("PATH, *HOME, GOPATH", false), *data.ParsePatterns("", false), false, false, "cyan", "green", "cyan", "reverse", "red")

	if !out.IsPathVariable("PATH") {
		t.Error("PATH should be a path variable")
	}
	if !out.IsPathVariable("JAVA_HOME") {
		t.Error("JAVA_HOME should match *HOME pattern")
	}
	if !out.IsPathVariable("GOPATH") {
		t.Error("GOPATH should be a path variable")
	}
	if out.IsPathVariable("FOO") {
		t.Error("FOO should not be a path variable")
	}
}

func TestReplaceHomeTilde(t *testing.T) {
	out := NewOutput("/home/user", *data.ParsePatterns("", false), *data.ParsePatterns("", false), false, false, "cyan", "green", "cyan", "reverse", "red")

	if got := out.ReplaceHomeTilde("/home/user/bin"); got != "~/bin" {
		t.Errorf("expected '~/bin', got %q", got)
	}
	if got := out.ReplaceHomeTilde("/other/path"); got != "/other/path" {
		t.Errorf("expected unchanged path, got %q", got)
	}
	if got := out.ReplaceHomeTilde("/home/user"); got != "~" {
		t.Errorf("expected '~', got %q", got)
	}
}

func TestReplaceHomeTildeEmpty(t *testing.T) {
	out := NewOutput("", *data.ParsePatterns("", false), *data.ParsePatterns("", false), false, false, "cyan", "green", "cyan", "reverse", "red")

	// When no home is set, nothing should be replaced
	if got := out.ReplaceHomeTilde("/home/user/bin"); got != "/home/user/bin" {
		t.Errorf("expected unchanged path, got %q", got)
	}
}

func TestSprintEnvPassword(t *testing.T) {
	NoColor(true)
	sh := shell.NewShell(false, false)
	out := NewOutput("", *data.ParsePatterns("", false), *data.ParsePatterns("SECRET_*", false), false, false, "cyan", "green", "cyan", "reverse", "red")

	result := out.SprintEnv(sh, "SECRET_KEY", "my-secret-value")
	if strings.Contains(result, "my-secret-value") {
		t.Error("password should be hidden")
	}
	if !strings.Contains(result, "hidden") {
		t.Error("expected hidden marker in output")
	}
}

func TestSprintEnvPathFormatting(t *testing.T) {
	NoColor(true)
	sh := shell.NewShell(false, false)
	sep := string(os.PathListSeparator)
	out := NewOutput("/home/user", *data.ParsePatterns("PATH", false), *data.ParsePatterns("", false), false, false, "cyan", "green", "cyan", "reverse", "red")

	result := out.SprintEnv(sh, "PATH", strings.Join([]string{"/home/user/bin", "/usr/bin"}, sep))
	if !strings.Contains(result, "~/bin") {
		t.Errorf("expected tilde replacement in path, got %q", result)
	}
}

func TestSprintEnvRaw(t *testing.T) {
	sh := shell.NewShell(false, false)
	out := NewOutput("", *data.ParsePatterns("", false), *data.ParsePatterns("", false), true, false, "cyan", "green", "cyan", "reverse", "red")

	result := out.SprintEnv(sh, "FOO", "hello world")
	// In raw mode, values get shell-escaped
	if !strings.Contains(result, "'hello world'") {
		t.Errorf("expected shell-escaped value in raw mode, got %q", result)
	}
}

func TestSprintEnvRawStillMasksSensitiveValue(t *testing.T) {
	sh := shell.NewShell(false, false)
	out := NewOutput("", *data.ParsePatterns("", false), *data.ParsePatterns("SECRET_*", false), true, false, "cyan", "green", "cyan", "reverse", "red")

	result := out.SprintEnv(sh, "SECRET_KEY", "plaintext-secret")
	if strings.Contains(result, "plaintext-secret") || !strings.Contains(result, "hidden") {
		t.Errorf("raw display must still mask sensitive values, got %q", result)
	}
}

func TestSPrintProfileListEmpty(t *testing.T) {
	out := NewOutput("", *data.ParsePatterns("", false), *data.ParsePatterns("", false), false, false, "cyan", "green", "cyan", "reverse", "red")

	result := out.SPrintProfileList(nil, nil)
	if result != "" {
		t.Errorf("expected empty string for no profiles, got %q", result)
	}
}

func TestSPrintProfileListNoActive(t *testing.T) {
	NoColor(true)
	out := NewOutput("", *data.ParsePatterns("", false), *data.ParsePatterns("", false), false, false, "cyan", "green", "cyan", "reverse", "red")

	result := out.SPrintProfileList([]string{"dev", "prod"}, nil)
	if !strings.Contains(result, "dev") || !strings.Contains(result, "prod") {
		t.Errorf("expected profile names in output, got %q", result)
	}
}
