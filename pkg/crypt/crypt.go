// Package crypt implements symmetric encryption of individual environment
// variable values. Values are stored as self-contained tokens with the
// "enc:v1:" prefix and decrypted when applied to the shell.
package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
)

// TokenPrefix marks an encrypted value in profiles and .env files.
const TokenPrefix = "enc:v1:"

const keySize = 32 // AES-256

var (
	ErrMalformedToken = errors.New("malformed encrypted value")
	ErrDecryptFailed  = errors.New("decryption failed")
	ErrInvalidKey     = errors.New("invalid key")
)

// tokenEncoding is base64 without "+", "/" or padding. The INI parser scans
// whole lines for the "+=" operator, so a standard-base64 token ending in
// "+=" would misparse; RawURL characters are also shell-safe unquoted.
var tokenEncoding = base64.RawURLEncoding

// IsEncrypted reports whether a value is an encrypted token.
func IsEncrypted(value string) bool {
	return strings.HasPrefix(value, TokenPrefix)
}

// Encrypt seals plaintext with AES-256-GCM and returns an enc:v1: token.
func Encrypt(key []byte, plaintext string) (string, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return TokenPrefix + tokenEncoding.EncodeToString(sealed), nil
}

// Decrypt opens an enc:v1: token. Returns ErrMalformedToken for anything
// that is not a well-formed token and ErrDecryptFailed when the key does
// not match or the token was tampered with.
func Decrypt(key []byte, token string) (string, error) {
	if !IsEncrypted(token) {
		return "", ErrMalformedToken
	}
	raw, err := tokenEncoding.DecodeString(token[len(TokenPrefix):])
	if err != nil {
		return "", ErrMalformedToken
	}
	gcm, err := newGCM(key)
	if err != nil {
		return "", err
	}
	if len(raw) <= gcm.NonceSize() {
		return "", ErrMalformedToken
	}
	plaintext, err := gcm.Open(nil, raw[:gcm.NonceSize()], raw[gcm.NonceSize():], nil)
	if err != nil {
		return "", ErrDecryptFailed
	}
	return string(plaintext), nil
}

func newGCM(key []byte) (cipher.AEAD, error) {
	if len(key) != keySize {
		return nil, ErrInvalidKey
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// DeriveKey derives a 32-byte key from a passphrase using PBKDF2-SHA256.
func DeriveKey(passphrase string, salt []byte, iterations int) ([]byte, error) {
	return pbkdf2.Key(sha256.New, passphrase, salt, iterations, keySize)
}

// EncodeKey encodes a key for storage in the ENVIROU_KEY variable.
func EncodeKey(key []byte) string {
	return tokenEncoding.EncodeToString(key)
}

// DecodeKey decodes an ENVIROU_KEY value and validates its length.
func DecodeKey(s string) ([]byte, error) {
	key, err := tokenEncoding.DecodeString(s)
	if err != nil || len(key) != keySize {
		return nil, ErrInvalidKey
	}
	return key, nil
}
