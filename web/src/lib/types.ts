export type User = {
  id: string;
  email: string;
  displayName: string;
};

export const WORK_OBJECT_STATUSES = ["planning", "ready", "in_progress", "in_review", "shipped"] as const;
export const WORK_OBJECT_PRIORITIES = ["low", "medium", "high", "urgent"] as const;

export type WorkObjectStatus = (typeof WORK_OBJECT_STATUSES)[number];
export type WorkObjectPriority = (typeof WORK_OBJECT_PRIORITIES)[number];
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
  type: string;
  title: string;
  body: string;
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

export function formatClusterStatus(status: string) {
  return status.replaceAll("_", " ");
}

export type Organization = {
  id: string;
  slug: string;
  name: string;
};

export type AuthSessionResponse = {
  authenticated: boolean;
  user?: User;
  organization?: Organization;
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
