import { Link } from "react-router-dom";

import { EmptyState } from "../components/EmptyState";
import { PageHeader } from "../components/PageHeader";
import { filterWorkObjectsByProject, useProjectScope } from "../lib/projectScope";
import { formatWorkObjectPriority, formatWorkObjectStatus, type Cluster, type WorkObject, type WorkspaceMode } from "../lib/types";
import type { Project } from "../lib/types";

type WorkObjectsPageProps = {
  projects?: Project[];
  workObjects: WorkObject[];
  clusters: Cluster[];
  workspaceMode?: WorkspaceMode;
};

function statusTone(status: WorkObject["status"]) {
  switch (status) {
    case "planning":
      return "border-zinc-400/30 bg-zinc-400/16 text-zinc-100";
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
    case "low":
      return "border-slate-300/35 bg-slate-300/16 text-slate-100";
    default:
      return "border-border bg-secondary/70 text-muted-foreground";
  }
}

export function WorkObjectsPage({ projects = [], workObjects, clusters, workspaceMode }: WorkObjectsPageProps) {
  const clusterByID = new Map(clusters.map((cluster) => [cluster.id, cluster]));
  const { selectedProject, selectedProjectId, linkTo } = useProjectScope(projects);
  const visibleWorkObjects = filterWorkObjectsByProject(workObjects, selectedProjectId);

  if (visibleWorkObjects.length === 0) {
    return (
      <EmptyState
        eyebrow="Work Objects"
        title={selectedProject ? "No Work Objects in this Project" : "No Work Objects yet"}
        description={
          selectedProject
            ? `${selectedProject.name} does not have any Work Objects yet.`
            : workspaceMode === "sample"
              ? "The sample Work Object hero path has not been initialized yet."
              : "Live workspaces start clean. A Work Object appears here after product context is triaged from the Context Inbox."
        }
        action={
          <Link className="rounded-full bg-primary px-5 py-2 text-sm font-medium text-primary-foreground" to={linkTo("/inbox")}>
            Open Context Inbox
          </Link>
        }
      />
    );
  }

  return (
    <div>
      <PageHeader
        title="Work Objects"
        description={
          selectedProject
            ? `Review the execution objects scoped to ${selectedProject.name}.`
            : workspaceMode === "sample"
              ? "Review the sample execution objects created from the Cleartify hero context."
              : "Review the execution objects created from current clustered product context."
        }
      />
      <div className="space-y-4">
        {visibleWorkObjects.map((workObject) => {
          const linkedCluster = workObject.clusterId ? clusterByID.get(workObject.clusterId) : undefined;

          return (
            <Link
              key={workObject.id}
              to={linkTo(`/work/${workObject.id}`)}
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
