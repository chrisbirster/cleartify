import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { Effect } from "effect";

import { EmptyState } from "../components/EmptyState";
import { PageHeader } from "../components/PageHeader";
import { StatePanel } from "../components/StatePanel";
import { filterWorkObjectsByProject, useProjectScope } from "../lib/projectScope";
import { loadDecisions } from "../lib/api";
import type { Decision, Project, WorkObject } from "../lib/types";

type DecisionsPageProps = {
  projects?: Project[];
  workObjects: WorkObject[];
};

export function DecisionsPage({ projects = [], workObjects }: DecisionsPageProps) {
  const { selectedProject, selectedProjectId, linkTo } = useProjectScope(projects);
  const visibleWorkObjects = filterWorkObjectsByProject(workObjects, selectedProjectId);
  const [state, setState] = useState<{ status: "loading" | "error" | "ready"; decisions: Decision[]; error?: string }>({
    status: "loading",
    decisions: [],
  });

  useEffect(() => {
    let active = true;
    Effect.runPromise(loadDecisions(selectedProjectId))
      .then((decisions) => {
        if (active) {
          setState({ status: "ready", decisions });
        }
      })
      .catch((error) => {
        if (active) {
          setState({ status: "error", decisions: [], error: error instanceof Error ? error.message : "Unable to load Decisions." });
        }
      });
    return () => {
      active = false;
    };
  }, [selectedProjectId]);

  if (state.status === "loading") {
    return <StatePanel surface="page" tone="loading" eyebrow="Decisions" title="Loading Decisions" description="Gathering persisted product judgments across the workspace." role="status" />;
  }

  if (state.status === "error") {
    return <StatePanel surface="page" eyebrow="Decisions" title="Decisions unavailable" description={state.error ?? "Unable to load Decisions."} role="alert" />;
  }

  if (state.decisions.length === 0) {
    return (
      <EmptyState
        title="No Decisions yet"
        description={
          selectedProject
            ? `Capture a Decision for ${selectedProject.name} to preserve product judgment next to the work it informs.`
            : "Capture a Decision to preserve product judgment next to the work it informs."
        }
        action={
          <Link className="rounded-full bg-primary px-5 py-2 text-sm font-medium text-primary-foreground" to={linkTo("/decisions/new")}>
            New Decision
          </Link>
        }
      />
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow="Decisions"
        title="Product judgments"
        description={
          selectedProject
            ? `Review persisted Decisions attached to ${selectedProject.name}.`
            : "Review and manage persisted Decisions attached to Work Objects."
        }
      />
      <div className="flex justify-end">
        <Link className="rounded-full bg-primary px-5 py-2 text-sm font-medium text-primary-foreground" to={linkTo("/decisions/new")}>
          New Decision
        </Link>
      </div>
      <div className="space-y-4">
        {state.decisions.map((decision) => {
          const workObject = visibleWorkObjects.find((entry) => entry.id === decision.workObjectId);
          return (
            <Link
              key={decision.id}
              to={linkTo(`/decisions/${decision.id}`)}
              className="block rounded-3xl border border-border bg-card/90 p-5 shadow-xl transition hover:border-primary/30"
            >
              <div className="flex flex-wrap items-center gap-2">
                <span className="rounded-full border border-border bg-secondary/70 px-3 py-1 text-xs font-medium text-muted-foreground">
                  {workObject?.title ?? "Unlinked Work Object"}
                </span>
              </div>
              <h2 className="mt-3 text-lg font-medium text-foreground">{decision.title}</h2>
              <p className="mt-2 text-sm text-muted-foreground">{decision.summary}</p>
            </Link>
          );
        })}
      </div>
    </div>
  );
}
