package api
import(
 "errors"
 "net"
 "strings"
)
var ErrInvalidTrustedProxyCIDRs=errors.New("invalid trusted proxy CIDRs")
func ParseTrustedProxyCIDRs(raw string)([]*net.IPNet,error){
 if strings.TrimSpace(raw)==""{return nil,ErrInvalidTrustedProxyCIDRs}
 parts:=strings.Split(raw,",");if len(parts)>32{return nil,ErrInvalidTrustedProxyCIDRs}
 out:=make([]*net.IPNet,0,len(parts))
 for _,part:=range parts{
  _,cidr,err:=net.ParseCIDR(strings.TrimSpace(part))
  if err!=nil{return nil,ErrInvalidTrustedProxyCIDRs}
  out=append(out,cidr)
 }
 return out,nil
}
