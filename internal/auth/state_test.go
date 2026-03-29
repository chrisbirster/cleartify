package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

type failingStateReader struct{}

func (failingStateReader) Read([]byte) (int, error) {
	return 0, errors.New("entropy failed")
}

func TestSignedStateCodecRoundTrip(t *testing.T) {
	t.Parallel()

	codec := NewSignedStateCodec("state-secret")
	signed, err := codec.Sign(StatePayload{Flow: "signup"})
	if err != nil {
		t.Fatalf("expected state to sign: %v", err)
	}

	payload, err := codec.Verify(signed, time.Now())
	if err != nil {
		t.Fatalf("expected state to verify: %v", err)
	}

	if payload.Flow != "signup" {
		t.Fatalf("expected flow signup, got %q", payload.Flow)
	}
	if payload.Nonce == "" {
		t.Fatal("expected nonce to be populated")
	}
}

func TestSignedStateCodecRejectsInvalidAndExpiredState(t *testing.T) {
	t.Parallel()

	codec := NewSignedStateCodec("state-secret")
	signed, err := codec.Sign(StatePayload{
		Flow:      "login",
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	})
	if err != nil {
		t.Fatalf("expected state to sign: %v", err)
	}

	if _, err := codec.Verify("nope", time.Now()); err == nil {
		t.Fatal("expected malformed state error")
	}

	tampered := signed + "extra"
	if _, err := codec.Verify(tampered, time.Now()); err == nil {
		t.Fatal("expected tampered state error")
	}

	if _, err := codec.Verify(signed, time.Now()); err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected expired state error, got %v", err)
	}
}

func TestSignedStateCodecPreservesProvidedNonceAndExpiry(t *testing.T) {
	t.Parallel()

	expiresAt := time.Now().Add(5 * time.Minute).UTC().Round(time.Second)
	codec := NewSignedStateCodec("state-secret")
	signed, err := codec.Sign(StatePayload{
		Flow:      "login",
		Nonce:     "provided-nonce",
		ExpiresAt: expiresAt,
	})
	if err != nil {
		t.Fatalf("expected state to sign: %v", err)
	}

	payload, err := codec.Verify(signed, time.Now())
	if err != nil {
		t.Fatalf("expected state to verify: %v", err)
	}
	if payload.Nonce != "provided-nonce" {
		t.Fatalf("expected provided nonce, got %q", payload.Nonce)
	}
	if !payload.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("expected provided expiry, got %v", payload.ExpiresAt)
	}
}

func TestSignedStateCodecRejectsDecodeAndJSONErrors(t *testing.T) {
	t.Parallel()

	codec := NewSignedStateCodec("state-secret")
	if _, err := codec.Verify("%%%.abc", time.Now()); err == nil {
		t.Fatal("expected payload decode error")
	}
	if _, err := codec.Verify(base64.RawURLEncoding.EncodeToString([]byte("{}"))+".%%%", time.Now()); err == nil {
		t.Fatal("expected signature decode error")
	}

	payloadRaw := []byte("not-json")
	mac := hmac.New(sha256.New, []byte("state-secret"))
	mac.Write(payloadRaw)
	sig := mac.Sum(nil)
	value := base64.RawURLEncoding.EncodeToString(payloadRaw) + "." + base64.RawURLEncoding.EncodeToString(sig)

	if _, err := codec.Verify(value, time.Now()); err == nil {
		t.Fatal("expected json error")
	}
}

func TestRandomHexFallsBackWhenEntropyFails(t *testing.T) {
	originalReader := stateRandReader
	t.Cleanup(func() { stateRandReader = originalReader })

	stateRandReader = failingStateReader{}

	if value := randomHex(12); value == "" {
		t.Fatal("expected fallback nonce")
	}
}

func TestSignedStateCodecReturnsMarshalError(t *testing.T) {
	originalMarshal := stateJSONMarshal
	t.Cleanup(func() { stateJSONMarshal = originalMarshal })

	stateJSONMarshal = func(any) ([]byte, error) {
		return nil, fmt.Errorf("marshal failed")
	}

	if _, err := NewSignedStateCodec("state-secret").Sign(StatePayload{Flow: "signup"}); err == nil {
		t.Fatal("expected marshal error")
	}
}
