package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sverrirab/envirou/pkg/data"
)

// setTempConfigDir routes all config/snapshot storage to a temp directory
// so tests never touch the real ~/.config/envirou.
func setTempConfigDir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("ENVIROU_CONFIG_DIR", tmpDir)
	return tmpDir
}

func TestConfigDirOverride(t *testing.T) {
	tmpDir := setTempConfigDir(t)
	if got := GetDefaultConfigFileFolder(); got != tmpDir {
		t.Errorf("Expected config folder %q, got %q", tmpDir, got)
	}
	if got := GetSnapshotFilePath(); got != filepath.Join(tmpDir, snapshotFileName) {
		t.Errorf("Unexpected snapshot path %q", got)
	}
}

func TestSaveLoadSnapshot(t *testing.T) {
	tmpDir := setTempConfigDir(t)

	profile := data.NewProfile(false)
	profile.Set("FOO", "bar")
	profile.Set("BAZ", "qux")
	profile.Set("IGNORED_VAR", "secret")

	groups := data.NewGroups()
	groups.ParseAndAdd("..ignore", "IGNORED_*", false)

	err := SaveSnapshot(profile, groups, false)
	if err != nil {
		t.Fatalf("SaveSnapshot failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmpDir, snapshotFileName)); err != nil {
		t.Fatalf("Snapshot not written to temp config dir: %v", err)
	}

	loaded, err := LoadSnapshot(false)
	if err != nil {
		t.Fatalf("LoadSnapshot failed: %v", err)
	}
	if loaded == nil {
		t.Fatal("Expected non-nil snapshot")
	}

	if v, ok := loaded.Get("FOO"); !ok || v != "bar" {
		t.Errorf("Expected FOO=bar, got %s", v)
	}
	if v, ok := loaded.Get("BAZ"); !ok || v != "qux" {
		t.Errorf("Expected BAZ=qux, got %s", v)
	}
	// IGNORED_VAR should not be in the snapshot
	if _, ok := loaded.Get("IGNORED_VAR"); ok {
		t.Error("IGNORED_VAR should not be in snapshot")
	}
}

func TestLoadSnapshotNoFile(t *testing.T) {
	setTempConfigDir(t)

	loaded, err := LoadSnapshot(false)
	if err != nil {
		t.Fatalf("Expected nil error, got: %v", err)
	}
	if loaded != nil {
		t.Error("Expected nil snapshot when no file exists")
	}
}

func TestRemoveSnapshot(t *testing.T) {
	setTempConfigDir(t)

	// Should not error even when file doesn't exist
	err := RemoveSnapshot()
	if err != nil {
		t.Fatalf("RemoveSnapshot should not error for missing file: %v", err)
	}
}
