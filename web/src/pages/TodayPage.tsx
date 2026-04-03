import { Link } from "react-router-dom";

import { EmptyState } from "../components/EmptyState";
import { PageHeader } from "../components/PageHeader";
import type { Cluster, Signal, WorkObject, WorkspaceMode } from "../lib/types";

type TodayPageProps = {
  signals: Signal[];
  clusters: Cluster[];
  workObjects: WorkObject[];
  workspaceMode?: WorkspaceMode;
};

export function TodayPage({ signals, clusters, workObjects, workspaceMode }: TodayPageProps) {
  const featuredCluster = clusters[0];
  const featuredWorkObject = workObjects[0];
  const isLiveWorkspace = workspaceMode === "live";

  if (!featuredCluster || !featuredWorkObject) {
    return (
      <EmptyState
        eyebrow={isLiveWorkspace ? "Live workspace" : "Workspace"}
        title={isLiveWorkspace ? "Live workspace ready" : "Workspace ready"}
        description={
          isLiveWorkspace
            ? "This live workspace starts clean so real Signals, Clusters, and Work Objects can arrive without sample hero data."
            : "Choose sample mode to load the hero flow or start a live workspace without demo data."
        }
        action={
          isLiveWorkspace ? (
            <Link className="rounded-full bg-primary px-5 py-2 text-sm font-medium text-primary-foreground" to="/signals/new">
              New Signal
            </Link>
          ) : undefined
        }
      />
    );
  }

  return (
    <div>
      <PageHeader
        title="Today"
        description={
          workspaceMode === "sample"
            ? "Review the sample hero path across the Context Inbox, Work Objects, and release loop."
            : "Review the current workspace across the Context Inbox, Work Objects, and active product execution."
        }
      />
      <div className="grid gap-4 md:grid-cols-3">
        <div className="rounded-3xl border border-border bg-card p-5">
          <p className="text-sm text-muted-foreground">Signals</p>
          <p className="mt-2 text-3xl font-semibold text-foreground">{signals.length}</p>
        </div>
        <div className="rounded-3xl border border-border bg-card p-5">
          <p className="text-sm text-muted-foreground">Clusters</p>
          <p className="mt-2 text-3xl font-semibold text-foreground">{clusters.length}</p>
        </div>
        <div className="rounded-3xl border border-border bg-card p-5">
          <p className="text-sm text-muted-foreground">Work Objects</p>
          <p className="mt-2 text-3xl font-semibold text-foreground">{workObjects.length}</p>
        </div>
      </div>
      <div className="mt-8 grid gap-6 lg:grid-cols-2">
        <section className="rounded-3xl border border-border bg-card p-6">
          <h2 className="text-lg font-medium text-foreground">Featured Cluster</h2>
          <p className="mt-3 text-sm text-muted-foreground">{featuredCluster.summary}</p>
          <Link className="mt-4 inline-flex text-sm font-medium text-primary" to={`/inbox/${featuredCluster.id}`}>
            Open {featuredCluster.title}
          </Link>
        </section>
        <section className="rounded-3xl border border-border bg-card p-6">
          <h2 className="text-lg font-medium text-foreground">Featured Work Object</h2>
          <p className="mt-3 text-sm text-muted-foreground">{featuredWorkObject.summary}</p>
          <Link className="mt-4 inline-flex text-sm font-medium text-primary" to={`/work/${featuredWorkObject.id}`}>
            Open {featuredWorkObject.title}
          </Link>
        </section>
      </div>
    </div>
  );
}
