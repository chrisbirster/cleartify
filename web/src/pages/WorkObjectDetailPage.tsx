import { useEffect, useState, type FormEvent } from "react";
import { Effect } from "effect";
import { Link, useParams } from "react-router-dom";

import { EmptyState } from "../components/EmptyState";
import { PageHeader } from "../components/PageHeader";
import { StatePanel } from "../components/StatePanel";
import { useProjectScope } from "../lib/projectScope";
import {
  createArtifactForWorkObject,
  createDecisionForWorkObject,
  updateWorkObjectCoreFields,
  updateWorkObjectPriority,
  updateWorkObjectPrototype,
  updateWorkObjectStatus,
} from "../lib/api";
import {
  WORK_OBJECT_PRIORITIES,
  WORK_OBJECT_STATUSES,
  formatArtifactType,
  formatClusterStatus,
  formatWorkObjectPriority,
  formatWorkObjectStatus,
  type Artifact,
  type ArtifactInput,
  type Cluster,
  type Decision,
  type DecisionInput,
  type Project,
  type WorkObject,
  type WorkObjectCoreFieldsInput,
  type WorkObjectPriority,
  type WorkObjectPrototypeInput,
  type WorkObjectStatus,
} from "../lib/types";

type WorkObjectDetailPageProps = {
  projects?: Project[];
  workObjects: WorkObject[];
  clusters: Cluster[];
  onWorkObjectUpdated?: (workObject: WorkObject) => void;
};

function normalizeCollection<T>(items: T[] | null | undefined) {
  return Array.isArray(items) ? items : [];
}

