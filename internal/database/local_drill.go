package database

import (
 "context"
 "crypto/sha256"
 "database/sql"
 "errors"
 "fmt"
 "io"
 "os"
 "path/filepath"
 "strings"
)

// DrillCopyClosed is deliberately restricted to files under an OS temporary
// directory. Call only after closing all database connections and checkpointing
// local WALs; this is NOT a live or remote database backup facility.
func DrillCopyClosed(src,dst string) error {
 temp,err:=filepath.EvalSymlinks(os.TempDir());if err!=nil{return err}
 valid:=func(path string)bool{
  absolute,e:=filepath.Abs(path);if e!=nil{return false}
  parent,e:=filepath.EvalSymlinks(filepath.Dir(absolute));if e!=nil{return false}
  rel,e:=filepath.Rel(temp,filepath.Join(parent,filepath.Base(absolute)))
  return e==nil&&rel!="."&&!strings.HasPrefix(rel,".."+string(os.PathSeparator))&&rel!=".."&&filepath.Ext(absolute)==".db"
 }
 if !valid(src)||!valid(dst)||src==dst{return errors.New("drill requires distinct .db paths below OS temp directory")}
 if _,err=os.Stat(dst);!errors.Is(err,os.ErrNotExist){return errors.New("restore destination must not exist")}
 in,err:=os.Open(src);if err!=nil{return err};defer in.Close()
 if inStat,e:=in.Stat();e!=nil||!inStat.Mode().IsRegular(){return errors.New("source must be a regular file")}
 for _,suffix:=range []string{"-wal","-shm","-journal"}{if _,e:=os.Stat(src+suffix);e==nil{return errors.New("source has live or uncheckpointed sidecar")}}
 out,err:=os.OpenFile(dst,os.O_CREATE|os.O_EXCL|os.O_WRONLY,0600);if err!=nil{return err}
 good:=false;defer func(){out.Close();if !good{os.Remove(dst)}}()
 h1:=sha256.New();n,e:=io.Copy(io.MultiWriter(out,h1),in);if e!=nil{return e};if n==0{return errors.New("empty database")}
 if e=out.Sync();e!=nil{return e};if e=out.Close();e!=nil{return e}
 f,e:=os.Open(dst);if e!=nil{return e};defer f.Close()
 h2:=sha256.New();if _,e=io.Copy(h2,f);e!=nil{return e}
 if !strings.EqualFold(fmt.Sprintf("%x",h1.Sum(nil)),fmt.Sprintf("%x",h2.Sum(nil))){return errors.New("backup digest mismatch")}
 good=true;return nil
}
func DrillVerify(ctx context.Context,path string)(int,error){
 // caller must use the disposable copy under OS temp only
 temp,_:=filepath.Abs(os.TempDir());abs,e:=filepath.Abs(path)
 if e!=nil{return 0,e};rel,e:=filepath.Rel(temp,abs)
 if e!=nil||rel==".."||strings.HasPrefix(rel,".."+string(os.PathSeparator)){return 0,errors.New("verification restricted to temp")}
 db,e:=sql.Open("turso",abs);if e!=nil{return 0,e};defer db.Close()
 var integrity string
 if e=db.QueryRowContext(ctx,"PRAGMA integrity_check").Scan(&integrity);e!=nil{return 0,e}
 if integrity!="ok"{return 0,errors.New("integrity check failed")}
 var count int
 e=db.QueryRowContext(ctx,"SELECT COUNT(*) FROM waitlist_entries").Scan(&count)
 return count,e
}
