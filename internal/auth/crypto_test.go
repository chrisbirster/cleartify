package auth

import (
	stdcipher "crypto/cipher"
	"encoding/base64"
	"errors"
	"testing"
)

type failingNonceReader struct{}

func (failingNonceReader) Read([]byte) (int, error) {
	return 0, errors.New("entropy failed")
}

func TestTokenCipherRoundTrip(t *testing.T) {
	t.Parallel()

	tokenCipher := NewTokenCipher("very-secret")
	sealed, err := tokenCipher.SealString("demo-token")
	if err != nil {
		t.Fatalf("expected token to seal: %v", err)
	}

	opened, err := tokenCipher.OpenString(sealed)
	if err != nil {
		t.Fatalf("expected token to open: %v", err)
	}
	if opened != "demo-token" {
		t.Fatalf("expected demo-token, got %q", opened)
	}
}

func TestTokenCipherHandlesEdgeCases(t *testing.T) {
	t.Parallel()

	tokenCipher := NewTokenCipher("very-secret")
	if sealed, err := tokenCipher.SealString(""); err != nil || sealed != "" {
		t.Fatalf("expected empty plaintext to round trip, got %q / %v", sealed, err)
	}
	if opened, err := tokenCipher.OpenString(""); err != nil || opened != "" {
		t.Fatalf("expected empty token to round trip, got %q / %v", opened, err)
	}
	if _, err := tokenCipher.OpenString("%%%"); err == nil {
		t.Fatal("expected invalid ciphertext error")
	}

	tooShort := base64.RawURLEncoding.EncodeToString([]byte("abc"))
	if _, err := tokenCipher.OpenString(tooShort); err == nil {
		t.Fatal("expected short ciphertext error")
	}

	sealed, err := tokenCipher.SealString("demo-token")
	if err != nil {
		t.Fatalf("expected token to seal: %v", err)
	}
	raw, err := base64.RawURLEncoding.DecodeString(sealed)
	if err != nil {
		t.Fatalf("expected ciphertext to decode: %v", err)
	}
	raw[len(raw)-1] ^= 0xFF
	tampered := base64.RawURLEncoding.EncodeToString(raw)
	if _, err := tokenCipher.OpenString(tampered); err == nil {
		t.Fatal("expected tampered ciphertext error")
	}
}

func TestTokenCipherReturnsInjectedErrors(t *testing.T) {
	tokenCipher := NewTokenCipher("very-secret")

	originalCipherFactory := newAESCipher
	originalGCMFactory := newGCMCipher
	originalEntropy := tokenCipherEntropy
	t.Cleanup(func() {
		newAESCipher = originalCipherFactory
		newGCMCipher = originalGCMFactory
		tokenCipherEntropy = originalEntropy
	})

	newAESCipher = func([]byte) (stdcipher.Block, error) {
		return nil, errors.New("cipher failed")
	}
	if _, err := tokenCipher.SealString("demo-token"); err == nil {
		t.Fatal("expected cipher construction error")
	}
	if _, err := tokenCipher.OpenString(base64.RawURLEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))); err == nil {
		t.Fatal("expected open cipher construction error")
	}

	newAESCipher = originalCipherFactory
	newGCMCipher = func(stdcipher.Block) (stdcipher.AEAD, error) {
		return nil, errors.New("gcm failed")
	}
	if _, err := tokenCipher.SealString("demo-token"); err == nil {
		t.Fatal("expected gcm construction error")
	}
	if _, err := tokenCipher.OpenString(base64.RawURLEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))); err == nil {
		t.Fatal("expected open gcm error")
	}

	newGCMCipher = originalGCMFactory
	tokenCipherEntropy = failingNonceReader{}
	if _, err := tokenCipher.SealString("demo-token"); err == nil {
		t.Fatal("expected nonce read error")
	}
}
