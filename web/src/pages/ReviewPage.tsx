import { useEffect, useState } from "react";
import { Effect } from "effect";
import { Link, useParams } from "react-router-dom";

import { EmptyState } from "../components/EmptyState";
import { PageHeader } from "../components/PageHeader";
import { StatePanel } from "../components/StatePanel";
import { useProjectScope } from "../lib/projectScope";
import { loadDraftPlanForWorkObject, markWorkObjectReadyForRelease } from "../lib/api";
import {
  formatArtifactType,
  formatClusterStatus,
  formatWorkObjectPriority,
  formatWorkObjectStatus,
  type Cluster,
  type DraftPlan,
  type Project,
  type WorkObject,
} from "../lib/types";

type ReviewPageProps = {
  projects?: Project[];
  workObjects: WorkObject[];
  clusters: Cluster[];
  onWorkObjectUpdated?: (workObject: WorkObject) => void;
};

function panelTone(value: string) {
  switch (value) {
    case "planning":
      return "border-zinc-400/30 bg-zinc-400/16 text-zinc-100";
    case "urgent":
      return "border-rose-400/30 bg-rose-400/10 text-rose-100";
    case "high":
    case "in_progress":
      return "border-primary/20 bg-primary/10 text-primary";
    case "ready":
      return "border-sky-400/30 bg-sky-400/10 text-sky-100";
    case "medium":
      return "border-violet-400/20 bg-violet-400/10 text-violet-100";
    case "in_review":
      return "border-amber-400/20 bg-amber-400/10 text-amber-200";
    case "shipped":
      return "border-emerald-400/30 bg-emerald-400/10 text-emerald-200";
    case "low":
      return "border-slate-300/35 bg-slate-300/16 text-slate-100";
    default:
      return "border-border bg-secondary/70 text-muted-foreground";
  }
}

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

