import { Effect } from "effect";

import type {
  Artifact,
  ArtifactInput,
  AppData,
  AuthSessionResponse,
  Cluster,
  DecisionRecordInput,
  DraftPlan,
  DraftPlanInput,
  Decision,
  DecisionInput,
  MeResponse,
  Signal,
  SignalInput,
  TeamResponse,
  WorkObject,
  WorkObjectCoreFieldsInput,
  WorkObjectPriority,
  WorkObjectPrototypeInput,
  WorkObjectStatus,
} from "./types";

function toError(cause: unknown, fallback: string) {
  return cause instanceof Error ? cause : new Error(fallback);
}

export function jsonFetch<T>(path: string, init?: RequestInit): Effect.Effect<T, Error> {
  return Effect.tryPromise({
    try: async () => {
      const headers = new Headers(init?.headers);
      headers.set("Accept", "application/json");

      const response = await fetch(path, {
        ...init,
        headers,
      });

      if (!response.ok) {
        let message = `Request failed: ${response.status}`;
        try {
          const payload = (await response.json()) as { message?: string };
          if (payload.message) {
            message = payload.message;
          }
        } catch {
          // ignore parse failures and keep the fallback message
        }
        throw new Error(message);
      }

      return (await response.json()) as T;
    },
    catch: (cause) => toError(cause, "Request failed."),
  });
}

export function loadAppData(): Effect.Effect<AppData, Error> {
  return Effect.gen(function* () {
    const session = yield* jsonFetch<AuthSessionResponse>("/api/auth/session");
    if (!session.authenticated) {
      return {
        session,
        signals: [],
        clusters: [],
        workObjects: [],
      };
    }

    const { signals, clusters, workObjects } = yield* Effect.all({
      signals: jsonFetch<Signal[]>("/api/signals"),
      clusters: jsonFetch<Cluster[]>("/api/clusters"),
      workObjects: jsonFetch<WorkObject[]>("/api/work-objects"),
    });

    return {
      session,
      signals,
      clusters,
      workObjects,
    };
  });
}

export function completeOnboarding(): Effect.Effect<AuthSessionResponse, Error> {
  return jsonFetch<AuthSessionResponse>("/api/me/onboarding/complete", {
    method: "POST",
  });
}

export function loadMe(): Effect.Effect<MeResponse, Error> {
  return jsonFetch<MeResponse>("/api/me");
}

export function loadTeam(): Effect.Effect<TeamResponse, Error> {
  return jsonFetch<TeamResponse>("/api/team");
}

export function loadDraftPlanForWorkObject(workObjectId: string): Effect.Effect<DraftPlan | null, Error> {
  return Effect.tryPromise({
    try: async () => {
      const response = await fetch(`/api/work-objects/${encodeURIComponent(workObjectId)}/plan`, {
        headers: {
          Accept: "application/json",
        },
      });

      if (response.status === 404) {
        return null;
      }
      if (!response.ok) {
        throw new Error(`Request failed: ${response.status}`);
      }

      return (await response.json()) as DraftPlan;
    },
    catch: (cause) => toError(cause, "Request failed."),
  });
}

export function generateDraftPlanForWorkObject(workObjectId: string): Effect.Effect<DraftPlan, Error> {
  return jsonFetch<DraftPlan>(`/api/work-objects/${encodeURIComponent(workObjectId)}/plan`, {
    method: "POST",
  });
}

export function updateDraftPlanForWorkObject(workObjectId: string, fields: DraftPlanInput): Effect.Effect<DraftPlan, Error> {
  return jsonFetch<DraftPlan>(`/api/work-objects/${encodeURIComponent(workObjectId)}/plan`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(fields),
  });
}

export function createWorkObjectFromCluster(clusterId: string): Effect.Effect<WorkObject, Error> {
  return jsonFetch<WorkObject>(`/api/clusters/${encodeURIComponent(clusterId)}/create-work-object`, {
    method: "POST",
  });
}

export function resolveCluster(clusterId: string): Effect.Effect<Cluster, Error> {
  return jsonFetch<Cluster>(`/api/clusters/${encodeURIComponent(clusterId)}/resolve`, {
    method: "POST",
  });
}

export function updateWorkObjectStatus(workObjectId: string, status: WorkObjectStatus): Effect.Effect<WorkObject, Error> {
  return jsonFetch<WorkObject>(`/api/work-objects/${encodeURIComponent(workObjectId)}/status`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ status }),
  });
}

export function updateWorkObjectCoreFields(
  workObjectId: string,
  fields: WorkObjectCoreFieldsInput,
): Effect.Effect<WorkObject, Error> {
  return jsonFetch<WorkObject>(`/api/work-objects/${encodeURIComponent(workObjectId)}/core`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(fields),
  });
}

export function updateWorkObjectPrototype(
  workObjectId: string,
  fields: WorkObjectPrototypeInput,
): Effect.Effect<WorkObject, Error> {
  return jsonFetch<WorkObject>(`/api/work-objects/${encodeURIComponent(workObjectId)}/prototype`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(fields),
  });
}

export function updateWorkObjectPriority(workObjectId: string, priority: WorkObjectPriority): Effect.Effect<WorkObject, Error> {
  return jsonFetch<WorkObject>(`/api/work-objects/${encodeURIComponent(workObjectId)}/priority`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ priority }),
  });
}

