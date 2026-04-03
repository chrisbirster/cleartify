export type User = {
  id: string;
  email: string;
  displayName: string;
};

export type WorkspaceMode = "sample" | "live";

export const SIGNAL_SOURCE_TYPES = [
  "feedback",
  "github",
  "google_doc",
  "jira",
  "linear",
  "note",
  "notion",
  "prototype",
  "release",
  "slack",
] as const;

export const ARTIFACT_LINK_TYPES = [
  "confluence",
  "figma",
  "github_issue",
  "github_pr",
  "google_doc",
  "generic",
  "notion",
  "slack_thread",
] as const;

export const WORK_OBJECT_STATUSES = ["planning", "ready", "in_progress", "in_review", "shipped"] as const;
export const WORK_OBJECT_PRIORITIES = ["low", "medium", "high", "urgent"] as const;

export type WorkObjectStatus = (typeof WORK_OBJECT_STATUSES)[number];
export type WorkObjectPriority = (typeof WORK_OBJECT_PRIORITIES)[number];
export type SignalSourceType = (typeof SIGNAL_SOURCE_TYPES)[number];
export type ArtifactLinkType = (typeof ARTIFACT_LINK_TYPES)[number];
export type WorkObjectCoreFieldsInput = {
  title: string;
  summary: string;
  goal: string;
};
export type WorkObjectPrototypeInput = {
  prototypeUrl: string;
  prototypeNote: string;
};
export type DecisionInput = {
  title: string;
  summary: string;
  rationale: string;
};
export type ArtifactInput = {
  workObjectId?: string;
  type: string;
  title: string;
  body: string;
  linkType?: ArtifactLinkType | "";
  url?: string;
};
export type SignalInput = {
  sourceType: SignalSourceType;
  sourceId: string;
  title: string;
  body: string;
  url?: string;
};
export type DecisionRecordInput = {
  workObjectId: string;
  title: string;
  summary: string;
  rationale: string;
};

export type DraftPlan = {
  workObjectId: string;
  overview: string;
  implementationSteps: string[];
  edgeCases: string[];
  openQuestions: string[];
  createdAt: string;
  updatedAt: string;
};

export type DraftPlanInput = {
  overview: string;
  implementationSteps: string[];
  edgeCases: string[];
  openQuestions: string[];
};

export function formatWorkObjectStatus(status: string) {
  return status.replaceAll("_", " ");
}

export function formatWorkObjectPriority(priority: string) {
  return priority;
}

export function formatArtifactType(artifactType: string) {
  return artifactType.replaceAll("_", " ");
}

export function formatSignalSourceType(sourceType: string) {
  return sourceType.replaceAll("_", " ");
}

export function formatArtifactLinkType(linkType: string) {
  return linkType.replaceAll("_", " ");
}

export function formatClusterStatus(status: string) {
  return status.replaceAll("_", " ");
}

export type Organization = {
  id: string;
  slug: string;
  name: string;
};

export type Workspace = {
  id: string;
  slug: string;
  name: string;
  mode?: WorkspaceMode;
  sampleInitializedAt?: string;
  liveInitializedAt?: string;
};

export type MeResponse = {
  user: User;
  organization?: Organization;
  workspace?: Workspace;
  needsOnboarding: boolean;
};

export type TeamResponse = {
  workspace: Workspace;
  organization?: Organization;
  currentUser?: User;
};

export type AuthSessionResponse = {
  authenticated: boolean;
  user?: User;
  organization?: Organization;
  workspace?: Workspace;
  needsOnboarding: boolean;
};

export type Signal = {
  id: string;
  sourceType: string;
  sourceId: string;
  title: string;
  body: string;
  url?: string;
  metadata?: Record<string, unknown>;
  createdAt: string;
};

export type Cluster = {
  id: string;
  title: string;
  summary: string;
  severity: string;
  status: string;
  signals: Signal[];
  createdAt: string;
  updatedAt: string;
};

export type Decision = {
  id: string;
  workObjectId: string;
  title: string;
  summary: string;
  rationale: string;
  createdAt: string;
};

export type Artifact = {
  id: string;
  workObjectId: string;
  artifactType: string;
  title: string;
  body: string;
  linkType?: string;
  url?: string;
  metadata?: Record<string, unknown>;
  createdAt: string;
};

export type WorkObject = {
  id: string;
  clusterId?: string;
  title: string;
  summary: string;
  goal: string;
  status: WorkObjectStatus;
  priority: WorkObjectPriority;
  ownerType: string;
  ownerId: string;
  prototypeUrl?: string;
  prototypeNote?: string;
  releaseReadyAt?: string;
  decisions: Decision[];
  artifacts: Artifact[];
  createdAt: string;
  updatedAt: string;
};

export type AppData = {
  session: AuthSessionResponse;
  signals: Signal[];
  clusters: Cluster[];
  workObjects: WorkObject[];
};
