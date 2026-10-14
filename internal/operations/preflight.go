package operations
import("errors";"net/url";"strconv";"strings")
var ErrPreflight=errors.New("preflight validation failed")
// Check only returns configuration NAMES and generic error keys; it never
// echoes secrets, URLs, hostnames, port values, or raw environment data.
type Finding struct{Name,Problem string}
func Preflight(env map[string]string)[]Finding{
 out:=[]Finding{}
 val:=func(k string)string{return strings.TrimSpace(env[k])}
 required:=func(k string){if val(k)==""{out=append(out,Finding{k,"missing"})}}
 required("APP_SECRET")
 required("APP_ORIGIN")
 required("DATABASE_PATH")
 if val("APP_SECRET")!=""&&len(val("APP_SECRET"))<32{out=append(out,Finding{"APP_SECRET","invalid_length"})}
 if raw:=val("APP_ORIGIN");raw!=""{
  u,e:=url.Parse(raw);if e!=nil||u.Host==""||u.User!=nil||u.RawQuery!=""||u.Fragment!=""||u.Path!=""||(u.Scheme!="https"&&!(u.Scheme=="http"&&(u.Hostname()=="localhost"||u.Hostname()=="127.0.0.1"))){out=append(out,Finding{"APP_ORIGIN","invalid_origin"})}
 }
 if v:=val("PORT");v!=""{n,e:=strconv.Atoi(v);if e!=nil||n<1||n>65535{out=append(out,Finding{"PORT","invalid_port"})}}
 if val("ATPROTO_OAUTH_ENABLED")=="true"{
  for _,k:=range []string{"ATPROTO_PUBLIC_ORIGIN","ATPROTO_SESSION_KEY_B64"}{required(k)}
 }
 if val("TURSO_REMOTE_URL")!=""{required("TURSO_AUTH_TOKEN")}
 if val("NOTIFICATION_MODE")=="debug"{out=append(out,Finding{"NOTIFICATION_MODE","unsafe_debug_mode"})}
 return out
}