export function markWorkObjectReadyForRelease(workObjectId: string): Effect.Effect<WorkObject, Error> {
  return jsonFetch<WorkObject>(`/api/work-objects/${encodeURIComponent(workObjectId)}/release-ready`, {
    method: "POST",
  });
}

export function createDecisionForWorkObject(workObjectId: string, fields: DecisionInput): Effect.Effect<Decision, Error> {
  return jsonFetch<Decision>(`/api/work-objects/${encodeURIComponent(workObjectId)}/decisions`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(fields),
  });
}

export function createArtifactForWorkObject(workObjectId: string, fields: ArtifactInput): Effect.Effect<Artifact, Error> {
  return jsonFetch<Artifact>(`/api/work-objects/${encodeURIComponent(workObjectId)}/artifacts`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      type: fields.type,
      title: fields.title,
      body: fields.body,
    }),
  });
}

export function generateImplementationArtifactForWorkObject(workObjectId: string): Effect.Effect<Artifact, Error> {
  return jsonFetch<Artifact>(`/api/work-objects/${encodeURIComponent(workObjectId)}/implementation-artifact`, {
    method: "POST",
  });
}

export function generateTaskBreakdownArtifactForWorkObject(workObjectId: string): Effect.Effect<Artifact, Error> {
  return jsonFetch<Artifact>(`/api/work-objects/${encodeURIComponent(workObjectId)}/task-breakdown-artifact`, {
    method: "POST",
  });
}

export function loadSignals(): Effect.Effect<Signal[], Error> {
  return jsonFetch<Signal[]>("/api/signals");
}

export function loadSignal(signalId: string): Effect.Effect<Signal, Error> {
  return jsonFetch<Signal>(`/api/signals/${encodeURIComponent(signalId)}`);
}

export function createSignal(fields: SignalInput): Effect.Effect<Signal, Error> {
  return jsonFetch<Signal>("/api/signals", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(fields),
  });
}

export function updateSignal(signalId: string, fields: SignalInput): Effect.Effect<Signal, Error> {
  return jsonFetch<Signal>(`/api/signals/${encodeURIComponent(signalId)}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(fields),
  });
}

export function deleteSignal(signalId: string): Effect.Effect<void, Error> {
  return Effect.tryPromise({
    try: async () => {
      const response = await fetch(`/api/signals/${encodeURIComponent(signalId)}`, { method: "DELETE" });
      if (!response.ok) {
        let message = `Request failed: ${response.status}`;
        try {
          const payload = (await response.json()) as { message?: string };
          if (payload.message) {
            message = payload.message;
          }
        } catch {}
        throw new Error(message);
      }
    },
    catch: (cause) => toError(cause, "Request failed."),
  });
}

export function loadDecisions(): Effect.Effect<Decision[], Error> {
  return jsonFetch<Decision[]>("/api/decisions");
}

export function loadDecision(decisionId: string): Effect.Effect<Decision, Error> {
  return jsonFetch<Decision>(`/api/decisions/${encodeURIComponent(decisionId)}`);
}

export function createDecision(fields: DecisionRecordInput): Effect.Effect<Decision, Error> {
  return jsonFetch<Decision>("/api/decisions", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(fields),
  });
}

export function updateDecision(decisionId: string, fields: DecisionRecordInput): Effect.Effect<Decision, Error> {
  return jsonFetch<Decision>(`/api/decisions/${encodeURIComponent(decisionId)}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(fields),
  });
}

export function deleteDecision(decisionId: string): Effect.Effect<void, Error> {
  return Effect.tryPromise({
    try: async () => {
      const response = await fetch(`/api/decisions/${encodeURIComponent(decisionId)}`, { method: "DELETE" });
      if (!response.ok) {
        let message = `Request failed: ${response.status}`;
        try {
          const payload = (await response.json()) as { message?: string };
          if (payload.message) {
            message = payload.message;
          }
        } catch {}
        throw new Error(message);
      }
    },
    catch: (cause) => toError(cause, "Request failed."),
  });
}

export function loadArtifacts(): Effect.Effect<Artifact[], Error> {
  return jsonFetch<Artifact[]>("/api/artifacts");
}

export function loadArtifact(artifactId: string): Effect.Effect<Artifact, Error> {
  return jsonFetch<Artifact>(`/api/artifacts/${encodeURIComponent(artifactId)}`);
}

export function createArtifact(fields: ArtifactInput): Effect.Effect<Artifact, Error> {
  return jsonFetch<Artifact>("/api/artifacts", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(fields),
  });
}

export function updateArtifact(artifactId: string, fields: ArtifactInput): Effect.Effect<Artifact, Error> {
  return jsonFetch<Artifact>(`/api/artifacts/${encodeURIComponent(artifactId)}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(fields),
  });
}

export function deleteArtifact(artifactId: string): Effect.Effect<void, Error> {
  return Effect.tryPromise({
    try: async () => {
      const response = await fetch(`/api/artifacts/${encodeURIComponent(artifactId)}`, { method: "DELETE" });
      if (!response.ok) {
        let message = `Request failed: ${response.status}`;
        try {
          const payload = (await response.json()) as { message?: string };
          if (payload.message) {
            message = payload.message;
          }
        } catch {}
        throw new Error(message);
      }
    },
    catch: (cause) => toError(cause, "Request failed."),
  });
}
