import { EmptyState } from "../components/EmptyState";
import { PageHeader } from "../components/PageHeader";
import { filterWorkObjectsByProject, useProjectScope } from "../lib/projectScope";
import type { Project, WorkObject } from "../lib/types";
import { Link } from "react-router-dom";

type PrototypesPageProps = {
  projects?: Project[];
  workObjects: WorkObject[];
};

export function PrototypesPage({ projects = [], workObjects }: PrototypesPageProps) {
  const { selectedProject, selectedProjectId, linkTo } = useProjectScope(projects);
  const attachedWorkObjects = filterWorkObjectsByProject(workObjects, selectedProjectId).filter(
    (workObject) => workObject.prototypeUrl || workObject.prototypeNote,
  );

  return (
    <div>
      <PageHeader
        title="Prototypes"
        description={
          selectedProject
            ? `Review prototype links and notes attached to ${selectedProject.name}.`
            : "Review the prototype links and notes attached to current Work Objects."
        }
      />
      {attachedWorkObjects.length === 0 ? (
        <EmptyState
          title="No prototypes attached"
          description={
            selectedProject
              ? `Attach a prototype link or note to a Work Object in ${selectedProject.name} to start the Prototype-to-Plan loop.`
              : "Attach a prototype link or note from a Work Object to start the Prototype-to-Plan loop."
          }
        />
      ) : (
        <div className="space-y-4">
          {attachedWorkObjects.map((workObject) => (
            <article key={workObject.id} className="rounded-3xl border border-border bg-card p-5 shadow-xl">
              <div className="flex flex-wrap items-start justify-between gap-4">
                <div className="min-w-0">
                  <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Work Object</p>
                  <h2 className="mt-2 text-lg font-medium text-foreground">{workObject.title}</h2>
                  <p className="mt-2 text-sm leading-6 text-muted-foreground">{workObject.summary}</p>
                </div>
                <Link
                  className="rounded-full border border-primary/20 bg-primary/10 px-4 py-2 text-sm font-medium text-primary transition-colors hover:bg-primary/15"
                  to={linkTo(`/work/${workObject.id}`)}
                >
                  Open Work Object
                </Link>
              </div>

              <div className="mt-5 grid gap-4 lg:grid-cols-2">
                <div className="rounded-2xl border border-border/80 bg-background/50 p-4">
                  <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Prototype link</p>
                  {workObject.prototypeUrl ? (
                    <a
                      className="mt-3 inline-flex items-center text-sm font-medium text-primary"
                      href={workObject.prototypeUrl}
                      rel="noreferrer"
                      target="_blank"
                    >
                      Open prototype
                    </a>
                  ) : (
                    <p className="mt-3 text-sm text-muted-foreground">No prototype link attached yet.</p>
                  )}
                </div>

                <div className="rounded-2xl border border-border/80 bg-background/50 p-4">
                  <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Prototype note</p>
                  <p className="mt-3 text-sm leading-6 text-muted-foreground">
                    {workObject.prototypeNote || "No prototype note attached yet."}
                  </p>
                </div>
              </div>
            </article>
          ))}
        </div>
      )}
    </div>
  );
}
