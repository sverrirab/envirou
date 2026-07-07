package crypt

import (
	"errors"
	"os"
	"runtime"
	"testing"
)

// Low iteration count keeps the KDF fast in tests.
const testIterations = 1000

func TestCreateLoadMaterial(t *testing.T) {
	dir := t.TempDir()
	material, key, err := CreateMaterial(dir, "test-passphrase", testIterations)
	if err != nil {
		t.Fatalf("CreateMaterial failed: %v", err)
	}
	if err := material.VerifyKey(key); err != nil {
		t.Fatalf("VerifyKey failed on fresh material: %v", err)
	}

	loaded, err := LoadMaterial(dir)
	if err != nil {
		t.Fatalf("LoadMaterial failed: %v", err)
	}
	if loaded.Iterations != testIterations {
		t.Errorf("iterations mismatch: %d", loaded.Iterations)
	}
	// Re-derive from the loaded material and verify against the canary.
	derived, err := DeriveKey("test-passphrase", loaded.Salt, loaded.Iterations)
	if err != nil {
		t.Fatal(err)
	}
	if err := loaded.VerifyKey(derived); err != nil {
		t.Errorf("re-derived key should verify: %v", err)
	}

	wrong, _ := DeriveKey("wrong-passphrase", loaded.Salt, loaded.Iterations)
	if err := loaded.VerifyKey(wrong); !errors.Is(err, ErrKeyMismatch) {
		t.Errorf("expected ErrKeyMismatch, got %v", err)
	}
}

func TestCreateMaterialRefusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	if _, _, err := CreateMaterial(dir, "pass", testIterations); err != nil {
		t.Fatal(err)
	}
	if _, _, err := CreateMaterial(dir, "pass", testIterations); err == nil {
		t.Error("CreateMaterial must not overwrite existing key material")
	}
}

func TestLoadMaterialNotInitialized(t *testing.T) {
	if _, err := LoadMaterial(t.TempDir()); !errors.Is(err, ErrNotInitialized) {
		t.Errorf("expected ErrNotInitialized, got %v", err)
	}
}

func TestLoadMaterialCorrupt(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(MaterialPath(dir), []byte("[crypt]\nsalt=!!!\niterations=x\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadMaterial(dir); err == nil {
		t.Error("expected error for corrupt material")
	}
}

func TestMaterialFilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file mode check not meaningful on Windows")
	}
	dir := t.TempDir()
	if _, _, err := CreateMaterial(dir, "pass", testIterations); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(MaterialPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected 0600 permissions, got %v", info.Mode().Perm())
	}
}
