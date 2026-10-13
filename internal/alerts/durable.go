package alerts
import(
 "context"
 "encoding/json"
 "errors"
 "os"
 "path/filepath"
 "strings"
 "sync"
 "time"
)
// Rule maps an accumulated metric count to an internal alert key.
type Rule struct{Metric string `json:"metric"`;Threshold uint64 `json:"threshold"`;Key string `json:"key"`}
type Config struct{Enabled bool `json:"enabled"`;Rules []Rule `json:"rules"`}
func(c Config)Validate()error{
 if len(c.Rules)>32{return errors.New("too many rules")}
 for _,r:=range c.Rules{if r.Metric==""||r.Key==""||r.Threshold==0||len(r.Metric)>80||len(r.Key)>80{return errors.New("invalid rule")}}
 return nil
}
// Durable is local, append-only JSONL; it does not emit to any network sink.
type Durable struct{mu sync.Mutex;path string;cfg Config;sink Sink;counts map[string]uint64;now func()time.Time}
func NewDurable(path string,cfg Config,sink Sink)(*Durable,error){
 if err:=cfg.Validate();err!=nil{return nil,err}
 abs,e:=filepath.Abs(path);if e!=nil{return nil,e};tmp,_:=filepath.Abs(os.TempDir())
 rel,e:=filepath.Rel(tmp,abs)
 if e!=nil||rel==".."||strings.HasPrefix(rel,".."+string(os.PathSeparator))||filepath.Ext(abs)!=".jsonl"{return nil,errors.New("durable metrics restricted to temp .jsonl")}
 if sink==nil{sink=Noop{}}
 d:=&Durable{path:abs,cfg:cfg,sink:sink,counts:map[string]uint64{},now:time.Now}
 if !cfg.Enabled{return d,nil}
 raw,e:=os.ReadFile(abs);if errors.Is(e,os.ErrNotExist){return d,nil};if e!=nil{return nil,e}
 for _,line:=range strings.Split(strings.TrimSpace(string(raw)),"\n"){
  if line==""{continue}
  var rec struct{Metric string `json:"metric"`;Value uint64 `json:"value"`}
  if e=json.Unmarshal([]byte(line),&rec);e!=nil||rec.Metric==""{return nil,errors.New("corrupt local metrics log")}
  d.counts[rec.Metric]=rec.Value
 }
 return d,nil
}
func(d *Durable)Increment(ctx context.Context,metric string)error{
 if !d.cfg.Enabled{return nil}
 if metric==""||len(metric)>80{return errors.New("invalid metric")}
 d.mu.Lock();defer d.mu.Unlock()
 if e:=ctx.Err();e!=nil{return e}
 value:=d.counts[metric]+1
 data,e:=json.Marshal(struct{Metric string `json:"metric"`;Value uint64 `json:"value"`}{metric,value});if e!=nil{return e}
 f,e:=os.OpenFile(d.path,os.O_CREATE|os.O_APPEND|os.O_WRONLY,0600);if e!=nil{return e}
 _,e=f.Write(append(data,'\n'));if e==nil{e=f.Sync()}
 closeErr:=f.Close();if e!=nil{return e};if closeErr!=nil{return closeErr}
 d.counts[metric]=value
 for _,rule:=range d.cfg.Rules{if rule.Metric==metric&&value==rule.Threshold{if e=d.sink.Emit(ctx,Event{Key:rule.Key,Stage:metric,At:d.now()});e!=nil{return e}}}
 return nil
}
func(d *Durable)Count(metric string)uint64{d.mu.Lock();defer d.mu.Unlock();return d.counts[metric]}
