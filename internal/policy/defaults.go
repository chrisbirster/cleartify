// Package policy contains owner-approved INTERNAL engineering defaults.
// These are not public promises, legal wording, or externally verified policy.
package policy
import "time"
const (
 ImpersonationDistinctReportHoldThreshold=2
 SpamDistinctReporterReviewThreshold=3
 AbuseDistinctReporterReviewThreshold=3
 ModerationTriageTarget=24*time.Hour
 AppealDecisionTarget=5*24*time.Hour
 AgentOwnershipChallengeTTL=5*time.Minute
 AgentDefaultGrantTTL=time.Hour
 AgentMaxGrantTTL=24*time.Hour
)
const (
 ScopeMessagePreview="message.preview"
 ScopeProtocolPreview="protocol.preview"
 ScopeLocalRead="local.read"
)
func AllowedAgentScope(scope string)bool{
 switch scope{case ScopeMessagePreview,ScopeProtocolPreview,ScopeLocalRead:return true}
 return false
}
// Bans are a manual, reviewed operator decision, never automatic from reports.
const AutomaticAccountBan=false
