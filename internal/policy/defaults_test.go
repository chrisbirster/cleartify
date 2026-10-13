package policy
import("testing";"time")
func TestOwnerApprovedInternalDefaults(t *testing.T){
 if ImpersonationDistinctReportHoldThreshold!=2||SpamDistinctReporterReviewThreshold!=3||AbuseDistinctReporterReviewThreshold!=3||AutomaticAccountBan{t.Fatal("unsafe automatic moderation defaults")}
 if ModerationTriageTarget!=24*time.Hour||AppealDecisionTarget!=120*time.Hour{t.Fatal("moderation targets drift")}
 if AgentOwnershipChallengeTTL!=5*time.Minute||AgentDefaultGrantTTL!=time.Hour||AgentMaxGrantTTL!=24*time.Hour||AgentDefaultGrantTTL>AgentMaxGrantTTL{t.Fatal("agent proof/grant defaults drift")}
 for _,s:=range []string{ScopeMessagePreview,ScopeProtocolPreview,ScopeLocalRead}{if !AllowedAgentScope(s){t.Fatalf("expected allowed %q",s)}}
 for _,s:=range []string{"","*","payment.charge","network.send","filesystem.write","admin","agent.execute"}{if AllowedAgentScope(s){t.Fatalf("unsafe grant %q",s)}}
}
