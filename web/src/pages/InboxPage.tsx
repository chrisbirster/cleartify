import { useState } from "react";
import { Link } from "react-router-dom";
import { Effect } from "effect";

import { EmptyState } from "../components/EmptyState";
import { PageHeader } from "../components/PageHeader";
import { resolveCluster } from "../lib/api";
import type { Cluster } from "../lib/types";

type InboxPageProps = {
  clusters: Cluster[];
  onDataChanged?: () => Promise<unknown>;
};

export function InboxPage({ clusters, onDataChanged }: InboxPageProps) {
  const [resolving, setResolving] = useState<Record<string, boolean>>({});

  async function handleResolve(clusterId: string) {
    setResolving((current) => ({ ...current, [clusterId]: true }));
    try {
      await Effect.runPromise(resolveCluster(clusterId));
      await onDataChanged?.();
    } finally {
      setResolving((current) => ({ ...current, [clusterId]: false }));
    }
  }

  if (clusters.length === 0) {
    return <EmptyState title="No Clusters yet" description="Seeded Signals will appear here as soon as they are grouped into a Context Inbox." />;
  }

  return (
    <div>
      <PageHeader title="Context Inbox" description="Review seeded clusters built from Signals and product context." />
      <div className="space-y-4">
        {clusters.map((cluster) => (
          <Link
            key={cluster.id}
            to={`/inbox/${cluster.id}`}
            className="group block rounded-3xl border border-border bg-card/90 p-5 shadow-xl transition-all hover:border-primary/30 hover:bg-card"
          >
            <div className="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
              <div className="min-w-0 flex-1">
                <div className="mb-3 flex flex-wrap items-center gap-2">
                  <span className="rounded-full border border-border bg-secondary/70 px-3 py-1 text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">
                    Cluster
                  </span>
                  <span className="rounded-full border border-primary/20 bg-primary/10 px-3 py-1 text-xs font-medium text-primary">
                    {cluster.signals.length} {cluster.signals.length === 1 ? "Signal" : "Signals"}
                  </span>
                </div>
                <h2 className="text-lg font-medium text-foreground">{cluster.title}</h2>
                <p className="mt-2 text-sm text-muted-foreground">{cluster.summary}</p>
                <div className="mt-4 flex flex-wrap gap-2">
                  {cluster.signals.slice(0, 2).map((signal) => (
                    <span
                      key={signal.id}
                      className="rounded-full border border-border bg-background/70 px-3 py-1 text-xs text-muted-foreground"
                    >
                      {signal.sourceType}: {signal.title}
                    </span>
                  ))}
                </div>
              </div>
              <div className="flex shrink-0 flex-col items-start gap-3 lg:items-end">
                <div className="flex flex-wrap gap-2 text-xs">
                  <span className="rounded-full border border-border bg-secondary/70 px-3 py-1 capitalize text-muted-foreground">
                    {cluster.severity}
                  </span>
                  <span className="rounded-full border border-border bg-secondary/70 px-3 py-1 capitalize text-muted-foreground">
                    {cluster.status}
                  </span>
                </div>
                <div className="flex items-center gap-3">
                  <button
                    className="rounded-full border border-border px-3 py-1 text-xs font-medium text-muted-foreground transition hover:border-border/70 hover:text-foreground"
                    disabled={Boolean(resolving[cluster.id])}
                    onClick={(event) => {
                      event.preventDefault();
                      event.stopPropagation();
                      void handleResolve(cluster.id);
                    }}
                    type="button"
                  >
                    {resolving[cluster.id] ? "Resolving…" : "Resolve"}
                  </button>
                  <span className="text-sm font-medium text-primary transition-transform group-hover:translate-x-0.5">
                    Open cluster
                  </span>
                </div>
              </div>
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}
