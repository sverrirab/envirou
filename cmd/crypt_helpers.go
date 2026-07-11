package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/sverrirab/envirou/pkg/config"
	"github.com/sverrirab/envirou/pkg/crypt"
	"github.com/sverrirab/envirou/pkg/data"
	"github.com/sverrirab/envirou/pkg/output"
)

// keyFromEnvSilent returns the key from ENVIROU_KEY, or nil when it is
// absent or invalid. Never prompts.
func keyFromEnvSilent() []byte {
	value := os.Getenv("ENVIROU_KEY")
	if value == "" {
		return nil
	}
	key, err := crypt.DecodeKey(value)
	if err != nil {
		if verbose {
			output.Printf("Warning: ignoring invalid ENVIROU_KEY\n")
		}
		return nil
	}
	return key
}

// notInitializedError explains how to set up or restore key material.
func notInitializedError() error {
	return fmt.Errorf("encryption is not set up (missing %s): run 'envirou encrypt' first or copy crypt.ini from the machine where the values were encrypted",
		crypt.MaterialPath(config.GetDefaultConfigFileFolder()))
}

// ensureKey returns a usable key, in order of preference: the key cached in
// this invocation, a canary-verified ENVIROU_KEY, or a passphrase prompt.
// Prompts at most once per invocation.
func ensureKey() ([]byte, error) {
	if app.cryptKey != nil {
		return app.cryptKey, nil
	}
	envKey := keyFromEnvSilent()
	material, err := crypt.LoadMaterial(config.GetDefaultConfigFileFolder())
	if err != nil {
		if errors.Is(err, crypt.ErrNotInitialized) {
			if envKey != nil {
				// No local key material but a key was provided (e.g. CI):
				// use it; individual tokens still authenticate via GCM.
				app.cryptKey = envKey
				return envKey, nil
			}
			return nil, notInitializedError()
		}
		return nil, err
	}
	if envKey != nil {
		if err := material.VerifyKey(envKey); err == nil {
			app.cryptKey = envKey
			return envKey, nil
		}
		output.Printf("Warning: ENVIROU_KEY does not match the key material, prompting for passphrase\n")
	}
	passphrase, err := crypt.ReadPassphrase("Envirou passphrase: ")
	if err != nil {
		return nil, err
	}
	key, err := crypt.DeriveKey(passphrase, material.Salt, material.Iterations)
	if err != nil {
		return nil, err
	}
	if err := material.VerifyKey(key); err != nil {
		return nil, errors.New("incorrect passphrase")
	}
	app.cryptKey = key
	return key, nil
}

// hasEncryptedValues reports whether any profile value is an enc:v1: token.
func hasEncryptedValues(p *data.Profile) bool {
	for _, name := range p.SortedNames(false) {
		if value, _ := p.Get(name); crypt.IsEncrypted(value) {
			return true
		}
	}
	return false
}

// decryptProfileInPlace replaces encrypted values with their plaintext,
// preserving merge modes. Errors name the failing variable.
func decryptProfileInPlace(p *data.Profile, key []byte) error {
	for _, name := range p.SortedNames(false) {
		value, _ := p.Get(name)
		if !crypt.IsEncrypted(value) {
			continue
		}
		plaintext, err := crypt.Decrypt(key, value)
		if err != nil {
			if errors.Is(err, crypt.ErrMalformedToken) {
				return fmt.Errorf("invalid encrypted value for %s: malformed token", name)
			}
			return fmt.Errorf("cannot decrypt %s: value was encrypted with different key material (crypt.ini mismatch?)", name)
		}
		p.SetWithMode(name, plaintext, p.GetMergeMode(name))
	}
	return nil
}

// decryptProfilesSilently decrypts all profiles that contain encrypted
// values using ENVIROU_KEY, when present and valid. Failures leave the
// profile untouched (tokens intact); it never prompts. Called from
// initConfig so active-profile detection sees plaintext when unlocked.
func decryptProfilesSilently(profiles data.Profiles) {
	key := keyFromEnvSilent()
	if key == nil {
		return
	}
	for name, profile := range profiles {
		if !hasEncryptedValues(&profile) {
			continue
		}
		decrypted := profile.Clone()
		if err := decryptProfileInPlace(decrypted, key); err != nil {
			if verbose {
				output.Printf("Warning: profile %s: %s\n", name, err.Error())
			}
			continue
		}
		profiles[name] = *decrypted
	}
}
