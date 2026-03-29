export type User = {
  id: string;
  email: string;
  displayName: string;
};

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
  status: string;
  priority: string;
  ownerType: string;
  ownerId: string;
  prototypeUrl?: string;
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
