package cmd

import (
	"runtime"
	"strings"
	"testing"

	"github.com/sverrirab/envirou/pkg/config"
	"github.com/sverrirab/envirou/pkg/crypt"
	"github.com/sverrirab/envirou/pkg/data"
)

func TestUnlockCommand(t *testing.T) {
	key := setupCrypt(t, "pass")
	calls := stubPassphrase(t, "pass", nil)

	out := executeCommandWithConfig(t, testConfigForCmd, "unlock")
	if !strings.Contains(out, "ENVIROU_KEY") || !strings.Contains(out, crypt.EncodeKey(key)) {
		t.Errorf("expected export of derived key, got: %q", out)
	}
	if *calls != 1 {
		t.Errorf("expected one prompt, got %d", *calls)
	}
}

func TestEnvirouKeyExcludedFromDiffsWithoutDisplayGroup(t *testing.T) {
	groups := data.NewGroups()
	filtered := filterIgnored([]string{"ENVIROU_KEY", "VISIBLE"}, groups, false)
	if len(filtered) != 1 || filtered[0] != "VISIBLE" {
		t.Errorf("expected only VISIBLE after filtering, got %v", filtered)
	}
}

func TestUnlockPrintKey(t *testing.T) {
	key := setupCrypt(t, "pass")
	stubPassphrase(t, "pass", nil)

	out := executeCommandWithConfig(t, testConfigForCmd, "unlock", "--print-key")
	if strings.TrimSpace(out) != crypt.EncodeKey(key) {
		t.Errorf("expected bare key on stdout, got: %q", out)
	}
	if strings.Contains(out, "export") {
		t.Error("--print-key must not emit shell commands")
	}
}

func TestLockCommand(t *testing.T) {
	setupCrypt(t, "pass")
	t.Setenv("ENVIROU_KEY", "whatever")

	out := executeCommandWithConfig(t, testConfigForCmd, "lock")
	want := "unset ENVIROU_KEY"
	if runtime.GOOS == "windows" {
		want = "set ENVIROU_KEY="
	}
	if !strings.Contains(out, want) {
		t.Errorf("expected %q command, got: %q", want, out)
	}
}

func TestLockAlreadyLocked(t *testing.T) {
	setupCrypt(t, "pass")

	out := executeCommandWithConfig(t, testConfigForCmd, "lock")
	if out != "" {
		t.Errorf("expected no shell commands when already locked, got: %q", out)
	}
}

// ENVIROU_KEY must never leak: masked in display and excluded from
// snapshots even though the test config overrides password patterns
// and groups (the protection is hardcoded, not config-dependent).
func TestEnvirouKeyHygiene(t *testing.T) {
	key := setupCrypt(t, "pass")
	t.Setenv("ENVIROU_KEY", crypt.EncodeKey(key))

	_ = executeCommandWithConfig(t, testConfigForCmd, "snapshot")
	snapshot, err := config.LoadSnapshot(false)
	if err != nil || snapshot == nil {
		t.Fatalf("failed to load snapshot: %v", err)
	}
	if _, ok := snapshot.Get("ENVIROU_KEY"); ok {
		t.Error("ENVIROU_KEY must be excluded from snapshots")
	}

	masked := app.out.SprintEnv(app.sh, "ENVIROU_KEY", crypt.EncodeKey(key))
	if strings.Contains(masked, crypt.EncodeKey(key)) {
		t.Error("ENVIROU_KEY value must be masked in display output")
	}
	if app.configuration.Groups.IsIgnored("ENVIROU_KEY", app.caseInsensitive) {
		t.Error("ENVIROU_KEY should be visible by name rather than filtered from display")
	}
}
