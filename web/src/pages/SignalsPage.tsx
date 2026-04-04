import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { Effect } from "effect";

import { EmptyState } from "../components/EmptyState";
import { PageHeader } from "../components/PageHeader";
import { StatePanel } from "../components/StatePanel";
import { useProjectScope } from "../lib/projectScope";
import { loadSignals } from "../lib/api";
import { formatSignalSourceType, type Project, type Signal } from "../lib/types";

type SignalsPageProps = {
  projects?: Project[];
};

export function SignalsPage({ projects = [] }: SignalsPageProps) {
  const { selectedProject, selectedProjectId, linkTo } = useProjectScope(projects);
  const [state, setState] = useState<{ status: "loading" | "error" | "ready"; signals: Signal[]; error?: string }>({
    status: "loading",
    signals: [],
  });

  useEffect(() => {
    let active = true;
    Effect.runPromise(loadSignals(selectedProjectId))
      .then((signals) => {
        if (active) {
          setState({ status: "ready", signals });
        }
      })
      .catch((error) => {
        if (active) {
          setState({ status: "error", signals: [], error: error instanceof Error ? error.message : "Unable to load Signals." });
        }
      });
    return () => {
      active = false;
    };
  }, [selectedProjectId]);

  if (state.status === "loading") {
    return <StatePanel surface="page" tone="loading" eyebrow="Signals" title="Loading Signals" description="Pulling the current product memory sources into view." role="status" />;
  }

  if (state.status === "error") {
    return <StatePanel surface="page" eyebrow="Signals" title="Signals unavailable" description={state.error ?? "Unable to load Signals."} role="alert" />;
  }

  if (state.signals.length === 0) {
    return (
      <EmptyState
        title="No Signals yet"
        description={
          selectedProject
            ? `Create a Signal for ${selectedProject.name} to capture product context before it becomes part of the inbox and execution flow.`
            : "Create a Signal to capture external product context before it becomes part of the inbox and execution flow."
        }
        action={
          <Link className="rounded-full bg-primary px-5 py-2 text-sm font-medium text-primary-foreground" to={linkTo("/signals/new")}>
            New Signal
          </Link>
        }
      />
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow="Signals"
        title="Product Memory sources"
        description={
          selectedProject
            ? `Review and manage Signals that feed ${selectedProject.name}.`
            : "Review and manage Signals that feed the Context Inbox and broader product memory."
        }
      />
      <div className="flex justify-end">
        <Link className="rounded-full bg-primary px-5 py-2 text-sm font-medium text-primary-foreground" to={linkTo("/signals/new")}>
          New Signal
        </Link>
      </div>
      <div className="space-y-4">
        {state.signals.map((signal) => (
          <Link
            key={signal.id}
            to={linkTo(`/signals/${signal.id}`)}
            className="block rounded-3xl border border-border bg-card/90 p-5 shadow-xl transition hover:border-primary/30"
          >
            <div className="flex flex-wrap items-center gap-2">
              <span className="rounded-full border border-border bg-secondary/70 px-3 py-1 text-xs font-medium capitalize text-muted-foreground">
                {formatSignalSourceType(signal.sourceType)}
              </span>
              {signal.url ? (
                <span className="rounded-full border border-primary/20 bg-primary/10 px-3 py-1 text-xs font-medium text-primary">
                  Linked source
                </span>
              ) : null}
            </div>
            <h2 className="mt-3 text-lg font-medium text-foreground">{signal.title}</h2>
            <p className="mt-2 text-sm leading-6 text-muted-foreground">{signal.body}</p>
          </Link>
        ))}
      </div>
    </div>
  );
}
