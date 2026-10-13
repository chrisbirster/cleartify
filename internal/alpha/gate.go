package alpha

import (
 "context"
 "database/sql"
 "errors"
 "time"
)
var ErrNotActivated=errors.New("user must complete invited beta activation")
type Gate struct{DB *sql.DB}
func(g Gate) Admit(ctx context.Context,userID string)error{
 tx,err:=g.DB.BeginTx(ctx,nil);if err!=nil{return err}
 defer tx.Rollback()
 var exists int
 err=tx.QueryRowContext(ctx,`SELECT 1 FROM beta_invites b JOIN users u ON u.id=b.activated_user_id WHERE u.id=? AND u.status='active' AND b.status='activated' AND b.activated_at IS NOT NULL LIMIT 1`,userID).Scan(&exists)
 if errors.Is(err,sql.ErrNoRows){return ErrNotActivated};if err!=nil{return err}
 _,err=tx.ExecContext(ctx,`INSERT INTO alpha_memberships(user_id,admitted_by,admitted_at,revoked_at) VALUES(?,'beta-activation-admin',?,NULL) ON CONFLICT(user_id) DO UPDATE SET revoked_at=NULL`,userID,time.Now().UTC().Format(time.RFC3339Nano))
 if err!=nil{return err}
 return tx.Commit()
}
func(g Gate) Eligible(ctx context.Context,userID string)(bool,error){
 var ok int
 err:=g.DB.QueryRowContext(ctx,`SELECT 1 FROM alpha_memberships a JOIN users u ON u.id=a.user_id JOIN beta_invites b ON b.activated_user_id=u.id WHERE u.id=? AND u.status='active' AND a.revoked_at IS NULL AND b.status='activated' AND b.activated_at IS NOT NULL LIMIT 1`,userID).Scan(&ok)
 if errors.Is(err,sql.ErrNoRows){return false,nil}
 return err==nil,err
}
func(g Gate) Revoke(ctx context.Context,userID string)error{
 _,err:=g.DB.ExecContext(ctx,`UPDATE alpha_memberships SET revoked_at=? WHERE user_id=? AND revoked_at IS NULL`,time.Now().UTC().Format(time.RFC3339Nano),userID)
 return err
}
