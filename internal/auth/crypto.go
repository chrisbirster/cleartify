package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

type TokenCipher struct {
	key []byte
}

var newAESCipher = aes.NewCipher
var newGCMCipher = cipher.NewGCM
var tokenCipherEntropy io.Reader = rand.Reader

func NewTokenCipher(secret string) *TokenCipher {
	sum := sha256.Sum256([]byte(secret))
	return &TokenCipher{key: sum[:]}
}

func (c *TokenCipher) SealString(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := newAESCipher(c.key)
	if err != nil {
		return "", err
	}

	gcm, err := newGCMCipher(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(tokenCipherEntropy, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

func (c *TokenCipher) OpenString(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	raw, err := base64.RawURLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := newAESCipher(c.key)
	if err != nil {
		return "", err
	}

	gcm, err := newGCMCipher(block)
	if err != nil {
		return "", err
	}

	if len(raw) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce := raw[:gcm.NonceSize()]
	payload := raw[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, payload, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
