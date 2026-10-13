// Package atproto contains identities and boundaries for optional public
// AT Protocol accounts. It does not implement OAuth or establish ownership.
package atproto

import (
 "errors"
 "regexp"
 "strings"
)

var (
 ErrInvalidDID = errors.New("invalid AT Protocol DID")
 ErrInvalidHandle = errors.New("invalid AT Protocol handle")
 didPLC = regexp.MustCompile(`^did:plc:[a-z2-7]{24}$`)
 didWeb = regexp.MustCompile(`^did:web:[A-Za-z0-9._:%-]+$`)
 handlePattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]*[a-z0-9])?)+$`)
)

// Identity identifies an AT Protocol account for public social features.
// DID is the durable key; Handle is a mutable presentation/discovery value.
// Validation is syntactic only; callers MUST verify via a trusted OAuth flow.
type Identity struct {
 DID string
 Handle string
}

// ParseIdentity rejects malformed identifier input without consulting a PDS.
// Never use this function as proof of account ownership.
func ParseIdentity(did, handle string) (Identity, error) {
 if len(did) > 2048 || (!didPLC.MatchString(did) && !didWeb.MatchString(did)) {
  return Identity{}, ErrInvalidDID
 }
 handle = strings.ToLower(strings.TrimSpace(handle))
 if len(handle) > 253 || !handlePattern.MatchString(handle) {
  return Identity{}, ErrInvalidHandle
 }
 for _, label := range strings.Split(handle, ".") {
  if len(label) > 63 { return Identity{}, ErrInvalidHandle }
 }
 return Identity{DID:did, Handle:handle},nil
}
