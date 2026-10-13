package atproto

import (
 "errors"
 "net/url"
 "strings"
)

var ErrInvalidOAuthOrigin = errors.New("invalid AT Protocol OAuth public origin")

// OAuthURLs derives both OAuth endpoints from a single trusted operator-set
// origin. Never accept callback or post-login redirect URLs from a user request.
type OAuthURLs struct {
 ClientID string
 Callback string
}

// ConfigureOAuthURLs deliberately requires the public origin to be an origin,
// not an arbitrary full URL. HTTPS is mandatory except literal loopback dev.
func ConfigureOAuthURLs(raw string) (OAuthURLs,error) {
 u,err:=url.Parse(raw)
 if err!=nil || u.Host=="" || u.User!=nil || u.Opaque!="" || u.RawQuery!="" || u.Fragment!="" || (u.Path!="" && u.Path!="/") {return OAuthURLs{},ErrInvalidOAuthOrigin}
 if strings.Contains(u.Host,"@") || strings.HasSuffix(u.Host,".") {return OAuthURLs{},ErrInvalidOAuthOrigin}
 loopback:=u.Hostname()=="127.0.0.1" || u.Hostname()=="localhost"
 if u.Scheme!="https" && !(u.Scheme=="http" && loopback) {return OAuthURLs{},ErrInvalidOAuthOrigin}
 if !loopback && (u.Hostname()=="" || u.Port()!="") {return OAuthURLs{},ErrInvalidOAuthOrigin}
 origin:=u.Scheme+"://"+u.Host
 return OAuthURLs{ClientID:origin+"/oauth/client-metadata.json",Callback:origin+"/oauth/callback"},nil
}
