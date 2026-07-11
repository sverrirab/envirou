package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sverrirab/envirou/pkg/crypt"
)

// setupCrypt creates isolated key material and returns the key. The config
// dir override also isolates snapshots for these tests.
func setupCrypt(t *testing.T, passphrase string) []byte {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("ENVIROU_CONFIG_DIR", dir)
	// Low iteration count keeps the KDF fast in tests.
	_, key, err := crypt.CreateMaterial(dir, passphrase, 1000)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

// stubPassphrase replaces the interactive prompt and counts invocations.
func stubPassphrase(t *testing.T, passphrase string, err error) *int {
	t.Helper()
	calls := 0
	orig := crypt.ReadPassphrase
	crypt.ReadPassphrase = func(prompt string) (string, error) {
		calls++
		return passphrase, err
	}
	t.Cleanup(func() { crypt.ReadPassphrase = orig })
	return &calls
}

func encryptValue(t *testing.T, key []byte, plaintext string) string {
	t.Helper()
	token, err := crypt.Encrypt(key, plaintext)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func encryptedProfileConfig(t *testing.T, key []byte) string {
	t.Helper()
	return fmt.Sprintf(`
[settings]
quiet=1

[profile:secretprofile]
TEST_SECRET=%s
TEST_PLAIN=visible

[profile:secretpath]
TEST_PATH^=%s
`, encryptValue(t, key, "hunter2"), encryptValue(t, key, "/opt/secret/bin"))
}

// --- set with encrypted values ---

func TestSetEncryptedWithEnvKey(t *testing.T) {
	key := setupCrypt(t, "pass")
	t.Setenv("ENVIROU_KEY", crypt.EncodeKey(key))
	calls := stubPassphrase(t, "", nil)

	out := executeCommandWithConfig(t, encryptedProfileConfig(t, key), "set", "secretprofile")
	if !strings.Contains(out, "hunter2") {
		t.Errorf("expected decrypted value in shell commands, got: %q", out)
	}
	if *calls != 0 {
		t.Errorf("prompt must not be called when ENVIROU_KEY is set, called %d times", *calls)
	}
}

func TestSetEncryptedPromptsForPassphrase(t *testing.T) {
	key := setupCrypt(t, "correct-pass")
	calls := stubPassphrase(t, "correct-pass", nil)

	out := executeCommandWithConfig(t, encryptedProfileConfig(t, key), "set", "secretprofile")
	if !strings.Contains(out, "hunter2") {
		t.Errorf("expected decrypted value after prompt, got: %q", out)
	}
	if *calls != 1 {
		t.Errorf("expected exactly one prompt, got %d", *calls)
	}
}

func TestSetEncryptedNoTerminal(t *testing.T) {
	key := setupCrypt(t, "pass")
	stubPassphrase(t, "", crypt.ErrNoTerminal)

	// Manual execution: executeCommand would t.Fatal on the os.Exit path,
	// so run in a subprocess-free way by checking exit via recover is not
	// possible; instead verify the helper directly.
	resetStateWithConfig(t, encryptedProfileConfig(t, key))
	if _, err := ensureKey(); err == nil {
		t.Fatal("expected error when no terminal is available")
	}
}

func TestSetEncryptedWrongPassphrase(t *testing.T) {
	key := setupCrypt(t, "correct-pass")
	stubPassphrase(t, "wrong-pass", nil)

	resetStateWithConfig(t, encryptedProfileConfig(t, key))
	if _, err := ensureKey(); err == nil || !strings.Contains(err.Error(), "incorrect passphrase") {
		t.Fatalf("expected incorrect passphrase error, got: %v", err)
	}
}

func TestSetEncryptedPrependMerges(t *testing.T) {
	key := setupCrypt(t, "pass")
	t.Setenv("ENVIROU_KEY", crypt.EncodeKey(key))
	t.Setenv("TEST_PATH", "/usr/bin")

	out := executeCommandWithConfig(t, encryptedProfileConfig(t, key), "set", "secretpath")
	if !strings.Contains(out, "/opt/secret/bin") || !strings.Contains(out, "/usr/bin") {
		t.Errorf("expected prepended decrypted path component, got: %q", out)
	}
}

func TestActiveDetectionWithKey(t *testing.T) {
	key := setupCrypt(t, "pass")
	t.Setenv("ENVIROU_KEY", crypt.EncodeKey(key))
	t.Setenv("TEST_SECRET", "hunter2")
	t.Setenv("TEST_PLAIN", "visible")

	_ = executeCommandWithConfig(t, encryptedProfileConfig(t, key), "profiles")
	if !app.isActiveProfile["secretprofile"] {
		t.Error("profile with matching decrypted values should be active when unlocked")
	}
}

func TestLockedProfilesNeverPrompt(t *testing.T) {
	key := setupCrypt(t, "pass")
	calls := stubPassphrase(t, "pass", nil)
	t.Setenv("TEST_SECRET", "hunter2")

	// Plain listing and profiles must not prompt even though a profile
	// contains encrypted values; the profile just shows inactive.
	_ = executeCommandWithConfig(t, encryptedProfileConfig(t, key), "profiles")
	if *calls != 0 {
		t.Errorf("prompt must not be called for read-only commands, called %d times", *calls)
	}
	if app.isActiveProfile["secretprofile"] {
		t.Error("locked profile must not be detected as active")
	}
}

func TestInvalidEnvKeyIgnored(t *testing.T) {
	key := setupCrypt(t, "pass")
	t.Setenv("ENVIROU_KEY", "not-a-valid-key")
	calls := stubPassphrase(t, "pass", nil)

	out := executeCommandWithConfig(t, encryptedProfileConfig(t, key), "set", "secretprofile")
	if !strings.Contains(out, "hunter2") {
		t.Errorf("expected fallback to prompt with invalid ENVIROU_KEY, got: %q", out)
	}
	if *calls != 1 {
		t.Errorf("expected one prompt after ignoring invalid key, got %d", *calls)
	}
}

// --- encrypt/decrypt commands ---

func TestEncryptCommandFirstUse(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("ENVIROU_CONFIG_DIR", dir)
	stubPassphrase(t, "new-passphrase", nil) // used for new+repeat+value prompts

	out := executeCommandWithConfig(t, testConfigForCmd, "encrypt", "--stdout")
	token := strings.TrimSpace(out)
	if !crypt.IsEncrypted(token) {
		t.Fatalf("expected token on stdout, got: %q", out)
	}
	if _, err := os.Stat(filepath.Join(dir, "crypt.ini")); err != nil {
		t.Errorf("crypt.ini not created: %v", err)
	}
	// The stub returns "new-passphrase" for the value prompt too.
	loaded, err := crypt.LoadMaterial(dir)
	if err != nil {
		t.Fatal(err)
	}
	derived, _ := crypt.DeriveKey("new-passphrase", loaded.Salt, loaded.Iterations)
	plaintext, err := crypt.Decrypt(derived, token)
	if err != nil || plaintext != "new-passphrase" {
		t.Errorf("token round trip failed: %q %v", plaintext, err)
	}
}

func TestValidateNewPassphrase(t *testing.T) {
	for _, passphrase := range []string{"", "password", "short-pass"} {
		if err := validateNewPassphrase(passphrase); err == nil {
			t.Errorf("validateNewPassphrase(%q) should reject a short passphrase", passphrase)
		}
	}
	for _, passphrase := range []string{"twelve-chars", "correct horse battery staple", "十二文字以上のパスフレーズ"} {
		if err := validateNewPassphrase(passphrase); err != nil {
			t.Errorf("validateNewPassphrase(%q) returned %v", passphrase, err)
		}
	}
}

func TestEncryptCommandWithArgAndExistingMaterial(t *testing.T) {
	key := setupCrypt(t, "pass")
	t.Setenv("ENVIROU_KEY", crypt.EncodeKey(key))

	out := executeCommandWithConfig(t, testConfigForCmd, "encrypt", "--stdout", "my-secret")
	token := strings.TrimSpace(out)
	plaintext, err := crypt.Decrypt(key, token)
	if err != nil || plaintext != "my-secret" {
		t.Errorf("expected decryptable token for my-secret, got %q (%v)", token, err)
	}
}

func TestDecryptCommand(t *testing.T) {
	key := setupCrypt(t, "pass")
	t.Setenv("ENVIROU_KEY", crypt.EncodeKey(key))
	token := encryptValue(t, key, "round-trip")

	// Plaintext goes to stderr; success here means the command completed
	// without hitting an error exit (os.Exit would abort the test binary).
	_ = executeCommandWithConfig(t, testConfigForCmd, "decrypt", token)
}

// --- dotenv with encrypted values ---

func TestDotenvEncrypted(t *testing.T) {
	key := setupCrypt(t, "pass")
	t.Setenv("ENVIROU_KEY", crypt.EncodeKey(key))
	envFile := filepath.Join(t.TempDir(), "test.env")
	content := "TEST_PLAINVAR=1\nTEST_DOTENV_SECRET=" + encryptValue(t, key, "dotenv-secret") + "\n"
	if err := os.WriteFile(envFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	out := executeCommandWithConfig(t, testConfigForCmd, "dotenv", envFile)
	if !strings.Contains(out, "dotenv-secret") {
		t.Errorf("expected decrypted dotenv value, got: %q", out)
	}
}

func TestDotenvEncryptedPrompts(t *testing.T) {
	key := setupCrypt(t, "pass")
	calls := stubPassphrase(t, "pass", nil)
	envFile := filepath.Join(t.TempDir(), "test.env")
	content := "TEST_DOTENV_SECRET=" + encryptValue(t, key, "dotenv-secret") + "\n"
	if err := os.WriteFile(envFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	out := executeCommandWithConfig(t, testConfigForCmd, "dotenv", envFile)
	if !strings.Contains(out, "dotenv-secret") {
		t.Errorf("expected decrypted dotenv value after prompt, got: %q", out)
	}
	if *calls != 1 {
		t.Errorf("expected one prompt, got %d", *calls)
	}
}
