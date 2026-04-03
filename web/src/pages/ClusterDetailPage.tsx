import { useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { Effect } from "effect";

import { EmptyState } from "../components/EmptyState";
import { PageHeader } from "../components/PageHeader";
import { createWorkObjectFromCluster, resolveCluster } from "../lib/api";
import type { Cluster, WorkObject } from "../lib/types";

type ClusterDetailPageProps = {
  clusters: Cluster[];
  createWorkObject?: (clusterId: string) => Effect.Effect<WorkObject, Error>;
  onCreateSuccess?: (workObject: WorkObject) => void;
  onDataChanged?: () => Promise<unknown>;
};

export function redirectToWorkObject(workObject: WorkObject) {
  browserRedirect.assign(`/work/${workObject.id}`);
}

export const browserRedirect = {
  assign(path: string) {
    window.location.assign(path);
  },
};

export function ClusterDetailPage({
  clusters,
  createWorkObject = createWorkObjectFromCluster,
  onCreateSuccess = redirectToWorkObject,
  onDataChanged,
}: ClusterDetailPageProps) {
  const navigate = useNavigate();
  const { clusterId } = useParams();
  const cluster = clusters.find((entry) => entry.id === clusterId);
  const [isCreating, setIsCreating] = useState(false);
  const [isResolving, setIsResolving] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);
  const [resolveError, setResolveError] = useState<string | null>(null);

  if (!cluster) {
    return <EmptyState title="Cluster not found" description="Select a Cluster from the Context Inbox to inspect it." />;
  }
  const currentCluster = cluster;

  async function handleCreateWorkObject() {
    setIsCreating(true);
    setCreateError(null);

    try {
      const workObject = await Effect.runPromise(createWorkObject(currentCluster.id));
      onCreateSuccess(workObject);
    } catch {
      setCreateError("Unable to create Work Object.");
      setIsCreating(false);
    }
  }

  async function handleResolveCluster() {
    setIsResolving(true);
    setResolveError(null);

    try {
      await Effect.runPromise(resolveCluster(currentCluster.id));
      await onDataChanged?.();
      navigate("/inbox");
    } catch {
      setResolveError("Unable to resolve Cluster.");
      setIsResolving(false);
    }
  }

  return (
    <div>
      <PageHeader title={cluster.title} description={cluster.summary} />
      <div className="mb-6 flex flex-col gap-4 rounded-3xl border border-border bg-card/90 p-5 shadow-xl lg:flex-row lg:items-center lg:justify-between">
        <div className="min-w-0">
          <Link className="text-sm font-medium text-primary" to="/inbox">
            Back to Context Inbox
          </Link>
          <div className="mt-4 flex flex-wrap gap-2 text-xs">
            <span className="rounded-full border border-border bg-secondary/70 px-3 py-1 capitalize text-muted-foreground">
              {cluster.severity}
            </span>
            <span className="rounded-full border border-border bg-secondary/70 px-3 py-1 capitalize text-muted-foreground">
              {cluster.status}
            </span>
            <span className="rounded-full border border-primary/20 bg-primary/10 px-3 py-1 font-medium text-primary">
              {cluster.signals.length} {cluster.signals.length === 1 ? "Signal" : "Signals"}
            </span>
          </div>
        </div>
        <div className="flex flex-col items-start gap-3 lg:items-end">
          <button
            className="rounded-full bg-primary px-5 py-2.5 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-60"
            disabled={isCreating}
            onClick={handleCreateWorkObject}
            type="button"
          >
            {isCreating ? "Creating Work Object..." : "Create Work Object"}
          </button>
          <button
            className="rounded-full border border-border px-5 py-2.5 text-sm font-medium text-foreground transition hover:border-border/70 hover:bg-card/70 disabled:cursor-not-allowed disabled:opacity-60"
            disabled={isResolving}
            onClick={handleResolveCluster}
            type="button"
          >
            {isResolving ? "Resolving Cluster..." : "Resolve Cluster"}
          </button>
          {createError ? <p className="text-sm text-red-300">{createError}</p> : null}
          {resolveError ? <p className="text-sm text-red-300">{resolveError}</p> : null}
        </div>
      </div>

      <div className="space-y-4">
        {cluster.signals.map((signal) => (
          <article key={signal.id} className="rounded-3xl border border-border bg-card/90 p-5 shadow-xl">
            <div className="flex flex-wrap items-center gap-2">
              <span className="rounded-full border border-border bg-secondary/70 px-3 py-1 text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">
                {signal.sourceType}
              </span>
              {signal.url ? (
                <a
                  className="text-xs font-medium text-primary"
                  href={signal.url}
                  rel="noreferrer"
                  target="_blank"
                >
                  Open source
                </a>
              ) : null}
            </div>
            <h2 className="mt-2 text-lg font-medium text-foreground">{signal.title}</h2>
            <p className="mt-2 text-sm leading-6 text-muted-foreground">{signal.body}</p>
          </article>
        ))}
      </div>
    </div>
  );
}
