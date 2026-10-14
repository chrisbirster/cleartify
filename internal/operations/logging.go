package operations
import("context";"log/slog";"strings")
type RedactingHandler struct{Next slog.Handler}
func(h RedactingHandler)Enabled(ctx context.Context,l slog.Level)bool{return h.Next.Enabled(ctx,l)}
func redact(a slog.Attr)slog.Attr{
 k:=strings.ToLower(a.Key)
 for _,w:=range []string{"secret","token","password","authorization","cookie","key","credential","dsn","url"}{
  if strings.Contains(k,w){return slog.String(a.Key,"[REDACTED]")}
 }
 if a.Value.Kind()==slog.KindGroup{
  attrs:=a.Value.Group();for i:=range attrs{attrs[i]=redact(attrs[i]}
  return slog.Group(a.Key,attrsToArgs(attrs)...)
 }
 // Disallow arbitrary error/text fields: structured logs should contain
 // named event IDs and fixed codes, not error messages with potential secrets.
 if a.Value.Kind()==slog.KindString {
  if k=="error"||k=="message"||k=="detail"||k=="query"{return slog.String(a.Key,"[REDACTED]")}
 }
 return a
}
func attrsToArgs(in []slog.Attr)[]any{out:=make([]any,0,len(in));for _,a:=range in{out=append(out,a)};return out}
func(h RedactingHandler)Handle(ctx context.Context,r slog.Record)error{
 copy:=slog.NewRecord(r.Time,r.Level,r.Message,r.PC)
 r.Attrs(func(a slog.Attr)bool{copy.AddAttrs(redact(a));return true})
 return h.Next.Handle(ctx,copy)
}
func(h RedactingHandler)WithAttrs(attrs []slog.Attr)slog.Handler{
 cleaned:=make([]slog.Attr,len(attrs));for i,a:=range attrs{cleaned[i]=redact(a)}
 return RedactingHandler{Next:h.Next.WithAttrs(cleaned)}
}
func(h RedactingHandler)WithGroup(name string)slog.Handler{return RedactingHandler{Next:h.Next.WithGroup(name)}}