function panelTone(value: string) {
  switch (value) {
    case "planning":
      return "border-zinc-400/30 bg-zinc-400/16 text-zinc-100";
    case "urgent":
      return "border-rose-400/30 bg-rose-400/10 text-rose-100";
    case "high":
    case "active":
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

function compactTimestamp(value: string) {
  return new Date(value).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function countLabel(count: number, singular: string, plural = `${singular}s`) {
  return `${count} ${count === 1 ? singular : plural}`;
}

function buildCoreDraft(workObject: WorkObject): WorkObjectCoreFieldsInput {
  return {
    title: workObject.title,
    summary: workObject.summary,
    goal: workObject.goal,
  };
}

function buildDecisionDraft(): DecisionInput {
  return {
    title: "",
    summary: "",
    rationale: "",
  };
}

function buildPrototypeDraft(workObject: WorkObject): WorkObjectPrototypeInput {
  return {
    prototypeUrl: workObject.prototypeUrl ?? "",
    prototypeNote: workObject.prototypeNote ?? "",
  };
}

function buildArtifactDraft(): ArtifactInput {
  return {
    type: "",
    title: "",
    body: "",
  };
}

function mergeDecisions(decisions: Decision[] | null | undefined, createdDecisions: Decision[]) {
  const normalizedDecisions = normalizeCollection(decisions);
  const seen = new Set(normalizedDecisions.map((decision) => decision.id));
  const merged = [...normalizedDecisions];
  for (const decision of createdDecisions) {
    if (seen.has(decision.id)) {
      continue;
    }
    seen.add(decision.id);
    merged.push(decision);
  }
  return merged;
}

function mergeArtifacts(artifacts: Artifact[] | null | undefined, createdArtifacts: Artifact[]) {
  const normalizedArtifacts = normalizeCollection(artifacts);
  const seen = new Set(normalizedArtifacts.map((artifact) => artifact.id));
  const merged = [...normalizedArtifacts];
  for (const artifact of createdArtifacts) {
    if (seen.has(artifact.id)) {
      continue;
    }
    seen.add(artifact.id);
    merged.push(artifact);
  }
  return merged;
}

export function WorkObjectDetailPage({ projects = [], workObjects, clusters, onWorkObjectUpdated }: WorkObjectDetailPageProps) {
  const { workObjectId } = useParams();
  const { linkTo } = useProjectScope(projects);
  const workObject = workObjects.find((entry) => entry.id === workObjectId);
  const [isEditingCore, setIsEditingCore] = useState(false);
  const [isSavingCore, setIsSavingCore] = useState(false);
  const [isEditingPrototype, setIsEditingPrototype] = useState(false);
  const [isSavingPrototype, setIsSavingPrototype] = useState(false);
  const [isAddingDecision, setIsAddingDecision] = useState(false);
  const [isSavingDecision, setIsSavingDecision] = useState(false);
  const [isAddingArtifact, setIsAddingArtifact] = useState(false);
  const [isSavingArtifact, setIsSavingArtifact] = useState(false);
  const [pendingStatus, setPendingStatus] = useState<WorkObjectStatus | null>(null);
  const [pendingPriority, setPendingPriority] = useState<WorkObjectPriority | null>(null);
  const [coreDraft, setCoreDraft] = useState<WorkObjectCoreFieldsInput>({ title: "", summary: "", goal: "" });
  const [prototypeDraft, setPrototypeDraft] = useState<WorkObjectPrototypeInput>({ prototypeUrl: "", prototypeNote: "" });
  const [decisionDraft, setDecisionDraft] = useState<DecisionInput>(buildDecisionDraft());
  const [artifactDraft, setArtifactDraft] = useState<ArtifactInput>(buildArtifactDraft());
  const [createdDecisions, setCreatedDecisions] = useState<Decision[]>([]);
  const [createdArtifacts, setCreatedArtifacts] = useState<Artifact[]>([]);
  const [coreError, setCoreError] = useState<string | null>(null);
  const [prototypeError, setPrototypeError] = useState<string | null>(null);
  const [decisionError, setDecisionError] = useState<string | null>(null);
  const [artifactError, setArtifactError] = useState<string | null>(null);
  const [statusError, setStatusError] = useState<string | null>(null);
  const [priorityError, setPriorityError] = useState<string | null>(null);

  useEffect(() => {
    setCreatedDecisions([]);
    setCreatedArtifacts([]);
    setPrototypeDraft(workObject ? buildPrototypeDraft(workObject) : { prototypeUrl: "", prototypeNote: "" });
    setPrototypeError(null);
    setIsEditingPrototype(false);
    setIsSavingPrototype(false);
    setDecisionDraft(buildDecisionDraft());
    setArtifactDraft(buildArtifactDraft());
    setDecisionError(null);
    setArtifactError(null);
    setIsAddingDecision(false);
    setIsAddingArtifact(false);
    setIsSavingDecision(false);
    setIsSavingArtifact(false);
  }, [workObjectId]);

  if (!workObject) {
    return (
      <EmptyState
        eyebrow="Work Object detail"
        title="Work Object not found"
        description="Select a Work Object from Work Objects to inspect its execution context."
      />
    );
  }

  const linkedCluster = workObject.clusterId ? clusters.find((entry) => entry.id === workObject.clusterId) : undefined;
  const linkedProject = workObject.projectId ? projects.find((entry) => entry.id === workObject.projectId) : undefined;
  const persistedDecisions = normalizeCollection(workObject.decisions);
  const persistedArtifacts = normalizeCollection(workObject.artifacts);
  const evidenceSignals = linkedCluster?.signals ?? [];
  const visibleStatus = pendingStatus ?? workObject.status;
  const visiblePriority = pendingPriority ?? workObject.priority;
  const visibleDecisions = mergeDecisions(persistedDecisions, createdDecisions);
  const visibleArtifacts = mergeArtifacts(persistedArtifacts, createdArtifacts);

  const beginEditingCore = () => {
    setCoreDraft(buildCoreDraft(workObject));
    setCoreError(null);
    setIsEditingCore(true);
  };

  const cancelEditingCore = () => {
    setCoreDraft(buildCoreDraft(workObject));
    setCoreError(null);
    setIsEditingCore(false);
    setIsSavingCore(false);
  };

  const handleCoreDraftChange = (field: keyof WorkObjectCoreFieldsInput, value: string) => {
    setCoreDraft((current) => ({
      ...current,
      [field]: value,
    }));
  };

  const beginAddingDecision = () => {
    setDecisionDraft(buildDecisionDraft());
    setDecisionError(null);
    setIsAddingDecision(true);
  };

  const cancelAddingDecision = () => {
    setDecisionDraft(buildDecisionDraft());
    setDecisionError(null);
    setIsAddingDecision(false);
    setIsSavingDecision(false);
  };

  const handleDecisionDraftChange = (field: keyof DecisionInput, value: string) => {
    setDecisionDraft((current) => ({
      ...current,
      [field]: value,
    }));
  };

  const beginAddingArtifact = () => {
    setArtifactDraft(buildArtifactDraft());
    setArtifactError(null);
    setIsAddingArtifact(true);
  };

  const cancelAddingArtifact = () => {
    setArtifactDraft(buildArtifactDraft());
    setArtifactError(null);
    setIsAddingArtifact(false);
    setIsSavingArtifact(false);
  };

  const handleArtifactDraftChange = (field: keyof ArtifactInput, value: string) => {
    setArtifactDraft((current) => ({
      ...current,
      [field]: value,
    }));
  };

  const beginEditingPrototype = () => {
    setPrototypeDraft(buildPrototypeDraft(workObject));
    setPrototypeError(null);
    setIsEditingPrototype(true);
  };

  const cancelEditingPrototype = () => {
    setPrototypeDraft(buildPrototypeDraft(workObject));
    setPrototypeError(null);
    setIsEditingPrototype(false);
    setIsSavingPrototype(false);
  };

  const handlePrototypeDraftChange = (field: keyof WorkObjectPrototypeInput, value: string) => {
    setPrototypeDraft((current) => ({
      ...current,
      [field]: value,
    }));
  };

  const handleCoreSave = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setIsSavingCore(true);
    setCoreError(null);

    Effect.runPromise(updateWorkObjectCoreFields(workObject.id, coreDraft))
      .then((updatedWorkObject) => {
        setIsSavingCore(false);
        setCoreDraft(buildCoreDraft(updatedWorkObject));
        setCoreError(null);
        setIsEditingCore(false);
        onWorkObjectUpdated?.(updatedWorkObject);
      })
      .catch(() => {
        setIsSavingCore(false);
        setCoreError("Unable to update Work Object details.");
      });
  };

  const handleStatusChange = (nextStatus: WorkObjectStatus) => {
    if (pendingStatus || nextStatus === workObject.status) {
      return;
    }

    setPendingStatus(nextStatus);
    setStatusError(null);

    Effect.runPromise(updateWorkObjectStatus(workObject.id, nextStatus))
      .then((updatedWorkObject) => {
        setPendingStatus(null);
        setStatusError(null);
        onWorkObjectUpdated?.(updatedWorkObject);
      })
      .catch(() => {
        setPendingStatus(null);
        setStatusError("Unable to update Work Object status.");
      });
  };

  const handlePrototypeSave = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setIsSavingPrototype(true);
    setPrototypeError(null);

    Effect.runPromise(updateWorkObjectPrototype(workObject.id, prototypeDraft))
      .then((updatedWorkObject) => {
        setIsSavingPrototype(false);
        setPrototypeDraft(buildPrototypeDraft(updatedWorkObject));
        setPrototypeError(null);
        setIsEditingPrototype(false);
        onWorkObjectUpdated?.(updatedWorkObject);
      })
      .catch(() => {
        setIsSavingPrototype(false);
        setPrototypeError("Unable to update Work Object prototype context.");
      });
  };

  const handlePriorityChange = (nextPriority: WorkObjectPriority) => {
    if (pendingPriority || nextPriority === workObject.priority) {
      return;
    }

    setPendingPriority(nextPriority);
    setPriorityError(null);

    Effect.runPromise(updateWorkObjectPriority(workObject.id, nextPriority))
      .then((updatedWorkObject) => {
        setPendingPriority(null);
        setPriorityError(null);
        onWorkObjectUpdated?.(updatedWorkObject);
      })
      .catch(() => {
        setPendingPriority(null);
        setPriorityError("Unable to update Work Object priority.");
      });
  };

  const handleDecisionSave = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setIsSavingDecision(true);
    setDecisionError(null);

    Effect.runPromise(createDecisionForWorkObject(workObject.id, decisionDraft))
      .then((decision) => {
        const nextVisibleDecisions = mergeDecisions(persistedDecisions, [...createdDecisions, decision]);
        setIsSavingDecision(false);
        setDecisionDraft(buildDecisionDraft());
        setDecisionError(null);
        setIsAddingDecision(false);
        setCreatedDecisions((current) => mergeDecisions(current, [decision]));
        onWorkObjectUpdated?.({
          ...workObject,
          decisions: nextVisibleDecisions,
          artifacts: visibleArtifacts,
        });
      })
      .catch(() => {
        setIsSavingDecision(false);
        setDecisionError("Unable to add Decision.");
      });
  };

  const handleArtifactSave = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setIsSavingArtifact(true);
    setArtifactError(null);

    Effect.runPromise(createArtifactForWorkObject(workObject.id, artifactDraft))
      .then((artifact) => {
        const nextVisibleArtifacts = mergeArtifacts(persistedArtifacts, [...createdArtifacts, artifact]);
        setIsSavingArtifact(false);
        setArtifactDraft(buildArtifactDraft());
        setArtifactError(null);
        setIsAddingArtifact(false);
        setCreatedArtifacts((current) => mergeArtifacts(current, [artifact]));
        onWorkObjectUpdated?.({
          ...workObject,
          decisions: visibleDecisions,
          artifacts: nextVisibleArtifacts,
        });
      })
      .catch(() => {
        setIsSavingArtifact(false);
        setArtifactError("Unable to add Artifact.");
      });
  };

  return (
    <div>
      <PageHeader
        eyebrow={linkedProject ? `Work Object • ${linkedProject.name}` : "Work Object"}
        title={workObject.title}
        description={workObject.summary}
      />

      <section className="grid gap-6 xl:grid-cols-[minmax(0,1.4fr)_minmax(18rem,0.8fr)]">
        <div className="rounded-3xl border border-border bg-card/95 p-6 shadow-xl">
          <div className="flex flex-wrap gap-2 text-xs font-medium">
            {linkedProject ? (
              <span className="rounded-full border border-border bg-background/70 px-3 py-1 text-muted-foreground">
                {linkedProject.name}
              </span>
            ) : null}
            <span className={`rounded-full border px-3 py-1 capitalize ${panelTone(visibleStatus)}`}>{formatWorkObjectStatus(visibleStatus)}</span>
            <span className={`rounded-full border px-3 py-1 capitalize ${panelTone(visiblePriority)}`}>{formatWorkObjectPriority(visiblePriority)} priority</span>
            <span className="rounded-full border border-border bg-secondary/70 px-3 py-1 text-muted-foreground">
              {countLabel(visibleDecisions.length, "Decision")}
            </span>
            <span className="rounded-full border border-border bg-secondary/70 px-3 py-1 text-muted-foreground">
              {countLabel(visibleArtifacts.length, "Artifact")}
            </span>
          </div>

          <div className="mt-5 rounded-2xl border border-border/80 bg-background/55 p-4">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p className="text-sm font-medium text-foreground">Work Object details</p>
                <p className="mt-1 text-sm text-muted-foreground">Keep the title, summary, and goal aligned with the product outcome this Work Object owns.</p>
              </div>
              <div className="flex flex-wrap items-center gap-3">
                <div className="rounded-2xl border border-border/80 bg-background/60 px-4 py-3 text-sm text-muted-foreground">
                  <p>Owner: {workObject.ownerId}</p>
                  <p className="mt-1">Updated: {compactTimestamp(workObject.updatedAt)}</p>
                </div>
                {!isEditingCore ? (
                  <button
                    type="button"
                    className="rounded-full border border-primary/20 bg-primary/10 px-4 py-2 text-sm font-medium text-primary transition-colors hover:bg-primary/15"
                    onClick={beginEditingCore}
                  >
                    Edit Work Object
                  </button>
                ) : null}
              </div>
            </div>

            {isEditingCore ? (
              <form className="mt-4 space-y-4" onSubmit={handleCoreSave}>
                <div>
                  <label className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground" htmlFor="work-object-title">
                    Title
                  </label>
                  <input
                    id="work-object-title"
                    className="mt-2 w-full rounded-2xl border border-border bg-card/80 px-4 py-3 text-sm text-foreground outline-none transition-colors focus:border-primary/40"
                    disabled={isSavingCore}
                    value={coreDraft.title}
                    onChange={(event) => handleCoreDraftChange("title", event.target.value)}
                  />
                </div>
                <div>
                  <label className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground" htmlFor="work-object-summary">
                    Summary
                  </label>
                  <textarea
                    id="work-object-summary"
                    className="mt-2 min-h-24 w-full rounded-2xl border border-border bg-card/80 px-4 py-3 text-sm leading-6 text-foreground outline-none transition-colors focus:border-primary/40"
                    disabled={isSavingCore}
                    value={coreDraft.summary}
                    onChange={(event) => handleCoreDraftChange("summary", event.target.value)}
                  />
                </div>
                <div>
                  <label className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground" htmlFor="work-object-goal">
                    Goal
                  </label>
                  <textarea
                    id="work-object-goal"
                    className="mt-2 min-h-24 w-full rounded-2xl border border-border bg-card/80 px-4 py-3 text-sm leading-6 text-foreground outline-none transition-colors focus:border-primary/40"
                    disabled={isSavingCore}
                    value={coreDraft.goal}
                    onChange={(event) => handleCoreDraftChange("goal", event.target.value)}
                  />
                </div>
                <div className="flex flex-wrap items-center gap-3">
                  <button
                    type="submit"
                    className="rounded-full bg-primary px-5 py-2.5 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-60"
                    disabled={isSavingCore}
                  >
                    {isSavingCore ? "Saving Work Object..." : "Save Work Object"}
                  </button>
                  <button
                    type="button"
                    className="rounded-full border border-border bg-card/80 px-5 py-2.5 text-sm font-medium text-muted-foreground transition-colors hover:border-primary/30 hover:text-foreground disabled:cursor-not-allowed disabled:opacity-60"
                    disabled={isSavingCore}
                    onClick={cancelEditingCore}
                  >
                    Cancel
                  </button>
                </div>
                {coreError ? (
                  <p role="alert" className="text-sm text-amber-200">
                    {coreError}
                  </p>
                ) : null}
              </form>
            ) : (
              <div className="mt-4 grid gap-4">
                <div>
                  <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Title</p>
                  <p className="mt-2 text-base font-medium text-foreground">{workObject.title}</p>
                </div>
                <div>
                  <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Summary</p>
                  <p className="mt-2 text-sm leading-7 text-muted-foreground">{workObject.summary}</p>
                </div>
                <div>
                  <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Goal</p>
                  <p className="mt-2 max-w-3xl text-sm leading-7 text-foreground/95">{workObject.goal}</p>
                </div>
              </div>
            )}
          </div>

          <div className="mt-5 rounded-2xl border border-border/80 bg-background/55 p-4">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p className="text-sm font-medium text-foreground">Prototype context</p>
                <p className="mt-1 text-sm text-muted-foreground">
                  Attach the current prototype link or note so the next Prototype-to-Plan pass stays anchored in product context.
                </p>
              </div>
              {!isEditingPrototype ? (
                <button
                  type="button"
                  className="rounded-full border border-primary/20 bg-primary/10 px-4 py-2 text-sm font-medium text-primary transition-colors hover:bg-primary/15"
                  onClick={beginEditingPrototype}
                >
                  {workObject.prototypeUrl || workObject.prototypeNote ? "Edit prototype context" : "Attach prototype context"}
                </button>
              ) : null}
            </div>

            {isEditingPrototype ? (
              <form className="mt-4 space-y-4" onSubmit={handlePrototypeSave}>
                <div>
                  <label className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground" htmlFor="prototype-link">
                    Prototype link
                  </label>
                  <input
                    id="prototype-link"
                    className="mt-2 w-full rounded-2xl border border-border bg-card/80 px-4 py-3 text-sm text-foreground outline-none transition-colors focus:border-primary/40"
                    disabled={isSavingPrototype}
                    value={prototypeDraft.prototypeUrl}
                    onChange={(event) => handlePrototypeDraftChange("prototypeUrl", event.target.value)}
                  />
                </div>
                <div>
                  <label className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground" htmlFor="prototype-note">
                    Prototype note
                  </label>
                  <textarea
                    id="prototype-note"
                    className="mt-2 min-h-24 w-full rounded-2xl border border-border bg-card/80 px-4 py-3 text-sm leading-6 text-foreground outline-none transition-colors focus:border-primary/40"
                    disabled={isSavingPrototype}
                    value={prototypeDraft.prototypeNote}
                    onChange={(event) => handlePrototypeDraftChange("prototypeNote", event.target.value)}
                  />
                </div>
                <div className="flex flex-wrap items-center gap-3">
                  <button
                    type="submit"
                    className="rounded-full bg-primary px-5 py-2.5 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-60"
                    disabled={isSavingPrototype}
                  >
                    {isSavingPrototype ? "Saving prototype context..." : "Save prototype context"}
                  </button>
                  <button
                    type="button"
                    className="rounded-full border border-border bg-card/80 px-5 py-2.5 text-sm font-medium text-muted-foreground transition-colors hover:border-primary/30 hover:text-foreground disabled:cursor-not-allowed disabled:opacity-60"
                    disabled={isSavingPrototype}
                    onClick={cancelEditingPrototype}
                  >
                    Cancel
                  </button>
                </div>
                {prototypeError ? (
                  <p role="alert" className="text-sm text-amber-200">
                    {prototypeError}
                  </p>
                ) : null}
              </form>
            ) : workObject.prototypeUrl || workObject.prototypeNote ? (
              <div className="mt-4 grid gap-4 lg:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)]">
                <div className="rounded-2xl border border-border/80 bg-card/70 p-4">
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
                <div className="rounded-2xl border border-border/80 bg-card/70 p-4">
                  <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Prototype note</p>
                  <p className="mt-3 text-sm leading-6 text-muted-foreground">
                    {workObject.prototypeNote || "No prototype note attached yet."}
                  </p>
                </div>
              </div>
            ) : (
              <div className="mt-4">
                <StatePanel
                  description="Attach a prototype link or note here before the next Prototype-to-Plan pass so this Work Object stays anchored in visual context."
                  eyebrow="Prototype-to-Plan"
                  title="No prototype attached"
                />
              </div>
            )}
          </div>

          <div className="mt-5 rounded-2xl border border-border/80 bg-background/55 p-4">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <p className="text-sm font-medium text-foreground">Status</p>
                <p className="mt-1 text-sm text-muted-foreground">Show where this Work Object sits in the current execution flow.</p>
              </div>
              {pendingStatus ? (
                <span className="rounded-full border border-primary/20 bg-primary/10 px-3 py-1 text-xs font-medium text-primary">
                  Saving status...
                </span>
              ) : null}
            </div>
            <div className="mt-4 flex flex-wrap gap-2">
              {WORK_OBJECT_STATUSES.map((status) => {
                const isSelected = visibleStatus === status;
                return (
                  <button
                    key={status}
                    type="button"
                    aria-pressed={isSelected}
                    className={`rounded-full border px-3 py-2 text-xs font-medium transition-colors ${isSelected ? panelTone(status) : "border-border bg-card/80 text-muted-foreground hover:border-primary/30 hover:text-foreground"}`}
                    disabled={Boolean(pendingStatus)}
                    onClick={() => handleStatusChange(status)}
                  >
                    {formatWorkObjectStatus(status)}
                  </button>
                );
              })}
            </div>
            {statusError ? (
              <p role="alert" className="mt-3 text-sm text-amber-200">
                {statusError}
              </p>
            ) : null}
          </div>

          <div className="mt-5 rounded-2xl border border-border/80 bg-background/55 p-4">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <p className="text-sm font-medium text-foreground">Priority</p>
                <p className="mt-1 text-sm text-muted-foreground">Keep urgency legible as the Work Object moves from planning toward release.</p>
              </div>
              {pendingPriority ? (
                <span className="rounded-full border border-primary/20 bg-primary/10 px-3 py-1 text-xs font-medium text-primary">
                  Saving priority…
                </span>
              ) : null}
            </div>
            <div className="mt-4 flex flex-wrap gap-2">
              {WORK_OBJECT_PRIORITIES.map((priority) => {
                const isSelected = visiblePriority === priority;
                return (
                  <button
                    key={priority}
                    type="button"
                    aria-pressed={isSelected}
                    className={`rounded-full border px-3 py-2 text-xs font-medium transition-colors capitalize ${isSelected ? panelTone(priority) : "border-border bg-card/80 text-muted-foreground hover:border-primary/30 hover:text-foreground"}`}
                    disabled={Boolean(pendingPriority)}
                    onClick={() => handlePriorityChange(priority)}
                  >
                    {formatWorkObjectPriority(priority)}
                  </button>
                );
              })}
            </div>
            {priorityError ? (
              <p role="alert" className="mt-3 text-sm text-amber-200">
                {priorityError}
              </p>
            ) : null}
          </div>
        </div>

        <aside className="rounded-3xl border border-border bg-card/95 p-6 shadow-xl">
          <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Context origin</p>
          {linkedCluster ? (
            <div className="mt-4 space-y-4">
              <div>
                <h2 className="text-lg font-medium text-foreground">Linked Cluster</h2>
                <p className="mt-2 text-sm font-medium text-foreground">{linkedCluster.title}</p>
                <p className="mt-2 text-sm leading-6 text-muted-foreground">{linkedCluster.summary}</p>
              </div>
              <div className="flex flex-wrap gap-2 text-xs font-medium">
                <span className={`rounded-full border px-3 py-1 capitalize ${panelTone(linkedCluster.severity)}`}>{linkedCluster.severity}</span>
                <span className={`rounded-full border px-3 py-1 capitalize ${panelTone(linkedCluster.status)}`}>{formatClusterStatus(linkedCluster.status)}</span>
                <span className="rounded-full border border-border bg-secondary/70 px-3 py-1 text-muted-foreground">
                  {countLabel(linkedCluster.signals.length, "Signal")}
                </span>
              </div>
              <Link
                className="inline-flex items-center rounded-full border border-primary/20 bg-primary/10 px-4 py-2 text-sm font-medium text-primary transition-colors hover:bg-primary/15"
                to={linkTo(`/inbox/${linkedCluster.id}`)}
              >
                Open Context Inbox
              </Link>
            </div>
          ) : (
            <div className="mt-4">
              <StatePanel
                description="This Work Object stands on its own for now, so the Context Inbox relationship is still open."
                eyebrow="Context origin"
                title="No linked Cluster"
              />
            </div>
          )}
        </aside>
      </section>

      <section className="mt-6 grid gap-6 xl:grid-cols-[minmax(0,1.15fr)_minmax(0,0.85fr)_minmax(0,0.85fr)]">
        <div className="rounded-3xl border border-border bg-card/95 p-6 shadow-xl">
          <div className="flex items-center justify-between gap-4">
            <div>
              <h2 className="text-lg font-medium text-foreground">Signal evidence</h2>
              <p className="mt-1 text-sm text-muted-foreground">Signals attached through the linked Cluster that still explain why this Work Object exists.</p>
            </div>
            <span className="rounded-full border border-border bg-secondary/70 px-3 py-1 text-xs font-medium text-muted-foreground">
              {countLabel(evidenceSignals.length, "Signal")}
            </span>
          </div>

          <div className="mt-4 space-y-3">
            {evidenceSignals.length > 0 ? (
              evidenceSignals.map((signal) => (
                <article key={signal.id} className="rounded-2xl border border-border/80 bg-background/50 p-4">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="rounded-full border border-border bg-secondary/70 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">
                      {signal.sourceType}
                    </span>
                    {signal.url ? (
                      <a
                        className="text-xs font-medium text-primary"
                        href={signal.url}
                        rel="noreferrer"
                        target="_blank"
                      >
                        Open source
                      </a>
                    ) : null}
                  </div>
                  <h3 className="mt-3 text-sm font-medium text-foreground">{signal.title}</h3>
                  <p className="mt-2 text-sm leading-6 text-muted-foreground">{signal.body}</p>
                </article>
              ))
            ) : (
              <StatePanel
                description="Link a Cluster to this Work Object to keep the underlying product context attached."
                eyebrow="Evidence"
                title="No supporting Signals yet"
              />
            )}
          </div>
        </div>

        <div className="rounded-3xl border border-border bg-card/95 p-6 shadow-xl">
          <div className="flex items-center justify-between gap-4">
            <div>
              <h2 className="text-lg font-medium text-foreground">Decisions</h2>
              <p className="mt-1 text-sm text-muted-foreground">Product judgment already recorded on this Work Object.</p>
            </div>
            <div className="flex items-center gap-3">
              <span className="rounded-full border border-border bg-secondary/70 px-3 py-1 text-xs font-medium text-muted-foreground">
                {countLabel(visibleDecisions.length, "Decision")}
              </span>
              {!isAddingDecision ? (
                <button
                  type="button"
                  className="rounded-full border border-primary/20 bg-primary/10 px-4 py-2 text-sm font-medium text-primary transition-colors hover:bg-primary/15"
                  onClick={beginAddingDecision}
                >
                  Add Decision
                </button>
              ) : null}
            </div>
          </div>
          {isAddingDecision ? (
            <form className="mt-4 space-y-4 rounded-2xl border border-border/80 bg-background/45 p-4" onSubmit={handleDecisionSave}>
              <div>
                <label className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground" htmlFor="decision-title">
                  Decision title
                </label>
                <input
                  id="decision-title"
                  className="mt-2 w-full rounded-2xl border border-border bg-card/80 px-4 py-3 text-sm text-foreground outline-none transition-colors focus:border-primary/40"
                  disabled={isSavingDecision}
                  value={decisionDraft.title}
                  onChange={(event) => handleDecisionDraftChange("title", event.target.value)}
                />
              </div>
              <div>
                <label className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground" htmlFor="decision-summary">
                  Summary
                </label>
                <textarea
                  id="decision-summary"
                  className="mt-2 min-h-20 w-full rounded-2xl border border-border bg-card/80 px-4 py-3 text-sm leading-6 text-foreground outline-none transition-colors focus:border-primary/40"
                  disabled={isSavingDecision}
                  value={decisionDraft.summary}
                  onChange={(event) => handleDecisionDraftChange("summary", event.target.value)}
                />
              </div>
              <div>
                <label className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground" htmlFor="decision-rationale">
                  Rationale
                </label>
                <textarea
                  id="decision-rationale"
                  className="mt-2 min-h-24 w-full rounded-2xl border border-border bg-card/80 px-4 py-3 text-sm leading-6 text-foreground outline-none transition-colors focus:border-primary/40"
                  disabled={isSavingDecision}
                  value={decisionDraft.rationale}
                  onChange={(event) => handleDecisionDraftChange("rationale", event.target.value)}
                />
              </div>
              <div className="flex flex-wrap items-center gap-3">
                <button
                  type="submit"
                  className="rounded-full bg-primary px-5 py-2.5 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-60"
                  disabled={isSavingDecision}
                >
                  {isSavingDecision ? "Saving Decision..." : "Save Decision"}
                </button>
                <button
                  type="button"
                  className="rounded-full border border-border bg-card/80 px-5 py-2.5 text-sm font-medium text-muted-foreground transition-colors hover:border-primary/30 hover:text-foreground disabled:cursor-not-allowed disabled:opacity-60"
                  disabled={isSavingDecision}
                  onClick={cancelAddingDecision}
                >
                  Cancel
                </button>
              </div>
              {decisionError ? (
                <p role="alert" className="text-sm text-amber-200">
                  {decisionError}
                </p>
              ) : null}
            </form>
          ) : null}
          <div className="mt-4 space-y-3">
            {visibleDecisions.length > 0 ? (
              visibleDecisions.map((decision) => (
                <article key={decision.id} className="rounded-2xl border border-border/80 bg-background/50 p-4">
                  <span className="rounded-full border border-border bg-secondary/70 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">
                    Decision
                  </span>
                  <h3 className="mt-3 text-sm font-medium text-foreground">{decision.title}</h3>
                  <p className="mt-2 text-sm leading-6 text-muted-foreground">{decision.summary}</p>
                  <p className="mt-3 border-t border-border/70 pt-3 text-sm leading-6 text-foreground/90">{decision.rationale}</p>
                </article>
              ))
            ) : (
              <StatePanel
                description="This Work Object does not have any captured product judgment yet."
                eyebrow="Product judgment"
                title="No Decisions yet"
              />
            )}
          </div>
        </div>

        <div className="rounded-3xl border border-border bg-card/95 p-6 shadow-xl">
          <div className="flex items-center justify-between gap-4">
            <div>
              <h2 className="text-lg font-medium text-foreground">Artifacts</h2>
              <p className="mt-1 text-sm text-muted-foreground">Execution output and working material already attached to this Work Object.</p>
            </div>
            <div className="flex items-center gap-3">
              <span className="rounded-full border border-border bg-secondary/70 px-3 py-1 text-xs font-medium text-muted-foreground">
                {countLabel(visibleArtifacts.length, "Artifact")}
              </span>
              {!isAddingArtifact ? (
                <button
                  type="button"
                  className="rounded-full border border-primary/20 bg-primary/10 px-4 py-2 text-sm font-medium text-primary transition-colors hover:bg-primary/15"
                  onClick={beginAddingArtifact}
                >
                  Add Artifact
                </button>
              ) : null}
            </div>
          </div>
          {isAddingArtifact ? (
            <form className="mt-4 space-y-4 rounded-2xl border border-border/80 bg-background/45 p-4" onSubmit={handleArtifactSave}>
              <div>
                <label className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground" htmlFor="artifact-type">
                  Artifact type
                </label>
                <input
                  id="artifact-type"
                  className="mt-2 w-full rounded-2xl border border-border bg-card/80 px-4 py-3 text-sm text-foreground outline-none transition-colors focus:border-primary/40"
                  disabled={isSavingArtifact}
                  value={artifactDraft.type}
                  onChange={(event) => handleArtifactDraftChange("type", event.target.value)}
                />
              </div>
              <div>
                <label className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground" htmlFor="artifact-title">
                  Artifact title
                </label>
                <input
                  id="artifact-title"
                  className="mt-2 w-full rounded-2xl border border-border bg-card/80 px-4 py-3 text-sm text-foreground outline-none transition-colors focus:border-primary/40"
                  disabled={isSavingArtifact}
                  value={artifactDraft.title}
                  onChange={(event) => handleArtifactDraftChange("title", event.target.value)}
                />
              </div>
              <div>
                <label className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground" htmlFor="artifact-body">
                  Body
                </label>
                <textarea
                  id="artifact-body"
                  className="mt-2 min-h-24 w-full rounded-2xl border border-border bg-card/80 px-4 py-3 text-sm leading-6 text-foreground outline-none transition-colors focus:border-primary/40"
                  disabled={isSavingArtifact}
                  value={artifactDraft.body}
                  onChange={(event) => handleArtifactDraftChange("body", event.target.value)}
                />
              </div>
              <div className="flex flex-wrap items-center gap-3">
                <button
                  type="submit"
                  className="rounded-full bg-primary px-5 py-2.5 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-60"
                  disabled={isSavingArtifact}
                >
                  {isSavingArtifact ? "Saving Artifact..." : "Save Artifact"}
                </button>
                <button
                  type="button"
                  className="rounded-full border border-border bg-card/80 px-5 py-2.5 text-sm font-medium text-muted-foreground transition-colors hover:border-primary/30 hover:text-foreground disabled:cursor-not-allowed disabled:opacity-60"
                  disabled={isSavingArtifact}
                  onClick={cancelAddingArtifact}
                >
                  Cancel
                </button>
              </div>
              {artifactError ? (
                <p role="alert" className="text-sm text-amber-200">
                  {artifactError}
                </p>
              ) : null}
            </form>
          ) : null}
          <div className="mt-4 space-y-3">
            {visibleArtifacts.length > 0 ? (
              visibleArtifacts.map((artifact) => (
                <article key={artifact.id} className="rounded-2xl border border-border/80 bg-background/50 p-4">
                  <span className="rounded-full border border-border bg-secondary/70 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">
                    {formatArtifactType(artifact.artifactType)}
                  </span>
                  <h3 className="mt-3 text-sm font-medium text-foreground">{artifact.title}</h3>
                  <p className="mt-2 text-sm leading-6 text-muted-foreground">{artifact.body}</p>
                </article>
              ))
            ) : (
              <StatePanel
                description="This Work Object does not have any attached outputs yet."
                eyebrow="Execution output"
                title="No Artifacts yet"
              />
            )}
          </div>
        </div>
      </section>
    </div>
  );
}
