import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { Effect } from "effect";

import { EmptyState } from "../components/EmptyState";
import { PageHeader } from "../components/PageHeader";
import { StatePanel } from "../components/StatePanel";
import { loadArtifacts } from "../lib/api";
import { formatArtifactLinkType, formatArtifactType, type Artifact, type WorkObject } from "../lib/types";

type ArtifactsPageProps = {
  workObjects: WorkObject[];
};

export function ArtifactsPage({ workObjects }: ArtifactsPageProps) {
  const [state, setState] = useState<{ status: "loading" | "error" | "ready"; artifacts: Artifact[]; error?: string }>({
    status: "loading",
    artifacts: [],
  });

  useEffect(() => {
    let active = true;
    Effect.runPromise(loadArtifacts())
      .then((artifacts) => {
        if (active) {
          setState({ status: "ready", artifacts });
        }
      })
      .catch((error) => {
        if (active) {
          setState({ status: "error", artifacts: [], error: error instanceof Error ? error.message : "Unable to load Artifacts." });
        }
      });
    return () => {
      active = false;
    };
  }, []);

  if (state.status === "loading") {
    return <StatePanel surface="page" tone="loading" eyebrow="Artifacts" title="Loading Artifacts" description="Gathering persisted outputs and external references across the workspace." role="status" />;
  }

  if (state.status === "error") {
    return <StatePanel surface="page" eyebrow="Artifacts" title="Artifacts unavailable" description={state.error ?? "Unable to load Artifacts."} role="alert" />;
  }

  if (state.artifacts.length === 0) {
    return (
      <EmptyState
        title="No Artifacts yet"
        description="Create an Artifact to capture output, notes, or an external reference tied to a Work Object."
        action={
          <Link className="rounded-full bg-primary px-5 py-2 text-sm font-medium text-primary-foreground" to="/artifacts/new">
            New Artifact
          </Link>
        }
      />
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader eyebrow="Artifacts" title="Execution outputs" description="Review and manage Artifacts generated or captured across Work Objects." />
      <div className="flex justify-end">
        <Link className="rounded-full bg-primary px-5 py-2 text-sm font-medium text-primary-foreground" to="/artifacts/new">
          New Artifact
        </Link>
      </div>
      <div className="space-y-4">
        {state.artifacts.map((artifact) => {
          const workObject = workObjects.find((entry) => entry.id === artifact.workObjectId);
          return (
            <Link key={artifact.id} to={`/artifacts/${artifact.id}`} className="block rounded-3xl border border-border bg-card/90 p-5 shadow-xl transition hover:border-primary/30">
              <div className="flex flex-wrap items-center gap-2">
                <span className="rounded-full border border-border bg-secondary/70 px-3 py-1 text-xs font-medium text-muted-foreground">
                  {formatArtifactType(artifact.artifactType)}
                </span>
                <span className="rounded-full border border-border bg-secondary/70 px-3 py-1 text-xs font-medium text-muted-foreground">
                  {workObject?.title ?? "Unlinked Work Object"}
                </span>
                {artifact.linkType ? (
                  <span className="rounded-full border border-primary/20 bg-primary/10 px-3 py-1 text-xs font-medium text-primary">
                    {formatArtifactLinkType(artifact.linkType)}
                  </span>
                ) : null}
              </div>
              <h2 className="mt-3 text-lg font-medium text-foreground">{artifact.title}</h2>
              <p className="mt-2 text-sm text-muted-foreground">{artifact.body}</p>
            </Link>
          );
        })}
      </div>
    </div>
  );
}
