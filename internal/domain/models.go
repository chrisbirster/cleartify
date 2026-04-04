package domain

import "time"

type User struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

type Organization struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type Workspace struct {
	ID                  string     `json:"id"`
	Slug                string     `json:"slug"`
	Name                string     `json:"name"`
	Mode                string     `json:"mode,omitempty"`
	SampleInitializedAt *time.Time `json:"sampleInitializedAt,omitempty"`
	LiveInitializedAt   *time.Time `json:"liveInitializedAt,omitempty"`
}

type AuthSession struct {
	Authenticated   bool          `json:"authenticated"`
	User            *User         `json:"user,omitempty"`
	Organization    *Organization `json:"organization,omitempty"`
	Workspace       *Workspace    `json:"workspace,omitempty"`
	NeedsOnboarding bool          `json:"needsOnboarding"`
}

type Me struct {
	User            User          `json:"user"`
	Organization    *Organization `json:"organization,omitempty"`
	Workspace       *Workspace    `json:"workspace,omitempty"`
	NeedsOnboarding bool          `json:"needsOnboarding"`
}

type Team struct {
	Workspace    Workspace     `json:"workspace"`
	Organization *Organization `json:"organization,omitempty"`
	CurrentUser  *User         `json:"currentUser,omitempty"`
}

type Project struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspaceId"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Summary     string    `json:"summary"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Signal struct {
	ID         string         `json:"id"`
	ProjectID  string         `json:"projectId,omitempty"`
	SourceType string         `json:"sourceType"`
	SourceID   string         `json:"sourceId"`
	Title      string         `json:"title"`
	Body       string         `json:"body"`
	URL        string         `json:"url,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	CreatedAt  time.Time      `json:"createdAt"`
}

type Cluster struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"projectId,omitempty"`
	Title     string    `json:"title"`
	Summary   string    `json:"summary"`
	Severity  string    `json:"severity"`
	Status    string    `json:"status"`
	Signals   []Signal  `json:"signals"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Decision struct {
	ID           string    `json:"id"`
	WorkObjectID string    `json:"workObjectId"`
	Title        string    `json:"title"`
	Summary      string    `json:"summary"`
	Rationale    string    `json:"rationale"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Artifact struct {
	ID           string         `json:"id"`
	WorkObjectID string         `json:"workObjectId"`
	ArtifactType string         `json:"artifactType"`
	Title        string         `json:"title"`
	Body         string         `json:"body"`
	LinkType     string         `json:"linkType,omitempty"`
	URL          string         `json:"url,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	CreatedAt    time.Time      `json:"createdAt"`
}

type DraftPlan struct {
	WorkObjectID        string    `json:"workObjectId"`
	Overview            string    `json:"overview"`
	ImplementationSteps []string  `json:"implementationSteps"`
	EdgeCases           []string  `json:"edgeCases"`
	OpenQuestions       []string  `json:"openQuestions"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type WorkObject struct {
	ID             string     `json:"id"`
	ProjectID      string     `json:"projectId,omitempty"`
	ClusterID      string     `json:"clusterId,omitempty"`
	Title          string     `json:"title"`
	Summary        string     `json:"summary"`
	Goal           string     `json:"goal"`
	Status         string     `json:"status"`
	Priority       string     `json:"priority"`
	OwnerType      string     `json:"ownerType"`
	OwnerID        string     `json:"ownerId"`
	PrototypeURL   string     `json:"prototypeUrl,omitempty"`
	PrototypeNote  string     `json:"prototypeNote,omitempty"`
	ReleaseReadyAt *time.Time `json:"releaseReadyAt,omitempty"`
	Decisions      []Decision `json:"decisions"`
	Artifacts      []Artifact `json:"artifacts"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}
