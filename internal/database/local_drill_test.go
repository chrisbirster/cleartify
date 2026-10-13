package database
import("context";"os";"path/filepath";"testing")
func TestLocalBackupRestoreDrill(t *testing.T){
 ctx:=context.Background();dir:=t.TempDir();src:=filepath.Join(dir,"source.db");backup:=filepath.Join(dir,"backup.db");restored:=filepath.Join(dir,"restored.db")
 store,err:=Open(ctx,Config{Path:src});if err!=nil{t.Fatal(err)}
 if err=store.Migrate(ctx);err!=nil{t.Fatal(err)}
 _,err=store.DB.ExecContext(ctx,`INSERT INTO waitlist_entries(id,name,email,email_normalized,phone_e164,created_at,updated_at) VALUES('w','Test','t@example.test','t@example.test','+17175550111','now','now')`);if err!=nil{t.Fatal(err)}
 if _,err=store.DB.ExecContext(ctx,"PRAGMA wal_checkpoint(TRUNCATE)");err!=nil{t.Fatal(err)}
 if err=store.Close();err!=nil{t.Fatal(err)}
 // The local Turso driver can leave empty WAL/SHM sidecars after close.
 // Remove only zero-byte remnants; nonempty logs must remain a stop condition.
 for _,suffix:=range []string{"-wal","-shm","-journal"}{p:=src+suffix;if info,e:=os.Stat(p);e==nil&&info.Size()==0{if e=os.Remove(p);e!=nil{t.Fatal(e)}}}
 if err=DrillCopyClosed(src,backup);err!=nil{t.Fatal(err)}
 if err=DrillCopyClosed(backup,restored);err!=nil{t.Fatal(err)}
 n,err:=DrillVerify(ctx,restored);if err!=nil||n!=1{t.Fatalf("restored %d %v",n,err)}
 if err=DrillCopyClosed(src,restored);err==nil{t.Fatal("overwrote restored data")}
 if err=DrillCopyClosed("/etc/passwd",filepath.Join(dir,"forbidden.db"));err==nil{t.Fatal("non-temp input accepted")}
 if err=os.WriteFile(src+"-wal",[]byte("uncheckpointed"),0600);err!=nil{t.Fatal(err)}
 if err=DrillCopyClosed(src,filepath.Join(dir,"no-wal.db"));err==nil{t.Fatal("live sidecar accepted")}
}
