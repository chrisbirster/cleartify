package protocols

import "testing"

func TestGuideStepsOptInAndExecutableRejection(t *testing.T){
 good:=[]byte(`{"steps":[{"operation":"guide","text":"Write down one achievable next step."}]}`)
 if ValidateDraftSpec(good,"CC0-1.0")==nil{t.Fatal("guide enabled by default")}
 if err:=ValidateDraftSpecWithGuides(good,"CC0-1.0");err!=nil{t.Fatal(err)}
 for _,bad:=range []string{
  `{"steps":[{"operation":"guide","text":" "}]}`,
  `{"steps":[{"operation":"guide","text":"A","shell":"echo pwned"}]}`,
  `{"steps":[{"operation":"guide","text":"A","key":"x"}]}`,
  `{"steps":[{"operation":"guide","text":"A","value":"unsafe"}]}`,
  `{"steps":[{"operation":"http.get","text":"A"}]}`,
  `{"steps":[{"operation":"local.set","key":"a","value":"b","text":"oops"}]}`,
  `{"steps":[{"operation":"guide","text":"A"}],"executable":true}`,
 }{
  if err:=ValidateDraftSpecWithGuides([]byte(bad),"CC0-1.0");err==nil{t.Fatalf("accepted %s",bad)}
 }
 many:=`{"steps":[`
 for i:=0;i<33;i++ {if i>0{many+=","};many+=`{"operation":"guide","text":"step"}`}
 many+=`]}`
 if ValidateDraftSpecWithGuides([]byte(many),"CC0-1.0")==nil{t.Fatal("accepted 33 steps")}
}
