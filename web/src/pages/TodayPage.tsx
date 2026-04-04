import { Link } from "react-router-dom";

import { EmptyState } from "../components/EmptyState";
import { PageHeader } from "../components/PageHeader";
import { filterClustersByProject, filterWorkObjectsByProject, useProjectScope } from "../lib/projectScope";
import type { Cluster, Project, Signal, WorkObject, WorkspaceMode } from "../lib/types";

type TodayPageProps = {
  projects?: Project[];
  signals: Signal[];
  clusters: Cluster[];
  workObjects: WorkObject[];
  workspaceMode?: WorkspaceMode;
};

function formatTimestamp(value: string) {
  return new Date(value).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

export function TodayPage({ projects = [], signals, clusters, workObjects, workspaceMode }: TodayPageProps) {
  const { selectedProject, selectedProjectId, linkTo } = useProjectScope(projects);
  const visibleClusters = filterClustersByProject(clusters, selectedProjectId);
  const visibleWorkObjects = filterWorkObjectsByProject(workObjects, selectedProjectId);
  const visibleProjectIds = new Set(visibleWorkObjects.map((workObject) => workObject.projectId).filter(Boolean));
  const releaseReadyCount = visibleWorkObjects.filter((workObject) => Boolean(workObject.releaseReadyAt)).length;
  const activeWorkCount = visibleWorkObjects.filter((workObject) => workObject.status !== "shipped").length;
  const recentArtifacts = [...visibleWorkObjects.flatMap((workObject) => workObject.artifacts)]
    .sort((left, right) => right.createdAt.localeCompare(left.createdAt))
    .slice(0, 4);
  const isLiveWorkspace = workspaceMode === "live";

  if (projects.length === 0) {
    return (
      <EmptyState
        eyebrow={isLiveWorkspace ? "Live workspace" : "Workspace"}
        title={isLiveWorkspace ? "Create your first project" : "Workspace ready"}
        description={
          isLiveWorkspace
            ? "Projects are the new unit of product execution in the live workspace. Create one to start scoping inbox, work, and releases."
            : "The sample workspace is ready for a project-scoped hero flow."
        }
        action={
          <Link className="rounded-full bg-primary px-5 py-2 text-sm font-medium text-primary-foreground" to="/projects/new">
            New Project
          </Link>
        }
      />
    );
  }

  return (
    <div>
      <PageHeader
        title="Today"
        description={
          selectedProject
            ? `Review the current execution state for ${selectedProject.name}.`
            : "Review the workspace across all projects, open clusters, active work, and recent outputs."
        }
      />

      <div className="grid gap-4 md:grid-cols-4">
        <div className="rounded-3xl border border-border bg-card p-5">
          <p className="text-sm text-muted-foreground">Projects</p>
          <p className="mt-2 text-3xl font-semibold text-foreground">{selectedProject ? 1 : projects.length}</p>
        </div>
        <div className="rounded-3xl border border-border bg-card p-5">
          <p className="text-sm text-muted-foreground">Open Clusters</p>
          <p className="mt-2 text-3xl font-semibold text-foreground">{visibleClusters.length}</p>
        </div>
        <div className="rounded-3xl border border-border bg-card p-5">
          <p className="text-sm text-muted-foreground">Active Work</p>
          <p className="mt-2 text-3xl font-semibold text-foreground">{activeWorkCount}</p>
        </div>
        <div className="rounded-3xl border border-border bg-card p-5">
          <p className="text-sm text-muted-foreground">Release-ready</p>
          <p className="mt-2 text-3xl font-semibold text-foreground">{releaseReadyCount}</p>
        </div>
      </div>

      <div className="mt-8 grid gap-6 xl:grid-cols-[minmax(0,1.25fr)_minmax(320px,0.95fr)]">
        <section className="rounded-3xl border border-border bg-card/95 p-6 shadow-xl shadow-black/10">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">
                {selectedProject ? "Selected project" : "All projects"}
              </p>
              <h2 className="mt-2 text-xl font-medium text-foreground">
                {selectedProject ? selectedProject.name : `${projects.length} active projects`}
              </h2>
            </div>
            <Link
              className="inline-flex items-center rounded-full border border-border bg-background/70 px-4 py-2 text-sm font-medium text-foreground transition-colors hover:bg-background"
              to={selectedProject ? `/projects/${selectedProject.id}` : "/projects"}
            >
              {selectedProject ? "Open project dashboard" : "Open Projects"}
            </Link>
          </div>

          {selectedProject ? (
            <div className="mt-5 rounded-2xl border border-border/80 bg-background/55 p-4">
              <p className="text-sm leading-6 text-muted-foreground">{selectedProject.summary}</p>
              <div className="mt-4 flex flex-wrap gap-3">
                <Link className="text-sm font-medium text-primary" to={linkTo("/inbox")}>
                  Context Inbox
                </Link>
                <Link className="text-sm font-medium text-primary" to={linkTo("/work")}>
                  Work Objects
                </Link>
                <Link className="text-sm font-medium text-primary" to={linkTo("/releases")}>
                  Releases
                </Link>
              </div>
            </div>
          ) : (
            <div className="mt-5 space-y-4">
              {projects.map((project) => {
                const projectClusters = clusters.filter((cluster) => cluster.projectId === project.id);
                const projectWorkObjects = workObjects.filter((workObject) => workObject.projectId === project.id);

                return (
                  <Link
                    key={project.id}
                    className="block rounded-2xl border border-border/80 bg-background/55 p-4 transition hover:border-primary/30"
                    to={`/projects/${project.id}`}
                  >
                    <div className="flex flex-wrap items-center justify-between gap-3">
                      <div>
                        <h3 className="text-base font-medium text-foreground">{project.name}</h3>
                        <p className="mt-2 text-sm leading-6 text-muted-foreground">{project.summary}</p>
                      </div>
                      <div className="grid grid-cols-2 gap-3 text-center text-xs text-muted-foreground">
                        <div className="rounded-2xl border border-border bg-card/80 px-4 py-3">
                          <div>Clusters</div>
                          <div className="mt-2 text-base font-medium text-foreground">{projectClusters.length}</div>
                        </div>
                        <div className="rounded-2xl border border-border bg-card/80 px-4 py-3">
                          <div>Work</div>
                          <div className="mt-2 text-base font-medium text-foreground">{projectWorkObjects.length}</div>
                        </div>
                      </div>
                    </div>
                  </Link>
                );
              })}
            </div>
          )}
        </section>

        <section className="rounded-3xl border border-border bg-card/95 p-6 shadow-xl shadow-black/10">
          <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Recent outputs</p>
          <h2 className="mt-2 text-xl font-medium text-foreground">Artifacts and release trail</h2>

          {recentArtifacts.length === 0 ? (
            <div className="mt-5 rounded-2xl border border-border/80 bg-background/55 p-4 text-sm leading-6 text-muted-foreground">
              No Artifacts have been captured for {selectedProject ? selectedProject.name : "the current scope"} yet.
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

          <div className="mt-5 rounded-2xl border border-border/80 bg-background/55 p-4">
            <p className="text-xs font-medium uppercase tracking-[0.16em] text-muted-foreground">Signals in scope</p>
            <p className="mt-2 text-2xl font-semibold text-foreground">
              {selectedProject ? signals.filter((signal) => signal.projectId === selectedProject.id).length : signals.length}
            </p>
            <p className="mt-2 text-sm leading-6 text-muted-foreground">
              {selectedProject
                ? `${selectedProject.name} currently contains ${visibleProjectIds.size} Work Object lanes.`
                : "Use the project selector in the top bar to focus the rest of the app on a single project without losing your place."}
            </p>
          </div>
        </section>
      </div>
    </div>
  );
}
