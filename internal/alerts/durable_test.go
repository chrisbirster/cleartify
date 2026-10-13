package alerts
import("context";"path/filepath";"testing")
func TestDurableLocalRulesAndFlagOff(t *testing.T){
 ctx:=context.Background();path:=filepath.Join(t.TempDir(),"metrics.jsonl");sink:=NewLocal(5)
 disabled,err:=NewDurable(path,Config{},sink);if err!=nil{t.Fatal(err)}
 if err=disabled.Increment(ctx,"failed");err!=nil||disabled.Count("failed")!=0{t.Fatal("off must be inert",err)}
 cfg:=Config{Enabled:true,Rules:[]Rule{{Metric:"failed",Threshold:2,Key:"threshold_crossed"}}}
 a,err:=NewDurable(path,cfg,sink);if err!=nil{t.Fatal(err)}
 if err=a.Increment(ctx,"failed");err!=nil{t.Fatal(err)}
 b,err:=NewDurable(path,cfg,sink);if err!=nil||b.Count("failed")!=1{t.Fatalf("not durable: %v",err)}
 if err=b.Increment(ctx,"failed");err!=nil{t.Fatal(err)}
 if err=b.Increment(ctx,"failed");err!=nil{t.Fatal(err)}
 events:=sink.Snapshot();if len(events)!=1||events[0].Key!="threshold_crossed"{t.Fatalf("unexpected alerts %+v",events)}
 if _,err=NewDurable(path,Config{Enabled:true,Rules:[]Rule{{Metric:"x",Threshold:0,Key:"bad"}}},sink);err==nil{t.Fatal("invalid threshold")}
 if _,err=NewDurable("/etc/metrics.jsonl",cfg,sink);err==nil{t.Fatal("non-local path")}
}
