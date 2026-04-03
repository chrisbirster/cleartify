import { Link } from "react-router-dom";

import { EmptyState } from "../components/EmptyState";
import { PageHeader } from "../components/PageHeader";
import { formatWorkObjectPriority, formatWorkObjectStatus, type Cluster, type WorkObject } from "../lib/types";

type ReleasesPageProps = {
  workObjects: WorkObject[];
  clusters: Cluster[];
};

function panelTone(value: string) {
  switch (value) {
    case "planning":
      return "border-zinc-400/30 bg-zinc-400/16 text-zinc-100";
    case "urgent":
      return "border-rose-400/30 bg-rose-400/10 text-rose-100";
    case "high":
    case "in_progress":
      return "border-primary/20 bg-primary/10 text-primary";
    case "ready":
      return "border-sky-400/30 bg-sky-400/10 text-sky-100";
    case "medium":
      return "border-violet-400/20 bg-violet-400/10 text-violet-100";
    case "in_review":
      return "border-amber-400/20 bg-amber-400/10 text-amber-200";
    case "shipped":
      return "border-emerald-400/30 bg-emerald-400/10 text-emerald-200";
    case "low":
      return "border-slate-300/35 bg-slate-300/16 text-slate-100";
    default:
      return "border-border bg-secondary/70 text-muted-foreground";
  }
}

function formatTimestamp(value: string) {
  return new Date(value).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

export function ReleasesPage({ workObjects, clusters }: ReleasesPageProps) {
  const releaseReadyWorkObjects = [...workObjects]
    .filter((workObject) => Boolean(workObject.releaseReadyAt))
    .sort((left, right) => (right.releaseReadyAt ?? "").localeCompare(left.releaseReadyAt ?? ""));

  return (
    <div>
      <PageHeader
        eyebrow="Review and release"
        title="Release queue"
        description="Track the Work Objects that appear here after Review marks them release-ready."
      />

      {releaseReadyWorkObjects.length === 0 ? (
        <EmptyState
          eyebrow="Release queue"
          title="No release-ready Work Objects yet"
          description="A Work Object appears here after the Review checkpoint marks it release-ready."
        />
      ) : (
        <div className="grid gap-4">
          {releaseReadyWorkObjects.map((workObject) => {
            const linkedCluster = workObject.clusterId ? clusters.find((cluster) => cluster.id === workObject.clusterId) : undefined;

            return (
              <article key={workObject.id} className="rounded-3xl border border-border bg-card/95 p-5 shadow-xl shadow-black/10">
                <div className="flex flex-wrap items-start justify-between gap-4">
                  <div className="min-w-0">
                    <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Release-ready</p>
                    <h2 className="mt-2 text-xl font-medium text-foreground">{workObject.title}</h2>
                    <p className="mt-3 max-w-3xl text-sm leading-6 text-muted-foreground">{workObject.summary}</p>
                  </div>
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="rounded-full border border-emerald-400/30 bg-emerald-400/10 px-3 py-1 text-xs font-medium text-emerald-200">
                      Release-ready on {formatTimestamp(workObject.releaseReadyAt ?? "")}
                    </span>
                    <span className={`rounded-full border px-3 py-1 text-xs font-medium ${panelTone(workObject.status)}`}>{formatWorkObjectStatus(workObject.status)}</span>
                    <span className={`rounded-full border px-3 py-1 text-xs font-medium ${panelTone(workObject.priority)}`}>{formatWorkObjectPriority(workObject.priority)} priority</span>
                  </div>
                </div>

                <div className="mt-4 flex flex-wrap items-center gap-2">
                  {linkedCluster ? (
                    <>
                      <span className="rounded-full border border-border bg-background/70 px-3 py-1 text-xs font-medium text-foreground">
                        {linkedCluster.title}
                      </span>
                      <span className={`rounded-full border px-3 py-1 text-xs font-medium ${panelTone(linkedCluster.severity)}`}>
                        {linkedCluster.severity} severity
                      </span>
                    </>
                  ) : (
                    <span className="rounded-full border border-border bg-background/70 px-3 py-1 text-xs font-medium text-muted-foreground">
                      No linked Cluster
                    </span>
                  )}
                </div>

                <div className="mt-4 flex flex-wrap items-center gap-4 text-sm text-muted-foreground">
                  <span>{workObject.decisions.length} Decisions</span>
                  <span>{workObject.artifacts.length} Artifacts</span>
                </div>

                <div className="mt-5 flex flex-wrap gap-3">
                  <Link
                    className="inline-flex items-center rounded-full border border-border bg-background/70 px-4 py-2 text-sm font-medium text-foreground transition-colors hover:bg-background"
                    to={`/review/${workObject.id}`}
                  >
                    Open review
                  </Link>
                  <Link
                    className="inline-flex items-center rounded-full border border-border bg-background/70 px-4 py-2 text-sm font-medium text-foreground transition-colors hover:bg-background"
                    to={`/work/${workObject.id}`}
                  >
                    Open Work Object
                  </Link>
                </div>
              </article>
            );
          })}
        </div>
      )}
    </div>
  );
}
