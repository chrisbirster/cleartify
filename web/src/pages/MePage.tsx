import { useEffect, useState } from "react";
import { Effect } from "effect";

import { PageHeader } from "../components/PageHeader";
import { StatePanel } from "../components/StatePanel";
import { loadMe } from "../lib/api";
import type { MeResponse } from "../lib/types";

export function MePage() {
  const [state, setState] = useState<{ status: "loading" | "error" | "ready"; me?: MeResponse; error?: string }>({
    status: "loading",
  });

  useEffect(() => {
    let active = true;
    Effect.runPromise(loadMe())
      .then((me) => {
        if (active) {
          setState({ status: "ready", me });
        }
      })
      .catch((error) => {
        if (active) {
          setState({ status: "error", error: error instanceof Error ? error.message : "Unable to load your profile." });
        }
      });
    return () => {
      active = false;
    };
  }, []);

  if (state.status === "loading") {
    return <StatePanel surface="page" tone="loading" eyebrow="Profile" title="Loading your profile" description="Gathering your active workspace context." role="status" />;
  }

  if (state.status === "error" || !state.me) {
    return <StatePanel surface="page" eyebrow="Profile" title="Profile unavailable" description={state.error ?? "Unable to load your profile."} role="alert" />;
  }

  const { me } = state;

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow="Profile"
        title={me.user.displayName}
        description="Current-user profile for the active product workspace and onboarding state."
      />
      <section className="grid gap-4 lg:grid-cols-2">
        <article className="rounded-3xl border border-border bg-card/90 p-6 shadow-xl">
          <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Identity</p>
          <h2 className="mt-3 text-xl font-medium text-foreground">{me.user.displayName}</h2>
          <p className="mt-2 text-sm text-muted-foreground">{me.user.email}</p>
        </article>
        <article className="rounded-3xl border border-border bg-card/90 p-6 shadow-xl">
          <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Workspace</p>
          <h2 className="mt-3 text-xl font-medium text-foreground">{me.workspace?.name ?? "Current workspace"}</h2>
          <p className="mt-2 text-sm text-muted-foreground">
            {me.organization?.name ?? "No separate organization context is attached to this session."}
          </p>
        </article>
      </section>
      <article className="rounded-3xl border border-border bg-card/90 p-6 shadow-xl">
        <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Onboarding</p>
        <h2 className="mt-3 text-xl font-medium text-foreground">
          {me.needsOnboarding ? "First-run onboarding still pending" : "Onboarding complete"}
        </h2>
        <p className="mt-2 text-sm leading-6 text-muted-foreground">
          {me.needsOnboarding
            ? "This user still needs the first-run walkthrough before the full product flow should open."
            : "This user has already completed the first-run walkthrough and can move through the full product flow."}
        </p>
      </article>
    </div>
  );
}
