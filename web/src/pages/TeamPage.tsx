import { useEffect, useState } from "react";
import { Effect } from "effect";

import { PageHeader } from "../components/PageHeader";
import { StatePanel } from "../components/StatePanel";
import { loadTeam } from "../lib/api";
import type { TeamResponse } from "../lib/types";

export function TeamPage() {
  const [state, setState] = useState<{ status: "loading" | "error" | "ready"; team?: TeamResponse; error?: string }>({
    status: "loading",
  });

  useEffect(() => {
    let active = true;
    Effect.runPromise(loadTeam())
      .then((team) => {
        if (active) {
          setState({ status: "ready", team });
        }
      })
      .catch((error) => {
        if (active) {
          setState({ status: "error", error: error instanceof Error ? error.message : "Unable to load the current team." });
        }
      });
    return () => {
      active = false;
    };
  }, []);

  if (state.status === "loading") {
    return <StatePanel surface="page" tone="loading" eyebrow="Team" title="Loading team context" description="Gathering the active workspace and organization context." role="status" />;
  }

  if (state.status === "error" || !state.team) {
    return <StatePanel surface="page" eyebrow="Team" title="Team unavailable" description={state.error ?? "Unable to load the current team."} role="alert" />;
  }

  const { team } = state;

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow="Team"
        title={team.workspace.name}
        description="The current product workspace context attached to this authenticated session."
      />
      <section className="grid gap-4 lg:grid-cols-2">
        <article className="rounded-3xl border border-border bg-card/90 p-6 shadow-xl">
          <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Workspace</p>
          <h2 className="mt-3 text-xl font-medium text-foreground">{team.workspace.name}</h2>
          <p className="mt-2 text-sm text-muted-foreground">Slug: {team.workspace.slug}</p>
        </article>
        <article className="rounded-3xl border border-border bg-card/90 p-6 shadow-xl">
          <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Organization</p>
          <h2 className="mt-3 text-xl font-medium text-foreground">{team.organization?.name ?? "Workspace-owned team"}</h2>
          <p className="mt-2 text-sm text-muted-foreground">
            {team.currentUser ? `Current member: ${team.currentUser.displayName}` : "No active member was returned for this session."}
          </p>
        </article>
      </section>
    </div>
  );
}
