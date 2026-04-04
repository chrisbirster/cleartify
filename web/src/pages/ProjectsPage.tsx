import { Link } from "react-router-dom";

import { EmptyState } from "../components/EmptyState";
import { PageHeader } from "../components/PageHeader";
import type { Cluster, Project, WorkObject, WorkspaceMode } from "../lib/types";

type ProjectsPageProps = {
  projects?: Project[];
  clusters: Cluster[];
  workObjects: WorkObject[];
  workspaceMode?: WorkspaceMode;
};

export function ProjectsPage({ projects = [], clusters, workObjects, workspaceMode }: ProjectsPageProps) {
  if (projects.length === 0) {
    return (
      <EmptyState
        eyebrow="Projects"
        title="No Projects yet"
        description={
          workspaceMode === "live"
            ? "Create a Project to give the workspace a clear execution boundary before product context starts accumulating."
            : "The sample workspace project has not been initialized yet."
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
        eyebrow="Projects"
        title="Projects"
        description="Organize the workspace into clear project lanes so inbox, execution, and release views no longer feel like one flat bucket."
      />
      <div className="mb-6 flex justify-end">
        <Link className="rounded-full bg-primary px-5 py-2 text-sm font-medium text-primary-foreground" to="/projects/new">
          New Project
        </Link>
      </div>
      <div className="space-y-4">
        {projects.map((project) => {
          const projectClusters = clusters.filter((cluster) => cluster.projectId === project.id);
          const projectWorkObjects = workObjects.filter((workObject) => workObject.projectId === project.id);
          const releaseReadyCount = projectWorkObjects.filter((workObject) => Boolean(workObject.releaseReadyAt)).length;

          return (
            <Link
              key={project.id}
              to={`/projects/${project.id}`}
              className="block rounded-3xl border border-border bg-card/90 p-5 shadow-xl transition hover:border-primary/30"
            >
              <div className="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
                <div className="min-w-0">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="rounded-full border border-border bg-secondary/70 px-3 py-1 text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">
                      Project
                    </span>
                    <span className="rounded-full border border-primary/20 bg-primary/10 px-3 py-1 text-xs font-medium text-primary">
                      {project.status}
                    </span>
                  </div>
                  <h2 className="mt-3 text-xl font-medium text-foreground">{project.name}</h2>
                  <p className="mt-3 max-w-3xl text-sm leading-6 text-muted-foreground">{project.summary}</p>
                </div>
                <div className="grid grid-cols-3 gap-3 text-center text-xs text-muted-foreground">
                  <div className="rounded-2xl border border-border/80 bg-background/50 px-4 py-3">
                    <div>Clusters</div>
                    <div className="mt-2 text-base font-medium text-foreground">{projectClusters.length}</div>
                  </div>
                  <div className="rounded-2xl border border-border/80 bg-background/50 px-4 py-3">
                    <div>Work</div>
                    <div className="mt-2 text-base font-medium text-foreground">{projectWorkObjects.length}</div>
                  </div>
                  <div className="rounded-2xl border border-border/80 bg-background/50 px-4 py-3">
                    <div>Release-ready</div>
                    <div className="mt-2 text-base font-medium text-foreground">{releaseReadyCount}</div>
                  </div>
                </div>
              </div>
            </Link>
          );
        })}
      </div>
    </div>
  );
}
