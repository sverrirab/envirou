package config

import (
	"os"
	"testing"

	"github.com/sverrirab/envirou/pkg/data"
)

func TestSaveLoadSnapshot(t *testing.T) {
	// Use a temp dir so we don't write to the real config folder
	tmpDir, err := os.MkdirTemp("", "snapshot")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Override snapshot path by writing directly to a temp file
	profile := data.NewProfile(false)
	profile.Set("FOO", "bar")
	profile.Set("BAZ", "qux")
	profile.Set("IGNORED_VAR", "secret")

	groups := data.NewGroups()
	groups.ParseAndAdd("..ignore", "IGNORED_*", false)

	err = SaveSnapshot(profile, groups, false)
	if err != nil {
		t.Fatalf("SaveSnapshot failed: %v", err)
	}
	defer RemoveSnapshot()

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
	// Ensure no snapshot file exists
	RemoveSnapshot()

	loaded, err := LoadSnapshot(false)
	if err != nil {
		t.Fatalf("Expected nil error, got: %v", err)
	}
	if loaded != nil {
		t.Error("Expected nil snapshot when no file exists")
	}
}

func TestRemoveSnapshot(t *testing.T) {
	// Should not error even when file doesn't exist
	err := RemoveSnapshot()
	if err != nil {
		t.Fatalf("RemoveSnapshot should not error for missing file: %v", err)
	}
}
