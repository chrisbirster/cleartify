package atproto

import (
 "errors"
 "testing"
)

func TestParseIdentity(t *testing.T) {
 did:="did:plc:ewvi7nxzyoun6zhxrhs64oiz"
 identity,err:=ParseIdentity(did," Alice.Example ")
 if err!=nil { t.Fatal(err) }
 if identity.DID!=did || identity.Handle!="alice.example" {t.Fatalf("unexpected identity: %+v",identity)}
}

func TestRejectInvalidDID(t *testing.T) {
 for _,did:=range []string{"","did:plc:not-valid","did:foo:abc","did:web:example.com/path"} {
  _,err:=ParseIdentity(did,"alice.example")
  if !errors.Is(err,ErrInvalidDID) {t.Errorf("DID %q: got %v",did,err)}
 }
}

func TestRejectInvalidHandle(t *testing.T) {
 for _,handle:=range []string{"","no-dot","-bad.example","bad..example","bad_.example","x."} {
  _,err:=ParseIdentity("did:plc:ewvi7nxzyoun6zhxrhs64oiz",handle)
  if !errors.Is(err,ErrInvalidHandle) {t.Errorf("handle %q: got %v",handle,err)}
 }
}

func TestHandleChangePreservesDID(t *testing.T) {
 did:="did:plc:ewvi7nxzyoun6zhxrhs64oiz"
 before,_:=ParseIdentity(did,"old.example")
 after,_:=ParseIdentity(did,"new.example")
 if before.DID!=after.DID {t.Fatal("handle change changed durable identifier")}
}
