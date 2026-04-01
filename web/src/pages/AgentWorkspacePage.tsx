import { useEffect, useState } from "react";
import { Effect } from "effect";
import { Link, useParams } from "react-router-dom";

import { PageHeader } from "../components/PageHeader";
import { EmptyState } from "../components/EmptyState";
import { StatePanel } from "../components/StatePanel";
import {
  generateImplementationArtifactForWorkObject,
  generateTaskBreakdownArtifactForWorkObject,
  loadDraftPlanForWorkObject,
} from "../lib/api";
import { formatArtifactType, type Artifact, type Cluster, type DraftPlan, type WorkObject } from "../lib/types";

type AgentWorkspacePageProps = {
  workObjects: WorkObject[];
  clusters: Cluster[];
  onWorkObjectUpdated?: (workObject: WorkObject) => void;
};

function formatTimestamp(value: string) {
  return new Date(value).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function countLabel(count: number, singular: string, plural = `${singular}s`) {
  return `${count} ${count === 1 ? singular : plural}`;
}

function mergeArtifacts(artifacts: Artifact[], createdArtifacts: Artifact[]) {
  const seen = new Set(artifacts.map((artifact) => artifact.id));
  const merged = [...artifacts];
  for (const artifact of createdArtifacts) {
    if (seen.has(artifact.id)) {
      continue;
    }
    seen.add(artifact.id);
    merged.push(artifact);
  }
  return merged;
}

const agentArtifactTypes = new Set(["implementation_brief", "task_breakdown"]);

export function AgentWorkspacePage({ workObjects, clusters, onWorkObjectUpdated }: AgentWorkspacePageProps) {
  const { workObjectId } = useParams();
  const workObject = workObjects.find((entry) => entry.id === workObjectId);
  const linkedCluster = workObject?.clusterId ? clusters.find((entry) => entry.id === workObject.clusterId) : undefined;
  const [draftPlan, setDraftPlan] = useState<DraftPlan | null>(null);
  const [createdArtifacts, setCreatedArtifacts] = useState<Artifact[]>([]);
  const [isLoadingDraftPlan, setIsLoadingDraftPlan] = useState(true);
  const [pendingAction, setPendingAction] = useState<"implementation" | "task-breakdown" | null>(null);
  const [draftPlanError, setDraftPlanError] = useState<string | null>(null);
  const [artifactError, setArtifactError] = useState<string | null>(null);

  useEffect(() => {
    setCreatedArtifacts([]);
  }, [workObjectId]);

  useEffect(() => {
    if (!workObject) {
      setDraftPlan(null);
      setIsLoadingDraftPlan(false);
      setDraftPlanError(null);
      return;
    }

    let isActive = true;
    setIsLoadingDraftPlan(true);
    setDraftPlanError(null);

    Effect.runPromise(loadDraftPlanForWorkObject(workObject.id))
      .then((loadedDraftPlan) => {
        if (!isActive) {
          return;
        }
        setDraftPlan(loadedDraftPlan);
        setIsLoadingDraftPlan(false);
      })
      .catch(() => {
        if (!isActive) {
          return;
        }
        setDraftPlan(null);
        setDraftPlanError("Unable to load draft plan.");
        setIsLoadingDraftPlan(false);
      });

    return () => {
      isActive = false;
    };
  }, [workObject]);

  if (!workObject) {
    return (
      <EmptyState
        eyebrow="Agent Workspace"
        title="Work Object not found"
        description="Select a Work Object from Work Objects to run a deterministic Agent Workspace action."
      />
    );
  }

  const visibleArtifacts = mergeArtifacts(
    workObject.artifacts.filter((artifact) => agentArtifactTypes.has(artifact.artifactType)),
    createdArtifacts,
  );

  const runArtifactAction = (
    kind: "implementation" | "task-breakdown",
    effect: Effect.Effect<Artifact, Error>,
    failureMessage: string,
  ) => {
    setPendingAction(kind);
    setArtifactError(null);

    Effect.runPromise(effect)
      .then((artifact) => {
        const nextVisibleArtifacts = mergeArtifacts(workObject.artifacts, [...createdArtifacts, artifact]);
        setPendingAction(null);
        setArtifactError(null);
        setCreatedArtifacts((current) => mergeArtifacts(current, [artifact]));
        onWorkObjectUpdated?.({
          ...workObject,
          artifacts: nextVisibleArtifacts,
        });
      })
      .catch(() => {
        setPendingAction(null);
        setArtifactError(failureMessage);
      });
  };

  const handleGenerateImplementationArtifact = () =>
    runArtifactAction(
      "implementation",
      generateImplementationArtifactForWorkObject(workObject.id),
      "Unable to generate implementation Artifact.",
    );

  const handleGenerateTaskBreakdownArtifact = () =>
    runArtifactAction(
      "task-breakdown",
      generateTaskBreakdownArtifactForWorkObject(workObject.id),
      "Unable to generate task-breakdown Artifact.",
    );

  return (
    <div>
      <PageHeader
        eyebrow="Agent Workspace"
        title="Agent Workspace"
        description="Create deterministic execution artifacts from the current Work Object and persisted draft plan."
      />

      <div className="grid gap-5 xl:grid-cols-[minmax(0,1.35fr)_minmax(300px,0.95fr)]">
        <div className="space-y-5">
          {artifactError ? (
            <div role="alert" className="rounded-2xl border border-rose-400/30 bg-rose-400/10 px-4 py-3 text-sm text-rose-100">
              {artifactError}
            </div>
          ) : null}

          <section className="rounded-3xl border border-border bg-card/95 p-6 shadow-xl shadow-black/10">
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div className="min-w-0">
                <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Deterministic actions</p>
                <h2 className="mt-2 text-xl font-medium text-foreground">Create execution artifacts</h2>
                <p className="mt-3 max-w-2xl text-sm leading-6 text-muted-foreground">
                  Use the current Work Object, Context Inbox link, prototype context, and persisted draft plan to create deterministic outputs.
                </p>
              </div>
            </div>

            <div className="mt-5 grid gap-4 lg:grid-cols-2">
              <article className="rounded-2xl border border-border/80 bg-background/55 p-4">
                <p className="text-xs font-medium uppercase tracking-[0.16em] text-muted-foreground">Implementation brief</p>
                <h3 className="mt-2 text-base font-medium text-foreground">Summarize the next implementation pass</h3>
                <p className="mt-2 text-sm leading-6 text-muted-foreground">
                  Produce an implementation brief grounded in the current Work Object, linked Cluster context, and persisted draft plan.
                </p>
                <button
                  className="mt-4 inline-flex h-11 items-center justify-center rounded-full border border-primary/20 bg-primary/10 px-5 text-sm font-medium text-primary transition-colors hover:bg-primary/15 disabled:cursor-not-allowed disabled:opacity-60"
                  disabled={pendingAction !== null || isLoadingDraftPlan || !draftPlan}
                  onClick={handleGenerateImplementationArtifact}
                  type="button"
                >
                  {pendingAction === "implementation" ? "Generating implementation brief..." : "Generate implementation brief"}
                </button>
              </article>

              <article className="rounded-2xl border border-border/80 bg-background/55 p-4">
                <p className="text-xs font-medium uppercase tracking-[0.16em] text-muted-foreground">Task breakdown</p>
                <h3 className="mt-2 text-base font-medium text-foreground">Break the draft into concrete tasks</h3>
                <p className="mt-2 text-sm leading-6 text-muted-foreground">
                  Produce a deterministic task breakdown from the persisted draft plan sections so the next implementation pass feels concrete.
                </p>
                <button
                  className="mt-4 inline-flex h-11 items-center justify-center rounded-full border border-primary/20 bg-primary/10 px-5 text-sm font-medium text-primary transition-colors hover:bg-primary/15 disabled:cursor-not-allowed disabled:opacity-60"
                  disabled={pendingAction !== null || isLoadingDraftPlan || !draftPlan}
                  onClick={handleGenerateTaskBreakdownArtifact}
                  type="button"
                >
                  {pendingAction === "task-breakdown" ? "Generating task breakdown..." : "Generate task breakdown"}
                </button>
              </article>
            </div>

            {!draftPlan && !isLoadingDraftPlan && !draftPlanError ? (
              <div className="mt-5">
                <StatePanel
                  action={
                    <Link
                      className="inline-flex items-center rounded-full border border-border bg-background/70 px-4 py-2 text-sm font-medium text-foreground transition-colors hover:bg-background"
                      to={`/plans/${workObject.id}`}
                    >
                      Open draft plan
                    </Link>
                  }
                  description="These mock agent actions depend on the current persisted draft plan before they can create persisted Artifacts."
                  eyebrow="Draft plan prerequisite"
                  title="Persist a draft plan first"
                />
              </div>
            ) : null}
          </section>

          <section className="rounded-3xl border border-border bg-card/95 p-6 shadow-xl shadow-black/10">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Agent outputs</p>
                <h2 className="mt-2 text-xl font-medium text-foreground">{countLabel(visibleArtifacts.length, "Artifact")}</h2>
              </div>
              <Link
                className="inline-flex items-center rounded-full border border-border bg-background/70 px-4 py-2 text-sm font-medium text-foreground transition-colors hover:bg-background"
                to={`/work/${workObject.id}`}
              >
                Open Work Object
              </Link>
            </div>

            {visibleArtifacts.length === 0 ? (
              <div className="mt-5">
                <StatePanel
                  description="Run a mock agent action to create a persisted implementation brief or task breakdown from the current product context."
                  eyebrow="Agent output"
                  title="No generated Artifacts yet"
                />
              </div>
            ) : (
              <div className="mt-5 space-y-4">
                {visibleArtifacts.map((artifact) => (
                  <article key={artifact.id} className="rounded-2xl border border-border/80 bg-background/55 p-4">
                    <div className="flex flex-wrap items-center justify-between gap-3">
                      <div className="flex flex-wrap items-center gap-2">
                        <h3 className="text-sm font-medium text-foreground">{artifact.title}</h3>
                        <span className="rounded-full border border-primary/20 bg-primary/10 px-2.5 py-1 text-xs font-medium text-primary">
                          {formatArtifactType(artifact.artifactType)}
                        </span>
                      </div>
                      <p className="text-xs text-muted-foreground">Created {formatTimestamp(artifact.createdAt)}</p>
                    </div>
                    <p className="mt-4 whitespace-pre-line text-sm leading-6 text-foreground/90">{artifact.body}</p>
                  </article>
                ))}
              </div>
            )}
          </section>
        </div>

        <aside className="space-y-5">
          <section className="rounded-3xl border border-border bg-card/95 p-6 shadow-xl shadow-black/10">
            <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Work Object context</p>
            <h2 className="mt-2 text-lg font-medium text-foreground">{workObject.title}</h2>
            <p className="mt-3 text-sm leading-6 text-muted-foreground">{workObject.summary}</p>
            <p className="mt-4 text-sm text-foreground/90">{workObject.goal}</p>
            {linkedCluster ? (
              <div className="mt-4 flex flex-wrap items-center gap-2">
                <span className="rounded-full border border-border bg-background/70 px-3 py-1 text-xs font-medium text-foreground">
                  {linkedCluster.title}
                </span>
                <span className="rounded-full border border-primary/20 bg-primary/10 px-3 py-1 text-xs font-medium text-primary">
                  {linkedCluster.severity} severity
                </span>
              </div>
            ) : null}
          </section>

          <section className="rounded-3xl border border-border bg-card/95 p-6 shadow-xl shadow-black/10">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Prototype-to-Plan draft</p>
                <h2 className="mt-2 text-lg font-medium text-foreground">Current plan context</h2>
              </div>
              <Link
                className="inline-flex items-center rounded-full border border-border bg-background/70 px-4 py-2 text-sm font-medium text-foreground transition-colors hover:bg-background"
                to={`/plans/${workObject.id}`}
              >
                Open draft plan
              </Link>
            </div>

            {draftPlanError ? (
              <div role="alert" className="mt-4 rounded-2xl border border-rose-400/30 bg-rose-400/10 px-4 py-3 text-sm text-rose-100">
                {draftPlanError}
              </div>
            ) : null}

            {isLoadingDraftPlan ? (
              <div className="mt-4">
                <StatePanel
                  description="Loading the persisted draft so these actions stay grounded in the latest plan."
                  eyebrow="Draft plan context"
                  role="status"
                  title="Loading persisted draft plan..."
                  tone="loading"
                />
              </div>
            ) : draftPlan ? (
              <div className="mt-4 space-y-4">
                <p className="text-sm leading-6 text-foreground/90">{draftPlan.overview}</p>
                <div className="grid gap-3 sm:grid-cols-3">
                  <div className="rounded-2xl border border-border/80 bg-background/55 px-4 py-3">
                    <p className="text-xs font-medium uppercase tracking-[0.16em] text-muted-foreground">Implementation</p>
                    <p className="mt-2 text-sm text-foreground">{countLabel(draftPlan.implementationSteps.length, "step")}</p>
                  </div>
                  <div className="rounded-2xl border border-border/80 bg-background/55 px-4 py-3">
                    <p className="text-xs font-medium uppercase tracking-[0.16em] text-muted-foreground">Edge cases</p>
                    <p className="mt-2 text-sm text-foreground">{countLabel(draftPlan.edgeCases.length, "edge case")}</p>
                  </div>
                  <div className="rounded-2xl border border-border/80 bg-background/55 px-4 py-3">
                    <p className="text-xs font-medium uppercase tracking-[0.16em] text-muted-foreground">Open questions</p>
                    <p className="mt-2 text-sm text-foreground">{countLabel(draftPlan.openQuestions.length, "open question")}</p>
                  </div>
                </div>
                <div className="rounded-2xl border border-border/80 bg-background/55 p-4">
                  <p className="text-xs font-medium uppercase tracking-[0.16em] text-muted-foreground">First implementation step</p>
                  <p className="mt-2 text-sm leading-6 text-foreground/90">{draftPlan.implementationSteps[0]}</p>
                </div>
              </div>
            ) : (
              <div className="mt-4">
                <StatePanel
                  description="No persisted draft plan exists yet for this Work Object. Generate the plan first, then return here to create deterministic Artifacts."
                  eyebrow="Draft plan context"
                  title="Persist a draft plan first"
                />
              </div>
            )}
          </section>
        </aside>
      </div>
    </div>
  );
}