export function ReviewPage({ projects = [], workObjects, clusters, onWorkObjectUpdated }: ReviewPageProps) {
  const { workObjectId } = useParams();
  const { linkTo } = useProjectScope(projects);
  const workObject = workObjects.find((entry) => entry.id === workObjectId);
  const [draftPlan, setDraftPlan] = useState<DraftPlan | null>(null);
  const [draftPlanError, setDraftPlanError] = useState<string | null>(null);
  const [reviewError, setReviewError] = useState<string | null>(null);
  const [isLoadingDraftPlan, setIsLoadingDraftPlan] = useState(true);
  const [isMarkingReady, setIsMarkingReady] = useState(false);
  const [localWorkObject, setLocalWorkObject] = useState<WorkObject | null>(null);

  useEffect(() => {
    setLocalWorkObject(null);
    setReviewError(null);
  }, [workObjectId]);

  const visibleWorkObject = localWorkObject ?? workObject;
  const linkedCluster = visibleWorkObject?.clusterId ? clusters.find((entry) => entry.id === visibleWorkObject.clusterId) : undefined;
  const linkedProject = visibleWorkObject?.projectId ? projects.find((entry) => entry.id === visibleWorkObject.projectId) : undefined;

  useEffect(() => {
    if (!visibleWorkObject) {
      setDraftPlan(null);
      setDraftPlanError(null);
      setIsLoadingDraftPlan(false);
      return;
    }

    let isActive = true;
    setIsLoadingDraftPlan(true);
    setDraftPlanError(null);

    Effect.runPromise(loadDraftPlanForWorkObject(visibleWorkObject.id))
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
        setDraftPlanError("Unable to load persisted draft plan.");
        setIsLoadingDraftPlan(false);
      });

    return () => {
      isActive = false;
    };
  }, [visibleWorkObject]);

  if (!visibleWorkObject) {
    return (
      <EmptyState
        eyebrow="Review and release"
        title="Work Object not found"
        description="Select a Work Object from Work Objects to review it before release."
      />
    );
  }

  const handleMarkReadyForRelease = () => {
    setIsMarkingReady(true);
    setReviewError(null);

    Effect.runPromise(markWorkObjectReadyForRelease(visibleWorkObject.id))
      .then((updatedWorkObject) => {
        setIsMarkingReady(false);
        setReviewError(null);
        setLocalWorkObject(updatedWorkObject);
        onWorkObjectUpdated?.(updatedWorkObject);
      })
      .catch(() => {
        setIsMarkingReady(false);
        setReviewError("Unable to mark the Work Object ready for release.");
      });
  };

  return (
    <div>
      <PageHeader
        eyebrow={linkedProject ? `Review and release • ${linkedProject.name}` : "Review and release"}
        title="Review"
        description="Inspect the final execution context before this Work Object is marked release-ready and sent into the release queue."
      />

      <div className="grid gap-5 xl:grid-cols-[minmax(0,1.35fr)_minmax(300px,0.95fr)]">
        <div className="space-y-5">
          {reviewError ? (
            <div role="alert" className="rounded-2xl border border-rose-400/30 bg-rose-400/10 px-4 py-3 text-sm text-rose-100">
              {reviewError}
            </div>
          ) : null}

          <section className="rounded-3xl border border-border bg-card/95 p-6 shadow-xl shadow-black/10">
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div className="min-w-0">
                <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Work Object summary</p>
                <h2 className="mt-2 text-2xl font-medium text-foreground">{visibleWorkObject.title}</h2>
                <p className="mt-3 max-w-3xl text-sm leading-6 text-muted-foreground">{visibleWorkObject.summary}</p>
              </div>
              <button
                className="inline-flex h-11 items-center justify-center rounded-full border border-primary/20 bg-primary/10 px-5 text-sm font-medium text-primary transition-colors hover:bg-primary/15 disabled:cursor-not-allowed disabled:opacity-60"
                disabled={isMarkingReady || isLoadingDraftPlan || !draftPlan || Boolean(visibleWorkObject.releaseReadyAt)}
                onClick={handleMarkReadyForRelease}
                type="button"
              >
                {visibleWorkObject.releaseReadyAt
                  ? "Release-ready"
                  : isMarkingReady
                    ? "Marking release-ready..."
                    : "Mark release-ready"}
              </button>
            </div>

            <div className="mt-5 flex flex-wrap items-center gap-2">
              <span className={`rounded-full border px-3 py-1 text-xs font-medium ${panelTone(visibleWorkObject.status)}`}>
                {formatWorkObjectStatus(visibleWorkObject.status)}
              </span>
              <span className={`rounded-full border px-3 py-1 text-xs font-medium ${panelTone(visibleWorkObject.priority)}`}>
                {formatWorkObjectPriority(visibleWorkObject.priority)} priority
              </span>
              {visibleWorkObject.releaseReadyAt ? (
                <span className="rounded-full border border-emerald-400/30 bg-emerald-400/10 px-3 py-1 text-xs font-medium text-emerald-200">
                  Release-ready on {formatTimestamp(visibleWorkObject.releaseReadyAt)}
                </span>
              ) : (
                <span className="rounded-full border border-border bg-background/70 px-3 py-1 text-xs font-medium text-muted-foreground">
                  Not release-ready yet
                </span>
              )}
            </div>

            <div className="mt-5 rounded-2xl border border-border/80 bg-background/55 p-4">
              <p className="text-xs font-medium uppercase tracking-[0.16em] text-muted-foreground">Goal</p>
              <p className="mt-2 text-sm leading-6 text-foreground/90">{visibleWorkObject.goal}</p>
              <p className="mt-4 text-sm leading-6 text-muted-foreground">
                Marking this Work Object release-ready adds it to <span className="text-foreground">Releases</span> so the team can review what is queued next.
              </p>
            </div>
          </section>

          <section className="rounded-3xl border border-border bg-card/95 p-6 shadow-xl shadow-black/10">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Prototype-to-Plan draft</p>
                <h2 className="mt-2 text-xl font-medium text-foreground">Plan summary</h2>
              </div>
              <Link
                className="inline-flex items-center rounded-full border border-border bg-background/70 px-4 py-2 text-sm font-medium text-foreground transition-colors hover:bg-background"
                to={linkTo(`/plans/${visibleWorkObject.id}`)}
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
                  description="Loading the persisted draft so this review reflects the latest planning context."
                  eyebrow="Draft plan summary"
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
                  action={
                    <Link
                      className="inline-flex items-center rounded-full border border-border bg-background/70 px-4 py-2 text-sm font-medium text-foreground transition-colors hover:bg-background"
                      to={linkTo(`/plans/${visibleWorkObject.id}`)}
                    >
                      Open draft plan
                    </Link>
                  }
                  description="This review checkpoint expects a persisted draft plan so the execution context is visible before release."
                  eyebrow="Draft plan summary"
                  title="Persist a draft plan first"
                />
              </div>
            )}
          </section>

          <section className="rounded-3xl border border-border bg-card/95 p-6 shadow-xl shadow-black/10">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Context origin</p>
                <h2 className="mt-2 text-xl font-medium text-foreground">Linked Cluster</h2>
              </div>
              <Link
                className="inline-flex items-center rounded-full border border-border bg-background/70 px-4 py-2 text-sm font-medium text-foreground transition-colors hover:bg-background"
                to={linkTo(`/work/${visibleWorkObject.id}`)}
              >
                Open Work Object
              </Link>
            </div>

            {linkedCluster ? (
              <div className="mt-4 rounded-2xl border border-border/80 bg-background/55 p-4">
                <div className="flex flex-wrap items-center gap-2">
                  <h3 className="text-sm font-medium text-foreground">{linkedCluster.title}</h3>
                  <span className={`rounded-full border px-2.5 py-1 text-xs font-medium ${panelTone(linkedCluster.severity)}`}>
                    {linkedCluster.severity} severity
                  </span>
                  <span className={`rounded-full border px-2.5 py-1 text-xs font-medium ${panelTone(linkedCluster.status)}`}>
                    {formatClusterStatus(linkedCluster.status)}
                  </span>
                </div>
                <p className="mt-3 text-sm leading-6 text-muted-foreground">{linkedCluster.summary}</p>
                <p className="mt-3 text-sm text-foreground/90">{countLabel(linkedCluster.signals.length, "Signal")} linked to this review checkpoint.</p>
              </div>
            ) : (
              <StatePanel
                description="This Work Object can still be reviewed, but no upstream Cluster context is attached yet."
                eyebrow="Context origin"
                title="No linked Cluster"
              />
            )}
          </section>
        </div>

        <aside className="space-y-5">
          <section className="rounded-3xl border border-border bg-card/95 p-6 shadow-xl shadow-black/10">
            <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Decisions</p>
            <h2 className="mt-2 text-lg font-medium text-foreground">{countLabel(visibleWorkObject.decisions.length, "Decision")}</h2>

            {visibleWorkObject.decisions.length === 0 ? (
              <div className="mt-4">
                <StatePanel
                  description="Capture product judgment on the Work Object before using this review checkpoint."
                  eyebrow="Review evidence"
                  title="No Decisions yet"
                />
              </div>
            ) : (
              <div className="mt-4 space-y-3">
                {visibleWorkObject.decisions.map((decision) => (
                  <article key={decision.id} className="rounded-2xl border border-border/80 bg-background/55 p-4">
                    <div className="flex items-start justify-between gap-3">
                      <h3 className="text-sm font-medium text-foreground">{decision.title}</h3>
                      <p className="text-xs text-muted-foreground">{formatTimestamp(decision.createdAt)}</p>
                    </div>
                    <p className="mt-2 text-sm leading-6 text-muted-foreground">{decision.summary}</p>
                    <p className="mt-3 text-sm leading-6 text-foreground/90">{decision.rationale}</p>
                  </article>
                ))}
              </div>
            )}
          </section>

          <section className="rounded-3xl border border-border bg-card/95 p-6 shadow-xl shadow-black/10">
            <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Artifacts</p>
            <h2 className="mt-2 text-lg font-medium text-foreground">{countLabel(visibleWorkObject.artifacts.length, "Artifact")}</h2>

            {visibleWorkObject.artifacts.length === 0 ? (
              <div className="mt-4">
                <StatePanel
                  description="Add at least one Artifact if the release checkpoint needs a concrete output reference."
                  eyebrow="Review evidence"
                  title="No Artifacts yet"
                />
              </div>
            ) : (
              <div className="mt-4 space-y-3">
                {visibleWorkObject.artifacts.map((artifact) => (
                  <article key={artifact.id} className="rounded-2xl border border-border/80 bg-background/55 p-4">
                    <div className="flex flex-wrap items-center justify-between gap-3">
                      <div className="flex flex-wrap items-center gap-2">
                        <h3 className="text-sm font-medium text-foreground">{artifact.title}</h3>
                        <span className="rounded-full border border-primary/20 bg-primary/10 px-2.5 py-1 text-xs font-medium text-primary">
                          {formatArtifactType(artifact.artifactType)}
                        </span>
                      </div>
                      <p className="text-xs text-muted-foreground">{formatTimestamp(artifact.createdAt)}</p>
                    </div>
                    <p className="mt-3 whitespace-pre-line text-sm leading-6 text-foreground/90">{artifact.body}</p>
                  </article>
                ))}
              </div>
            )}
          </section>
        </aside>
      </div>
    </div>
  );
}
