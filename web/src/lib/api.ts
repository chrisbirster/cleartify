import { Effect } from "effect";

import type { AppData, AuthSessionResponse, Cluster, Signal, WorkObject } from "./types";

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

export function createWorkObjectFromCluster(clusterId: string): Effect.Effect<WorkObject, Error> {
  return jsonFetch<WorkObject>(`/api/clusters/${encodeURIComponent(clusterId)}/create-work-object`, {
    method: "POST",
  });
}
