package crypt

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/sverrirab/envirou/pkg/ini"
)

const materialFileName = "crypt.ini"
const checkPlaintext = "envirou-check-v1"
const saltSize = 16

// DefaultIterations is the PBKDF2 iteration count for new key material
// (OWASP recommendation for PBKDF2-SHA256). Stored in crypt.ini so it can
// be raised later without breaking existing tokens.
const DefaultIterations = 600000

var (
	ErrNotInitialized = errors.New("encryption is not set up")
	ErrKeyMismatch    = errors.New("key does not match key material")
)

// Material holds the persistent key derivation parameters and a check
// token used to verify passphrases without decrypting real values.
type Material struct {
	Salt       []byte
	Iterations int
	Check      string
}

// MaterialPath returns the full path of the key material file.
func MaterialPath(dir string) string {
	return filepath.Join(dir, materialFileName)
}

// LoadMaterial reads key material from dir. Returns ErrNotInitialized when
// the file does not exist. Material is never created implicitly: losing or
// regenerating the salt would silently invalidate all existing tokens.
func LoadMaterial(dir string) (*Material, error) {
	path := MaterialPath(dir)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, ErrNotInitialized
	}
	iniFile, err := ini.NewIni(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}
	salt, err := base64.RawURLEncoding.DecodeString(iniFile.GetString("crypt", "salt", ""))
	iterations, iterErr := strconv.Atoi(iniFile.GetString("crypt", "iterations", ""))
	check := iniFile.GetString("crypt", "check", "")
	if err != nil || len(salt) == 0 || iterErr != nil || iterations < 1 || !IsEncrypted(check) {
		return nil, fmt.Errorf("invalid key material in %s", path)
	}
	return &Material{Salt: salt, Iterations: iterations, Check: check}, nil
}

// CreateMaterial generates a new salt, derives a key from the passphrase
// with the given iteration count (normally DefaultIterations) and writes
// crypt.ini (0600). Fails if the file already exists.
// Returns the material and the derived key.
func CreateMaterial(dir string, passphrase string, iterations int) (*Material, []byte, error) {
	path := MaterialPath(dir)
	if _, err := os.Stat(path); err == nil {
		return nil, nil, fmt.Errorf("%s already exists", path)
	}
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, nil, err
	}
	key, err := DeriveKey(passphrase, salt, iterations)
	if err != nil {
		return nil, nil, err
	}
	check, err := Encrypt(key, checkPlaintext)
	if err != nil {
		return nil, nil, err
	}
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, nil, err
	}
	content := fmt.Sprintf(
		"; Envirou encryption key material - back this file up!\n"+
			"; Encrypted values cannot be recovered if it is lost.\n"+
			"[crypt]\nsalt=%s\niterations=%d\ncheck=%s\n",
		base64.RawURLEncoding.EncodeToString(salt), iterations, check)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return nil, nil, err
	}
	return &Material{Salt: salt, Iterations: iterations, Check: check}, key, nil
}

// VerifyKey checks a key against the material's check token.
func (m *Material) VerifyKey(key []byte) error {
	plaintext, err := Decrypt(key, m.Check)
	if err != nil || plaintext != checkPlaintext {
		return ErrKeyMismatch
	}
	return nil
}
