package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/sverrirab/envirou/pkg/config"
)

const testConfigForCmd = `
[settings]
quiet=1

[groups]
test=TEST_*

[profile:dev]
TEST_ENV=development

[profile:prod]
TEST_ENV=production
TEST_DEBUG
`

// executeCommand sets up a test config, resets global state, and executes
// the root command with the given args. Returns captured stdout.
func executeCommand(t *testing.T, args ...string) string {
	t.Helper()

	// Create temp config
	file, err := os.CreateTemp("", "config")
	if err != nil {
		t.Fatal(err)
	}
	name := file.Name()
	t.Cleanup(func() { os.Remove(name) })

	_, err = file.WriteString(testConfigForCmd)
	if err != nil {
		t.Fatal(err)
	}
	file.Close()

	// Reset global state
	cfgFile = name
	bashBootstrap = "#!/bin/bash\nfunction ev() { eval \"$(envirou \"$@\")\"; }"
	powershellBootstrap = "function ev { Invoke-Expression (envirou $args) }"
	powershellPrompt = "function prompt { \"PS> \" }"
	batBootstrap = "@FOR /F %%g IN (`envirou %*`) do @%%g"
	verbose = false
	noColor = true
	dryRun = false
	displayUnformatted = false
	outputPowerShell = false
	showAllGroups = false
	actionShowGroups = nil
	addPrompt = false
	showActiveProfilesOnly = false
	showInactiveProfilesOnly = false
	snapshotReset = false
	diffSaveProfile = ""

	// Reset cobra flag "changed" state so mutually exclusive checks work
	rootCmd.Flags().VisitAll(func(f *pflag.Flag) { f.Changed = false })
	rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) { f.Changed = false })
	for _, c := range rootCmd.Commands() {
		c.Flags().VisitAll(func(f *pflag.Flag) { f.Changed = false })
	}

	// Capture stdout (where shell commands are printed)
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs(args)
	err = rootCmd.Execute()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	if err != nil {
		t.Fatalf("Command %v failed: %v", args, err)
	}

	return buf.String()
}

// --- Bootstrap tests ---

func TestBootstrapBash(t *testing.T) {
	out := executeCommand(t, "bootstrap", "bash")
	if !strings.Contains(out, "function ev()") {
		t.Errorf("Expected bash ev function, got: %s", out)
	}
	// Shebang should be stripped
	if strings.Contains(out, "#!/bin/bash") {
		t.Error("Shebang line should be removed")
	}
}

func TestBootstrapZsh(t *testing.T) {
	out := executeCommand(t, "bootstrap", "zsh")
	if !strings.Contains(out, "function ev()") {
		t.Errorf("Expected zsh ev function (same as bash), got: %s", out)
	}
}

func TestBootstrapPowershell(t *testing.T) {
	out := executeCommand(t, "bootstrap", "powershell")
	if !strings.Contains(out, "Invoke-Expression") {
		t.Errorf("Expected PowerShell ev function, got: %s", out)
	}
	// Prompt should not be included without --prompt flag
	if strings.Contains(out, "function prompt") {
		t.Error("Prompt should not be included without --prompt flag")
	}
}

func TestBootstrapPowershellWithPrompt(t *testing.T) {
	out := executeCommand(t, "bootstrap", "powershell", "--prompt")
	if !strings.Contains(out, "Invoke-Expression") {
		t.Errorf("Expected PowerShell ev function, got: %s", out)
	}
	if !strings.Contains(out, "function prompt") {
		t.Errorf("Expected prompt function with --prompt flag, got: %s", out)
	}
}

func TestBootstrapBat(t *testing.T) {
	out := executeCommand(t, "bootstrap", "bat")
	if !strings.Contains(out, "FOR /F") {
		t.Errorf("Expected batch wrapper, got: %s", out)
	}
}

