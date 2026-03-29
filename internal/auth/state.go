package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

type SignedStateCodec struct {
	secret []byte
}

type StatePayload struct {
	Flow            string    `json:"flow"`
	TargetOrgID     string    `json:"targetOrgId,omitempty"`
	TargetSlug      string    `json:"targetSlug,omitempty"`
	ReturnTo        string    `json:"returnTo,omitempty"`
	PendingSignupID string    `json:"pendingSignupId,omitempty"`
	ExpiresAt       time.Time `json:"expiresAt"`
	Nonce           string    `json:"nonce"`
}

var stateRandReader io.Reader = rand.Reader
var stateJSONMarshal = json.Marshal

func NewSignedStateCodec(secret string) *SignedStateCodec {
	return &SignedStateCodec{secret: []byte(secret)}
}

func (c *SignedStateCodec) Sign(payload StatePayload) (string, error) {
	if payload.Nonce == "" {
		payload.Nonce = randomHex(12)
	}
	if payload.ExpiresAt.IsZero() {
		payload.ExpiresAt = time.Now().Add(10 * time.Minute)
	}

	raw, err := stateJSONMarshal(payload)
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, c.secret)
	mac.Write(raw)
	sig := mac.Sum(nil)

	return base64.RawURLEncoding.EncodeToString(raw) + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func (c *SignedStateCodec) Verify(value string, now time.Time) (StatePayload, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return StatePayload{}, fmt.Errorf("invalid state")
	}

	payloadRaw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return StatePayload{}, err
	}
	sigRaw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return StatePayload{}, err
	}

	mac := hmac.New(sha256.New, c.secret)
	mac.Write(payloadRaw)
	expected := mac.Sum(nil)
	if !hmac.Equal(sigRaw, expected) {
		return StatePayload{}, fmt.Errorf("invalid state signature")
	}

	var payload StatePayload
	if err := json.Unmarshal(payloadRaw, &payload); err != nil {
		return StatePayload{}, err
	}
	if !payload.ExpiresAt.IsZero() && now.After(payload.ExpiresAt) {
		return StatePayload{}, fmt.Errorf("state expired")
	}

	return payload, nil
}

func randomHex(size int) string {
	buf := make([]byte, size)
	if _, err := stateRandReader.Read(buf); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}
