package marketplace
import(
 "context"
 "database/sql"
 "errors"
 "strings"
 "time"
)
var ErrIdempotencyConflict=errors.New("fake payment idempotency key conflict")
// KeyedOperations is an opt-in application-level wrapper around the fake-only
// Service; no real provider can be supplied through this boundary.
type KeyedOperations struct{DB *sql.DB;Enabled bool}
func(k KeyedOperations)check(actor,key,target string)error{
 if !k.Enabled||k.DB==nil||!strings.HasPrefix(actor,"did:")||len(key)<8||len(key)>128||target==""{return ErrDenied}
 return nil
}
func(k KeyedOperations)read(ctx context.Context,actor,key,op,target string)(string,bool,error){
 var oldOp,oldTarget,result string
 e:=k.DB.QueryRowContext(ctx,`SELECT operation,target_id,result_id FROM fake_payment_idempotency WHERE actor_did=? AND request_key=?`,actor,key).Scan(&oldOp,&oldTarget,&result)
 if errors.Is(e,sql.ErrNoRows){return "",false,nil}
 if e!=nil{return "",false,e}
 if oldOp!=op||oldTarget!=target{return "",true,ErrIdempotencyConflict}
 if result==""{return "",true,ErrDenied} // reserved/in-flight: fail closed\n return result,true,nil
}
func(k KeyedOperations)reserve(ctx context.Context,actor,key,op,target string)error{
 _,e:=k.DB.ExecContext(ctx,`INSERT INTO fake_payment_idempotency(actor_did,request_key,operation,target_id,result_id,created_at) VALUES(?,?,?,?,'',?) ON CONFLICT(actor_did,request_key) DO NOTHING`,actor,key,op,target,time.Now().UTC().Format(time.RFC3339Nano))
 if e!=nil{return e}
 var oldOp,oldTarget,result string
 e=k.DB.QueryRowContext(ctx,`SELECT operation,target_id,result_id FROM fake_payment_idempotency WHERE actor_did=? AND request_key=?`,actor,key).Scan(&oldOp,&oldTarget,&result)
 if e!=nil{return e}
 if oldOp!=op||oldTarget!=target{return ErrIdempotencyConflict}
 if result!=""{return ErrDenied}
 return nil
}
func(k KeyedOperations)release(ctx context.Context,actor,key string){
 // Reservation is removed only on failure, never after a committed fake event.
 _,_=k.DB.ExecContext(ctx,`DELETE FROM fake_payment_idempotency WHERE actor_did=? AND request_key=? AND result_id=''`,actor,key)
}
func(k KeyedOperations)save(ctx context.Context,actor,key,op,target,result string)error{
 _,e:=k.DB.ExecContext(ctx,`UPDATE fake_payment_idempotency SET result_id=? WHERE actor_did=? AND request_key=? AND operation=? AND target_id=? AND result_id=''`,actor,key,op,target,result,time.Now().UTC().Format(time.RFC3339Nano))
 if e!=nil{return e}
 existing,_,e:=k.read(ctx,actor,key,op,target)
 if e!=nil{return e};if existing!=result{return ErrIdempotencyConflict}
 return nil
}
// Purchase uses the database's existing unique (listing,buyer) entitlement
// constraint; keyed persistence additionally rejects payload/key reuse.
func(k KeyedOperations)Purchase(ctx context.Context,actor,key,listing string)(string,error){
 if e:=k.check(actor,key,listing);e!=nil{return "",e}
 if existing,found,e:=k.read(ctx,actor,key,"purchase",listing);found||e!=nil{return existing,e}
 svc:=Service{DB:k.DB,Provider:FakeProvider{},Enabled:true}
 id,e:=svc.Purchase(ctx,actor,listing);if e!=nil{k.release(ctx,actor,key);return "",e}
 if e=k.save(ctx,actor,key,"purchase",listing,id);e!=nil{return "",e}
 return id,nil
}
// Refund is idempotent for the same actor/key/entitlement; the core fake
// entitlement transition ensures only one refund event is committed.
func(k KeyedOperations)Refund(ctx context.Context,actor,key,entitlement string)error{
 if e:=k.check(actor,key,entitlement);e!=nil{return e}
 if _,found,e:=k.read(ctx,actor,key,"refund",entitlement);found||e!=nil{return e}
 svc:=Service{DB:k.DB,Provider:FakeProvider{},Enabled:true}
 if e:=svc.Refund(ctx,actor,entitlement);e!=nil{k.release(ctx,actor,key);return e}
 return k.save(ctx,actor,key,"refund",entitlement,entitlement)
}