func TestBootstrapInvalidArg(t *testing.T) {
	// Can't use executeCommand because we expect an error
	rootCmd.SetArgs([]string{"bootstrap", "fish"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid shell type")
	}
}

func TestBootstrapNoArg(t *testing.T) {
	rootCmd.SetArgs([]string{"bootstrap"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error when no shell type provided")
	}
}

// --- Set tests ---

func TestSetProfile(t *testing.T) {
	t.Setenv("TEST_ENV", "old_value")
	out := executeCommand(t, "set", "dev")
	if !strings.Contains(out, "TEST_ENV") || !strings.Contains(out, "development") {
		t.Errorf("Expected TEST_ENV=development in output, got: %s", out)
	}
}

func TestSetProfileAlreadyActive(t *testing.T) {
	t.Setenv("TEST_ENV", "development")
	out := executeCommand(t, "set", "dev")
	// No shell commands should be generated since profile is already active
	if strings.Contains(out, "export") || strings.Contains(out, "TEST_ENV") {
		t.Errorf("Expected no shell commands for already-active profile, got: %s", out)
	}
}

func TestSetMultipleProfiles(t *testing.T) {
	t.Setenv("TEST_ENV", "old_value")
	out := executeCommand(t, "set", "dev", "prod")
	// prod is applied after dev, so TEST_ENV should be "production"
	if !strings.Contains(out, "production") {
		t.Errorf("Expected TEST_ENV=production (last profile wins), got: %s", out)
	}
}

func TestSetMissingProfile(t *testing.T) {
	out := executeCommand(t, "set", "nonexistent")
	// Should still succeed (exit 0) but no shell commands
	if strings.Contains(out, "export") {
		t.Errorf("Expected no shell commands for missing profile, got: %s", out)
	}
}

func TestSetPartialMissing(t *testing.T) {
	t.Setenv("TEST_ENV", "old_value")
	out := executeCommand(t, "set", "dev", "nonexistent")
	// dev should still be applied
	if !strings.Contains(out, "development") {
		t.Errorf("Expected dev profile to be applied despite missing profile, got: %s", out)
	}
}

// --- Profiles tests ---

func TestProfilesList(t *testing.T) {
	_ = executeCommand(t, "profiles")
	// Verify profiles were populated from config
	if len(app.profileNames) != 2 {
		t.Errorf("Expected 2 profiles, got %d: %v", len(app.profileNames), app.profileNames)
	}
	if !contains(app.profileNames, "dev") || !contains(app.profileNames, "prod") {
		t.Errorf("Expected dev and prod profiles, got: %v", app.profileNames)
	}
}

func TestProfilesActiveOnly(t *testing.T) {
	t.Setenv("TEST_ENV", "development")
	_ = executeCommand(t, "profiles", "--active")
	if !contains(app.activeProfileNames, "dev") {
		t.Errorf("Expected dev to be active, active: %v", app.activeProfileNames)
	}
}

func TestProfilesInactiveOnly(t *testing.T) {
	t.Setenv("TEST_ENV", "development")
	_ = executeCommand(t, "profiles", "--inactive")
	if !contains(app.inactiveProfileNames, "prod") {
		t.Errorf("Expected prod to be inactive, inactive: %v", app.inactiveProfileNames)
	}
}

// --- Groups tests ---

func TestGroupsList(t *testing.T) {
	_ = executeCommand(t, "groups")
	names := app.configuration.Groups.GetAllNames()
	if len(names) != 1 || names[0] != "test" {
		t.Errorf("Expected [test] group, got: %v", names)
	}
}

// --- Version tests ---

func TestVersionCommand(t *testing.T) {
	_ = executeCommand(t, "version")
	// version output goes to stderr, so just verify it doesn't error
}

// --- Root command tests ---

func TestRootCommand(t *testing.T) {
	t.Setenv("TEST_VAR", "hello")
	_ = executeCommand(t, "--no-color")
	// Just verify it runs without error
}

func TestRootCommandDryRun(t *testing.T) {
	_ = executeCommand(t, "--dry-run")
	// Just verify it runs without error
}

// --- Dotenv tests ---

func writeTempEnvFile(t *testing.T, content string) string {
	t.Helper()
	envFile, err := os.CreateTemp("", "dotenv")
	if err != nil {
		t.Fatal(err)
	}
	name := envFile.Name()
	t.Cleanup(func() { os.Remove(name) })
	envFile.WriteString(content)
	envFile.Close()
	return name
}

func TestDotenvCommand(t *testing.T) {
	name := writeTempEnvFile(t, "MY_VAR=hello\nMY_OTHER=world\n# comment\n")
	out := executeCommand(t, "dotenv", name)
	if !strings.Contains(out, "MY_VAR") || !strings.Contains(out, "hello") {
		t.Errorf("Expected MY_VAR=hello in output, got: %s", out)
	}
	if !strings.Contains(out, "MY_OTHER") || !strings.Contains(out, "world") {
		t.Errorf("Expected MY_OTHER=world in output, got: %s", out)
	}
}

func TestDotenvAlias(t *testing.T) {
	name := writeTempEnvFile(t, "ALIAS_VAR=works\n")
	out := executeCommand(t, ".env", name)
	if !strings.Contains(out, "ALIAS_VAR") || !strings.Contains(out, "works") {
		t.Errorf("Expected .env alias to work, got: %s", out)
	}
}

func TestDotenvQuotedValues(t *testing.T) {
	name := writeTempEnvFile(t, `QUOTED="hello world"`+"\n")
	out := executeCommand(t, "dotenv", name)
	if !strings.Contains(out, "hello world") {
		t.Errorf("Expected unquoted value 'hello world' in output, got: %s", out)
	}
}

func TestDotenvExportPrefix(t *testing.T) {
	name := writeTempEnvFile(t, "export EXPORTED_VAR=value\n")
	out := executeCommand(t, "dotenv", name)
	if !strings.Contains(out, "EXPORTED_VAR") || !strings.Contains(out, "value") {
		t.Errorf("Expected EXPORTED_VAR=value in output, got: %s", out)
	}
}

func TestDotenvMultipleFiles(t *testing.T) {
	base := writeTempEnvFile(t, "FOO=base\nBAR=only_in_base\n")
	override := writeTempEnvFile(t, "FOO=override\nBAZ=only_in_override\n")
	out := executeCommand(t, "dotenv", base, override)
	// FOO should be "override" (last file wins)
	if !strings.Contains(out, "override") {
		t.Errorf("Expected FOO=override (last file wins), got: %s", out)
	}
	// BAR from base should be present
	if !strings.Contains(out, "BAR") || !strings.Contains(out, "only_in_base") {
		t.Errorf("Expected BAR=only_in_base from base file, got: %s", out)
	}
	// BAZ from override should be present
	if !strings.Contains(out, "BAZ") || !strings.Contains(out, "only_in_override") {
		t.Errorf("Expected BAZ=only_in_override from override file, got: %s", out)
	}
}

// --- Config command tests ---

func TestConfigWithEditor(t *testing.T) {
	t.Setenv("EDITOR", "echo")
	out := executeCommand(t, "config")
	// Should generate a shell command to launch the editor
	if !strings.Contains(out, "echo") {
		t.Errorf("Expected editor command in output, got: %s", out)
	}
}

// --- Snapshot tests ---

func TestSnapshotCommand(t *testing.T) {
	t.Setenv("TEST_SNAP", "value1")
	_ = executeCommand(t, "snapshot")
	t.Cleanup(func() { config.RemoveSnapshot() })

	// Verify snapshot was saved by loading it
	snapshot, err := config.LoadSnapshot(false)
	if err != nil {
		t.Fatalf("Failed to load snapshot: %v", err)
	}
	if snapshot == nil {
		t.Fatal("Expected snapshot to be saved")
	}
	if v, ok := snapshot.Get("TEST_SNAP"); !ok || v != "value1" {
		t.Errorf("Expected TEST_SNAP=value1 in snapshot, got %s", v)
	}
}

func TestSnapshotReset(t *testing.T) {
	// First save a snapshot
	_ = executeCommand(t, "snapshot")
	// Now reset it
	_ = executeCommand(t, "snapshot", "--reset")

	snapshot, err := config.LoadSnapshot(false)
	if err != nil {
		t.Fatalf("Failed to load snapshot: %v", err)
	}
	if snapshot != nil {
		t.Error("Expected snapshot to be removed after reset")
	}
}

// --- Diff tests ---

func TestDiffNoSnapshot(t *testing.T) {
	config.RemoveSnapshot()
	// Should not error, just print message
	_ = executeCommand(t, "diff")
}

func TestDiffWithChanges(t *testing.T) {
	// Save a snapshot with TEST_DIFF set
	t.Setenv("TEST_DIFF", "before")
	_ = executeCommand(t, "snapshot")
	t.Cleanup(func() { config.RemoveSnapshot() })

	// Change the env and run diff
	t.Setenv("TEST_DIFF", "after")
	t.Setenv("TEST_NEW", "added")
	_ = executeCommand(t, "diff")
}
