package interop

import (
 "context"
 "encoding/json"
 "errors"
 "net"
 "net/http"
 "strings"
 "time"
)

// LoopbackHandler accepts only direct loopback TCP peers. It never trusts
// forwarded headers and delegates signed authorization/replay to LocalTransport.
// It is NOT registered on the application's public HTTP mux.
func LoopbackHandler(enabled bool,transport *LocalTransport)http.Handler{
 return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  w.Header().Set("Cache-Control","no-store")
  if !enabled||transport==nil{http.NotFound(w,r);return}
  if r.Method!="POST"||r.URL.Path!="/agent-envelope"{http.NotFound(w,r);return}
  host,_,err:=net.SplitHostPort(r.RemoteAddr)
  ip:=net.ParseIP(host)
  if err!=nil||ip==nil||!ip.IsLoopback(){http.Error(w,"forbidden",http.StatusForbidden);return}
  if r.Header.Get("Content-Type")!="application/json"{http.Error(w,"invalid content type",http.StatusUnsupportedMediaType);return}
  r.Body=http.MaxBytesReader(w,r.Body,16<<10)
  var envelope Envelope
  decoder:=json.NewDecoder(r.Body);decoder.DisallowUnknownFields()
  if err:=decoder.Decode(&envelope);err!=nil{http.Error(w,"invalid envelope",http.StatusBadRequest);return}
  var extra any
  if err:=decoder.Decode(&extra);err==nil{http.Error(w,"invalid envelope",http.StatusBadRequest);return}
  if envelope.MessageID==""||strings.TrimSpace(envelope.MessageID)!=envelope.MessageID{http.Error(w,"invalid envelope",http.StatusBadRequest);return}
  ctx,cancel:=context.WithTimeout(r.Context(),5*time.Second);defer cancel()
  err=transport.Send(ctx,envelope)
  switch {
  case err==nil:w.WriteHeader(http.StatusAccepted)
  case errors.Is(err,ErrReplay):http.Error(w,"replayed",http.StatusConflict)
  case errors.Is(err,ErrTransportDisabled):http.NotFound(w,r)
  case errors.Is(err,ErrTransportDenied),errors.Is(err,ErrInvalidEnvelope):http.Error(w,"denied",http.StatusForbidden)
  default:http.Error(w,"unavailable",http.StatusServiceUnavailable)
  }
 })
}

// ListenLoopback uses explicit 127.0.0.1, disallowing wildcard or LAN bindings.
// Ownership of listener lifetime stays with the local caller/test harness.
func ListenLoopback(enabled bool)(net.Listener,error){
 if !enabled{return nil,ErrTransportDisabled}
 return net.Listen("tcp4","127.0.0.1:0")
}
