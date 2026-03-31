import { Effect } from "effect";

import type {
  Artifact,
  ArtifactInput,
  AppData,
  AuthSessionResponse,
  Cluster,
  DraftPlan,
  DraftPlanInput,
  Decision,
  DecisionInput,
  Signal,
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
        throw new Error(`Request failed: ${response.status}`);
      }

      return (await response.json()) as T;
    },
    catch: (cause) => toError(cause, "Request failed."),
  });
}

export function loadAppData(): Effect.Effect<AppData, Error> {
  return Effect.all({
    session: jsonFetch<AuthSessionResponse>("/api/auth/session"),
    signals: jsonFetch<Signal[]>("/api/signals"),
    clusters: jsonFetch<Cluster[]>("/api/clusters"),
    workObjects: jsonFetch<WorkObject[]>("/api/work-objects"),
  });
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
    body: JSON.stringify(fields),
  });
}

export function generateImplementationArtifactForWorkObject(workObjectId: string): Effect.Effect<Artifact, Error> {
  return jsonFetch<Artifact>(`/api/work-objects/${encodeURIComponent(workObjectId)}/implementation-artifact`, {
    method: "POST",
  });
}
