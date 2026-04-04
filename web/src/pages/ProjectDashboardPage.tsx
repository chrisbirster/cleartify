import { Link, useParams } from "react-router-dom";

import { EmptyState } from "../components/EmptyState";
import { PageHeader } from "../components/PageHeader";
import { filterClustersByProject, filterSignalsByProject, filterWorkObjectsByProject } from "../lib/projectScope";
import type { Cluster, Project, Signal, WorkObject } from "../lib/types";

type ProjectDashboardPageProps = {
  projects?: Project[];
  signals: Signal[];
  clusters: Cluster[];
  workObjects: WorkObject[];
};

function formatTimestamp(value: string) {
  return new Date(value).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

export function ProjectDashboardPage({ projects = [], signals, clusters, workObjects }: ProjectDashboardPageProps) {
  const { projectId } = useParams();
  const project = projects.find((entry) => entry.id === projectId);

  if (!project) {
    return (
      <EmptyState
        eyebrow="Projects"
        title="Project not found"
        description="Choose a Project from the Projects list to open its dashboard."
      />
    );
  }

  const projectSignals = filterSignalsByProject(signals, project.id);
  const projectClusters = filterClustersByProject(clusters, project.id);
  const projectWorkObjects = filterWorkObjectsByProject(workObjects, project.id);
  const releaseReadyWorkObjects = projectWorkObjects.filter((workObject) => Boolean(workObject.releaseReadyAt));
  const recentArtifacts = [...projectWorkObjects.flatMap((workObject) => workObject.artifacts)]
    .sort((left, right) => right.createdAt.localeCompare(left.createdAt))
    .slice(0, 4);

  return (
    <div>
      <PageHeader eyebrow="Project dashboard" title={project.name} description={project.summary} />

      <div className="grid gap-4 md:grid-cols-4">
        <div className="rounded-3xl border border-border bg-card p-5">
          <p className="text-sm text-muted-foreground">Signals</p>
          <p className="mt-2 text-3xl font-semibold text-foreground">{projectSignals.length}</p>
        </div>
        <div className="rounded-3xl border border-border bg-card p-5">
          <p className="text-sm text-muted-foreground">Open Clusters</p>
          <p className="mt-2 text-3xl font-semibold text-foreground">{projectClusters.length}</p>
        </div>
        <div className="rounded-3xl border border-border bg-card p-5">
          <p className="text-sm text-muted-foreground">Active Work</p>
          <p className="mt-2 text-3xl font-semibold text-foreground">
            {projectWorkObjects.filter((workObject) => workObject.status !== "shipped").length}
          </p>
        </div>
        <div className="rounded-3xl border border-border bg-card p-5">
          <p className="text-sm text-muted-foreground">Release-ready</p>
          <p className="mt-2 text-3xl font-semibold text-foreground">{releaseReadyWorkObjects.length}</p>
        </div>
      </div>

      <div className="mt-8 grid gap-6 xl:grid-cols-[minmax(0,1.25fr)_minmax(320px,0.95fr)]">
        <section className="rounded-3xl border border-border bg-card/95 p-6 shadow-xl shadow-black/10">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Execution lanes</p>
              <h2 className="mt-2 text-xl font-medium text-foreground">Project scope</h2>
            </div>
            <div className="flex flex-wrap gap-3">
              <Link
                className="inline-flex items-center rounded-full border border-border bg-background/70 px-4 py-2 text-sm font-medium text-foreground transition-colors hover:bg-background"
                to={`/inbox?projectId=${encodeURIComponent(project.id)}`}
              >
                Open Context Inbox
              </Link>
              <Link
                className="inline-flex items-center rounded-full border border-border bg-background/70 px-4 py-2 text-sm font-medium text-foreground transition-colors hover:bg-background"
                to={`/work?projectId=${encodeURIComponent(project.id)}`}
              >
                Open Work Objects
              </Link>
            </div>
          </div>

          <div className="mt-5 space-y-4">
            {projectWorkObjects.length === 0 ? (
              <div className="rounded-2xl border border-border/80 bg-background/55 p-4 text-sm leading-6 text-muted-foreground">
                This project is ready for real product context. As Signals and Clusters accumulate, its execution lanes will appear here.
              </div>
            ) : (
              projectWorkObjects.slice(0, 3).map((workObject) => (
                <Link
                  key={workObject.id}
                  className="block rounded-2xl border border-border/80 bg-background/55 p-4 transition hover:border-primary/30"
                  to={`/work/${workObject.id}?projectId=${encodeURIComponent(project.id)}`}
                >
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="rounded-full border border-primary/20 bg-primary/10 px-3 py-1 text-xs font-medium text-primary">
                      {workObject.status}
                    </span>
                    <span className="rounded-full border border-border bg-secondary/70 px-3 py-1 text-xs font-medium text-muted-foreground">
                      {workObject.priority} priority
                    </span>
                  </div>
                  <h3 className="mt-3 text-base font-medium text-foreground">{workObject.title}</h3>
                  <p className="mt-2 text-sm leading-6 text-muted-foreground">{workObject.summary}</p>
                </Link>
              ))
            )}
          </div>
        </section>

        <section className="rounded-3xl border border-border bg-card/95 p-6 shadow-xl shadow-black/10">
          <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Recent outputs</p>
          <h2 className="mt-2 text-xl font-medium text-foreground">Artifacts and review trail</h2>

          {recentArtifacts.length === 0 ? (
            <div className="mt-5 rounded-2xl border border-border/80 bg-background/55 p-4 text-sm leading-6 text-muted-foreground">
              No Artifacts have been captured for this project yet.
            </div>
          ) : (
            <div className="mt-5 space-y-3">
              {recentArtifacts.map((artifact) => (
                <article key={artifact.id} className="rounded-2xl border border-border/80 bg-background/55 p-4">
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <div>
                      <h3 className="text-sm font-medium text-foreground">{artifact.title}</h3>
                      <p className="mt-1 text-xs text-muted-foreground">{artifact.artifactType.replaceAll("_", " ")}</p>
                    </div>
                    <span className="text-xs text-muted-foreground">{formatTimestamp(artifact.createdAt)}</span>
                  </div>
                  <p className="mt-3 text-sm leading-6 text-muted-foreground">{artifact.body}</p>
                </article>
              ))}
            </div>
          )}

          <div className="mt-5">
            <Link
              className="inline-flex items-center rounded-full border border-border bg-background/70 px-4 py-2 text-sm font-medium text-foreground transition-colors hover:bg-background"
              to={`/releases?projectId=${encodeURIComponent(project.id)}`}
            >
              Open Releases
            </Link>
          </div>
        </section>
      </div>
    </div>
  );
}
