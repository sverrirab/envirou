package crypt

import (
	"bytes"
	"strings"
	"testing"
)

func testKey(t *testing.T) []byte {
	t.Helper()
	key, err := DeriveKey("test-passphrase", []byte("0123456789abcdef"), 1000)
	if err != nil {
		t.Fatalf("DeriveKey failed: %v", err)
	}
	return key
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := testKey(t)
	for _, plaintext := range []string{"", "secret", "with spaces and $pecial % chars", "multi\nline"} {
		token, err := Encrypt(key, plaintext)
		if err != nil {
			t.Fatalf("Encrypt(%q) failed: %v", plaintext, err)
		}
		if !IsEncrypted(token) {
			t.Errorf("token %q missing prefix", token)
		}
		got, err := Decrypt(key, token)
		if err != nil {
			t.Fatalf("Decrypt failed: %v", err)
		}
		if got != plaintext {
			t.Errorf("round trip mismatch: got %q want %q", got, plaintext)
		}
	}
}

func TestTokenIsShellAndIniSafe(t *testing.T) {
	key := testKey(t)
	// Standard base64 could end in "+=" which the INI parser would treat
	// as an append operator. Verify the token alphabet avoids +, / and =.
	for i := 0; i < 50; i++ {
		token, err := Encrypt(key, strings.Repeat("x", i))
		if err != nil {
			t.Fatal(err)
		}
		if strings.ContainsAny(token[len(TokenPrefix):], "+/=") {
			t.Fatalf("token contains unsafe characters: %q", token)
		}
	}
}

func TestDecryptWrongKey(t *testing.T) {
	token, err := Encrypt(testKey(t), "secret")
	if err != nil {
		t.Fatal(err)
	}
	otherKey, _ := DeriveKey("other-passphrase", []byte("0123456789abcdef"), 1000)
	if _, err := Decrypt(otherKey, token); err != ErrDecryptFailed {
		t.Errorf("expected ErrDecryptFailed, got %v", err)
	}
}

func TestDecryptTamperedToken(t *testing.T) {
	key := testKey(t)
	token, err := Encrypt(key, "secret")
	if err != nil {
		t.Fatal(err)
	}
	// Flip a character in the ciphertext part
	raw := []byte(token)
	last := len(raw) - 1
	if raw[last] == 'A' {
		raw[last] = 'B'
	} else {
		raw[last] = 'A'
	}
	if _, err := Decrypt(key, string(raw)); err != ErrDecryptFailed {
		t.Errorf("expected ErrDecryptFailed for tampered token, got %v", err)
	}
}

func TestDecryptMalformed(t *testing.T) {
	key := testKey(t)
	for _, token := range []string{"", "plaintext", "enc:v1:", "enc:v1:!!notbase64!!", "enc:v1:AAAA", "enc:v2:AAAA"} {
		if _, err := Decrypt(key, token); err != ErrMalformedToken {
			t.Errorf("Decrypt(%q): expected ErrMalformedToken, got %v", token, err)
		}
	}
}

func TestIsEncrypted(t *testing.T) {
	if IsEncrypted("plain") || IsEncrypted("") || IsEncrypted("ENC:V1:x") {
		t.Error("false positive")
	}
	if !IsEncrypted("enc:v1:abc") {
		t.Error("false negative")
	}
}

func TestDeriveKeyDeterministic(t *testing.T) {
	salt := []byte("0123456789abcdef")
	k1, _ := DeriveKey("pass", salt, 1000)
	k2, _ := DeriveKey("pass", salt, 1000)
	if !bytes.Equal(k1, k2) {
		t.Error("same inputs must derive the same key")
	}
	k3, _ := DeriveKey("pass", []byte("fedcba9876543210"), 1000)
	if bytes.Equal(k1, k3) {
		t.Error("different salt must derive a different key")
	}
	k4, _ := DeriveKey("other", salt, 1000)
	if bytes.Equal(k1, k4) {
		t.Error("different passphrase must derive a different key")
	}
}

func TestEncodeDecodeKey(t *testing.T) {
	key := testKey(t)
	decoded, err := DecodeKey(EncodeKey(key))
	if err != nil {
		t.Fatalf("DecodeKey failed: %v", err)
	}
	if !bytes.Equal(key, decoded) {
		t.Error("key round trip mismatch")
	}
	for _, bad := range []string{"", "short", "!!!", EncodeKey(key)[:10]} {
		if _, err := DecodeKey(bad); err == nil {
			t.Errorf("DecodeKey(%q) should fail", bad)
		}
	}
}

func TestEncryptInvalidKeySize(t *testing.T) {
	if _, err := Encrypt([]byte("short"), "x"); err != ErrInvalidKey {
		t.Errorf("expected ErrInvalidKey, got %v", err)
	}
	if _, err := Decrypt([]byte("short"), TokenPrefix+"AAAAAAAAAAAAAAAAAAAAAAAAAAAA"); err != ErrInvalidKey {
		t.Errorf("expected ErrInvalidKey, got %v", err)
	}
}
