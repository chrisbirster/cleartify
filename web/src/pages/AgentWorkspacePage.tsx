import { useEffect, useState } from "react";
import { Effect } from "effect";
import { Link, useParams } from "react-router-dom";

import { PageHeader } from "../components/PageHeader";
import { EmptyState } from "../components/EmptyState";
import { generateImplementationArtifactForWorkObject, loadDraftPlanForWorkObject } from "../lib/api";
import type { Artifact, Cluster, DraftPlan, WorkObject } from "../lib/types";

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

export function AgentWorkspacePage({ workObjects, clusters, onWorkObjectUpdated }: AgentWorkspacePageProps) {
  const { workObjectId } = useParams();
  const workObject = workObjects.find((entry) => entry.id === workObjectId);
  const linkedCluster = workObject?.clusterId ? clusters.find((entry) => entry.id === workObject.clusterId) : undefined;
  const [draftPlan, setDraftPlan] = useState<DraftPlan | null>(null);
  const [createdArtifacts, setCreatedArtifacts] = useState<Artifact[]>([]);
  const [isLoadingDraftPlan, setIsLoadingDraftPlan] = useState(true);
  const [isGeneratingArtifact, setIsGeneratingArtifact] = useState(false);
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
    return <EmptyState title="Work Object not found" description="Select a Work Object from the list to run a deterministic mock agent action." />;
  }

  const visibleArtifacts = mergeArtifacts(
    workObject.artifacts.filter((artifact) => artifact.artifactType === "implementation_brief"),
    createdArtifacts,
  );

  const handleGenerateArtifact = () => {
    setIsGeneratingArtifact(true);
    setArtifactError(null);

    Effect.runPromise(generateImplementationArtifactForWorkObject(workObject.id))
      .then((artifact) => {
        const nextVisibleArtifacts = mergeArtifacts(workObject.artifacts, [...createdArtifacts, artifact]);
        setIsGeneratingArtifact(false);
        setArtifactError(null);
        setCreatedArtifacts((current) => mergeArtifacts(current, [artifact]));
        onWorkObjectUpdated?.({
          ...workObject,
          artifacts: nextVisibleArtifacts,
        });
      })
      .catch(() => {
        setIsGeneratingArtifact(false);
        setArtifactError("Unable to generate implementation Artifact.");
      });
  };

  return (
    <div>
      <PageHeader
        title="Agent Workspace"
        description="Run a deterministic mock action that turns the current Work Object and persisted draft plan into a concrete Artifact."
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
                <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Mock agent action</p>
                <h2 className="mt-2 text-xl font-medium text-foreground">Generate implementation Artifact</h2>
                <p className="mt-3 max-w-2xl text-sm leading-6 text-muted-foreground">
                  Use the persisted Work Object and Prototype-to-Plan context to generate a deterministic implementation brief without agent orchestration.
                </p>
              </div>
              <button
                className="inline-flex h-11 items-center justify-center rounded-full border border-primary/20 bg-primary/10 px-5 text-sm font-medium text-primary transition-colors hover:bg-primary/15 disabled:cursor-not-allowed disabled:opacity-60"
                disabled={isGeneratingArtifact || isLoadingDraftPlan || !draftPlan}
                onClick={handleGenerateArtifact}
                type="button"
              >
                {isGeneratingArtifact ? "Generating implementation Artifact..." : "Generate implementation Artifact"}
              </button>
            </div>

            {!draftPlan && !isLoadingDraftPlan && !draftPlanError ? (
              <div className="mt-5 rounded-2xl border border-dashed border-border/80 bg-background/40 px-4 py-5">
                <p className="text-sm font-medium text-foreground">Persist a draft plan first</p>
                <p className="mt-1 text-sm leading-6 text-muted-foreground">
                  This mock agent action depends on the current persisted draft plan before it can create an implementation Artifact.
                </p>
                <Link
                  className="mt-4 inline-flex items-center rounded-full border border-border bg-background/70 px-4 py-2 text-sm font-medium text-foreground transition-colors hover:bg-background"
                  to={`/plans/${workObject.id}`}
                >
                  Open draft plan
                </Link>
              </div>
            ) : null}
          </section>

          <section className="rounded-3xl border border-border bg-card/95 p-6 shadow-xl shadow-black/10">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Generated Artifacts</p>
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
              <div className="mt-5 rounded-2xl border border-dashed border-border/80 bg-background/40 px-4 py-5">
                <p className="text-sm font-medium text-foreground">No generated Artifacts yet</p>
                <p className="mt-1 text-sm leading-6 text-muted-foreground">
                  Run the mock agent action to create a persisted implementation brief from the current product context.
                </p>
              </div>
            ) : (
              <div className="mt-5 space-y-4">
                {visibleArtifacts.map((artifact) => (
                  <article key={artifact.id} className="rounded-2xl border border-border/80 bg-background/55 p-4">
                    <div className="flex flex-wrap items-center justify-between gap-3">
                      <div className="flex flex-wrap items-center gap-2">
                        <h3 className="text-sm font-medium text-foreground">{artifact.title}</h3>
                        <span className="rounded-full border border-primary/20 bg-primary/10 px-2.5 py-1 text-xs font-medium text-primary">
                          {artifact.artifactType}
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
          <section className="rounded-3xl border border-border bg-card/95 p-5 shadow-xl shadow-black/10">
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

          <section className="rounded-3xl border border-border bg-card/95 p-5 shadow-xl shadow-black/10">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Persisted draft plan</p>
                <h2 className="mt-2 text-lg font-medium text-foreground">Current planning context</h2>
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
              <p role="status" className="mt-4 text-sm text-muted-foreground">
                Loading persisted draft plan...
              </p>
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
              <p className="mt-4 text-sm leading-6 text-muted-foreground">
                No persisted draft plan exists yet for this Work Object. Generate the plan first, then return here to create an implementation Artifact.
              </p>
            )}
          </section>
        </aside>
      </div>
    </div>
  );
}
