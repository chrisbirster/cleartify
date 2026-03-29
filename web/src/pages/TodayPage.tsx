import { Link } from "react-router-dom";

import { PageHeader } from "../components/PageHeader";
import type { Cluster, Signal, WorkObject } from "../lib/types";

type TodayPageProps = {
  signals: Signal[];
  clusters: Cluster[];
  workObjects: WorkObject[];
};

export function TodayPage({ signals, clusters, workObjects }: TodayPageProps) {
  const featuredCluster = clusters[0];
  const featuredWorkObject = workObjects[0];

  return (
    <div>
      <PageHeader
        title="Today"
        description="A Phase 0 overview of your Context Inbox, Work Objects, and seeded product activity."
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
