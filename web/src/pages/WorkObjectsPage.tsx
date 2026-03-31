import { Link } from "react-router-dom";

import { PageHeader } from "../components/PageHeader";
import { formatWorkObjectPriority, formatWorkObjectStatus, type Cluster, type WorkObject } from "../lib/types";

type WorkObjectsPageProps = {
  workObjects: WorkObject[];
  clusters: Cluster[];
};

function statusTone(status: WorkObject["status"]) {
  switch (status) {
    case "ready":
      return "border-sky-400/30 bg-sky-400/10 text-sky-100";
    case "in_progress":
      return "border-primary/20 bg-primary/10 text-primary";
    case "in_review":
      return "border-amber-400/30 bg-amber-400/10 text-amber-200";
    case "shipped":
      return "border-emerald-400/30 bg-emerald-400/10 text-emerald-200";
    default:
      return "border-border bg-secondary/70 text-muted-foreground";
  }
}

function priorityTone(priority: WorkObject["priority"]) {
  switch (priority) {
    case "urgent":
      return "border-rose-400/30 bg-rose-400/10 text-rose-100";
    case "high":
      return "border-primary/20 bg-primary/10 text-primary";
    case "medium":
      return "border-violet-400/20 bg-violet-400/10 text-violet-100";
    default:
      return "border-slate-400/20 bg-slate-400/10 text-slate-200";
  }
}

export function WorkObjectsPage({ workObjects, clusters }: WorkObjectsPageProps) {
  const clusterByID = new Map(clusters.map((cluster) => [cluster.id, cluster]));

  return (
    <div>
      <PageHeader title="Work Objects" description="Review seeded execution objects created from clustered context." />
      <div className="space-y-4">
        {workObjects.map((workObject) => {
          const linkedCluster = workObject.clusterId ? clusterByID.get(workObject.clusterId) : undefined;

          return (
            <Link
              key={workObject.id}
              to={`/work/${workObject.id}`}
              className="block rounded-3xl border border-border bg-card p-5 transition-colors hover:border-primary/40"
            >
              <div className="flex items-start justify-between gap-4">
                <div className="min-w-0">
                  <h2 className="text-lg font-medium text-foreground">{workObject.title}</h2>
                  <p className="mt-2 text-sm text-muted-foreground">{workObject.summary}</p>
                  <div className="mt-4 flex flex-wrap gap-2 text-xs font-medium">
                    <span className={`rounded-full border px-3 py-1 capitalize ${statusTone(workObject.status)}`}>
                      {formatWorkObjectStatus(workObject.status)}
                    </span>
                    <span className={`rounded-full border px-3 py-1 capitalize ${priorityTone(workObject.priority)}`}>
                      {formatWorkObjectPriority(workObject.priority)} priority
                    </span>
                    {linkedCluster ? (
                      <>
                        <span className="rounded-full border border-border bg-background/60 px-3 py-1 text-muted-foreground">
                          {linkedCluster.title}
                        </span>
                        <span className={`rounded-full border px-3 py-1 capitalize ${priorityTone(linkedCluster.severity === "high" ? "high" : linkedCluster.severity === "low" ? "low" : "medium")}`}>
                          {linkedCluster.severity} severity
                        </span>
                      </>
                    ) : null}
                  </div>
                </div>
                <div className="rounded-2xl border border-border/80 bg-background/60 px-4 py-3 text-right text-xs text-muted-foreground">
                  <div>Owner</div>
                  <div className="mt-1 font-medium text-foreground/90">{workObject.ownerId}</div>
                </div>
              </div>
            </Link>
          );
        })}
      </div>
    </div>
  );
}
