package atproto

import "testing"

func TestConfigureOAuthURLs(t *testing.T){
 for _,tc:=range []struct{origin,callback string}{
  {"https://api.cleartify.com","https://api.cleartify.com/oauth/callback"},
  {"http://127.0.0.1:8080","http://127.0.0.1:8080/oauth/callback"},
 } {
  got,err:=ConfigureOAuthURLs(tc.origin)
  if err!=nil {t.Fatalf("%q: %v",tc.origin,err)}
  if got.Callback!=tc.callback || got.ClientID!=tc.origin+"/oauth/client-metadata.json" {t.Fatalf("URLs: %+v",got)}
 }
}
func TestConfigureOAuthURLsRejectsRedirectInjection(t *testing.T) {
 for _,raw:=range []string{"","http://evil.example","https://evil.example/path","https://evil.example/?redirect=evil","https://evil.example/#frag","https://a@example.com","https://example.com:443","//example.com","https://example.com.attacker.invalid/path","https://example.com."} {
  if _,err:=ConfigureOAuthURLs(raw);err==nil {t.Errorf("accepted %q",raw)}
 }
}
