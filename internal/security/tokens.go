package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
)

type Tokens struct {
	key []byte
}

func NewTokens(secret string) (*Tokens, error) {
	if len(secret) < 32 {
		return nil, errors.New("APP_SECRET must be at least 32 characters")
	}
	return &Tokens{key: []byte(secret)}, nil
}

func (t *Tokens) RandomURLToken(bytes int) (string, error) {
	if bytes < 16 {
		return "", errors.New("token size must be at least 16 bytes")
	}
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate random token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func (t *Tokens) FiveDigitCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(90000))
	if err != nil {
		return "", fmt.Errorf("generate challenge code: %w", err)
	}
	return fmt.Sprintf("%05d", n.Int64()+10000), nil
}

func (t *Tokens) Digest(value string) string {
	mac := hmac.New(sha256.New, t.key)
	_, _ = mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

func (t *Tokens) Matches(value, digest string) bool {
	got, err := hex.DecodeString(t.Digest(value))
	if err != nil {
		return false
	}
	want, err := hex.DecodeString(digest)
	if err != nil {
		return false
	}
	return hmac.Equal(got, want)
}
