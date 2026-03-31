import { useEffect, useState } from "react";
import { Effect } from "effect";
import { Link, useParams } from "react-router-dom";

import { EmptyState } from "../components/EmptyState";
import { PageHeader } from "../components/PageHeader";
import { StatePanel } from "../components/StatePanel";
import { generateDraftPlanForWorkObject, loadDraftPlanForWorkObject, updateDraftPlanForWorkObject } from "../lib/api";
import type { Cluster, DraftPlan, DraftPlanInput, WorkObject } from "../lib/types";

type PlanPageProps = {
  workObjects: WorkObject[];
  clusters: Cluster[];
};

function formatTimestamp(value: string) {
  return new Date(value).toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function PlanSection({ title, items, ordered = false }: { title: string; items: string[]; ordered?: boolean }) {
  const ListTag = ordered ? "ol" : "ul";

  return (
    <section className="rounded-3xl border border-border bg-card/90 p-5 shadow-xl shadow-black/10">
      <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">{title}</p>
      <ListTag className={ordered ? "mt-4 space-y-3 list-decimal pl-5" : "mt-4 space-y-3 list-disc pl-5"}>
        {items.map((item) => (
          <li key={item} className="text-sm leading-6 text-foreground/90">
            {item}
          </li>
        ))}
      </ListTag>
    </section>
  );
}

type DraftPlanEditor = {
  overview: string;
  implementationSteps: string;
  edgeCases: string;
  openQuestions: string;
};

function joinDraftPlanItems(items: string[]) {
  return items.join("\n");
}

function splitDraftPlanItems(value: string) {
  return value
    .split("\n")
    .map((item) => item.trim())
    .filter(Boolean);
}

function buildDraftPlanEditor(draftPlan: DraftPlan): DraftPlanEditor {
  return {
    overview: draftPlan.overview,
    implementationSteps: joinDraftPlanItems(draftPlan.implementationSteps),
    edgeCases: joinDraftPlanItems(draftPlan.edgeCases),
    openQuestions: joinDraftPlanItems(draftPlan.openQuestions),
  };
}

function buildDraftPlanInput(editor: DraftPlanEditor): DraftPlanInput {
  return {
    overview: editor.overview,
    implementationSteps: splitDraftPlanItems(editor.implementationSteps),
    edgeCases: splitDraftPlanItems(editor.edgeCases),
    openQuestions: splitDraftPlanItems(editor.openQuestions),
  };
}

export function PlanPage({ workObjects, clusters }: PlanPageProps) {
  const { workObjectId } = useParams();
  const workObject = workObjects.find((entry) => entry.id === workObjectId);
  const linkedCluster = workObject?.clusterId ? clusters.find((entry) => entry.id === workObject.clusterId) : undefined;
  const [draftPlan, setDraftPlan] = useState<DraftPlan | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isGenerating, setIsGenerating] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [editor, setEditor] = useState<DraftPlanEditor>({
    overview: "",
    implementationSteps: "",
    edgeCases: "",
    openQuestions: "",
  });
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!workObject) {
      setDraftPlan(null);
      setIsLoading(false);
      setError(null);
      return;
    }

    let isActive = true;
    setIsLoading(true);
    setError(null);

    Effect.runPromise(loadDraftPlanForWorkObject(workObject.id))
      .then((loadedDraftPlan) => {
        if (!isActive) {
          return;
        }
        setDraftPlan(loadedDraftPlan);
        setEditor(
          loadedDraftPlan
            ? buildDraftPlanEditor(loadedDraftPlan)
            : {
                overview: "",
                implementationSteps: "",
                edgeCases: "",
                openQuestions: "",
              },
        );
        setIsEditing(false);
        setIsLoading(false);
      })
      .catch(() => {
        if (!isActive) {
          return;
        }
        setError("Unable to load draft plan.");
        setIsLoading(false);
      });

    return () => {
      isActive = false;
    };
  }, [workObject]);

  if (!workObject) {
    return (
      <EmptyState
        eyebrow="Prototype-to-Plan"
        title="Work Object not found"
        description="Select a Work Object from Work Objects to review its persisted draft plan."
      />
    );
  }

  const prototypeContext = workObject.prototypeUrl || workObject.prototypeNote;

  const handleGenerate = () => {
    setIsGenerating(true);
    setError(null);

    Effect.runPromise(generateDraftPlanForWorkObject(workObject.id))
      .then((generatedDraftPlan) => {
        setDraftPlan(generatedDraftPlan);
        setEditor(buildDraftPlanEditor(generatedDraftPlan));
        setIsEditing(false);
        setIsGenerating(false);
      })
      .catch(() => {
        setError("Unable to generate draft plan.");
        setIsGenerating(false);
      });
  };

  const beginEditing = () => {
    if (!draftPlan) {
      return;
    }
    setEditor(buildDraftPlanEditor(draftPlan));
    setError(null);
    setIsEditing(true);
  };

  const cancelEditing = () => {
    if (draftPlan) {
      setEditor(buildDraftPlanEditor(draftPlan));
    }
    setError(null);
    setIsEditing(false);
    setIsSaving(false);
  };

  const handleEditorChange = (field: keyof DraftPlanEditor, value: string) => {
    setEditor((current) => ({
      ...current,
      [field]: value,
    }));
  };

  const handleSave = () => {
    if (!draftPlan) {
      return;
    }

    setIsSaving(true);
    setError(null);

    Effect.runPromise(updateDraftPlanForWorkObject(workObject.id, buildDraftPlanInput(editor)))
      .then((updatedDraftPlan) => {
        setDraftPlan(updatedDraftPlan);
        setEditor(buildDraftPlanEditor(updatedDraftPlan));
        setIsEditing(false);
        setIsSaving(false);
      })
      .catch(() => {
        setError("Unable to save draft plan edits.");
        setIsSaving(false);
      });
  };

  return (
    <div>
      <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
        <PageHeader
          title="Draft plan"
          description="Generate a deterministic first-pass plan from the current Work Object and Prototype-to-Plan context."
        />
        <div className="flex flex-wrap gap-3">
          {draftPlan && !isEditing ? (
            <button
              className="inline-flex h-11 items-center justify-center rounded-full border border-border bg-background/70 px-5 text-sm font-medium text-foreground transition-colors hover:bg-background"
              onClick={beginEditing}
              type="button"
            >
              Edit draft
            </button>
          ) : null}
          {isEditing ? (
            <>
              <button
                className="inline-flex h-11 items-center justify-center rounded-full border border-primary/20 bg-primary/10 px-5 text-sm font-medium text-primary transition-colors hover:bg-primary/15 disabled:cursor-not-allowed disabled:opacity-60"
                disabled={isSaving}
                onClick={handleSave}
                type="button"
              >
                {isSaving ? "Saving edits..." : "Save edits"}
              </button>
              <button
                className="inline-flex h-11 items-center justify-center rounded-full border border-border bg-background/70 px-5 text-sm font-medium text-foreground transition-colors hover:bg-background disabled:cursor-not-allowed disabled:opacity-60"
                disabled={isSaving}
                onClick={cancelEditing}
                type="button"
              >
                Cancel
              </button>
            </>
          ) : (
            <button
              className="inline-flex h-11 items-center justify-center rounded-full border border-primary/20 bg-primary/10 px-5 text-sm font-medium text-primary transition-colors hover:bg-primary/15 disabled:cursor-not-allowed disabled:opacity-60"
              disabled={isGenerating || isLoading}
              onClick={handleGenerate}
              type="button"
            >
              {isGenerating ? "Generating draft plan..." : draftPlan ? "Refresh Draft Plan" : "Generate Draft Plan"}
            </button>
          )}
        </div>
      </div>

      <div className="grid gap-5 xl:grid-cols-[minmax(0,1.8fr)_minmax(280px,0.95fr)]">
        <div className="space-y-5">
          {error ? (
            <div role="alert" className="rounded-2xl border border-rose-400/30 bg-rose-400/10 px-4 py-3 text-sm text-rose-100">
              {error}
            </div>
          ) : null}

          {isLoading ? (
            <StatePanel
              description="Bringing the current Prototype-to-Plan context into view for this Work Object."
              eyebrow="Prototype-to-Plan"
              role="status"
              surface="page"
              title="Loading persisted draft plan..."
              tone="loading"
            />
          ) : draftPlan ? (
            <div className="space-y-5">
              {isEditing ? (
                <div className="space-y-5">
                  <section className="rounded-3xl border border-border bg-card/90 p-6 shadow-xl shadow-black/10">
                    <div className="flex flex-wrap items-start justify-between gap-4">
                      <div className="min-w-0">
                        <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Overview</p>
                        <h2 className="mt-2 text-xl font-medium text-foreground">{workObject.title}</h2>
                      </div>
                      <p className="rounded-full border border-border bg-background/70 px-3 py-1 text-xs font-medium text-muted-foreground">
                        Updated {formatTimestamp(draftPlan.updatedAt)}
                      </p>
                    </div>
                    <label className="mt-5 block">
                      <span className="text-xs font-medium uppercase tracking-[0.16em] text-muted-foreground">Overview</span>
                    <textarea
                      aria-label="Overview"
                      className="mt-3 min-h-36 w-full rounded-2xl border border-border bg-background/70 px-4 py-3 text-sm leading-7 text-foreground outline-none transition-colors focus:border-primary/40"
                      onChange={(event) => handleEditorChange("overview", event.target.value)}
                      value={editor.overview}
                    />
                    </label>
                  </section>

                  <section className="rounded-3xl border border-border bg-card/90 p-5 shadow-xl shadow-black/10">
                    <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Implementation steps</p>
                    <p className="mt-2 text-sm text-muted-foreground">One step per line.</p>
                    <textarea
                      aria-label="Implementation steps"
                      className="mt-4 min-h-40 w-full rounded-2xl border border-border bg-background/70 px-4 py-3 text-sm leading-7 text-foreground outline-none transition-colors focus:border-primary/40"
                      onChange={(event) => handleEditorChange("implementationSteps", event.target.value)}
                      value={editor.implementationSteps}
                    />
                  </section>

                  <section className="rounded-3xl border border-border bg-card/90 p-5 shadow-xl shadow-black/10">
                    <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Edge cases</p>
                    <p className="mt-2 text-sm text-muted-foreground">One edge case per line.</p>
                    <textarea
                      aria-label="Edge cases"
                      className="mt-4 min-h-32 w-full rounded-2xl border border-border bg-background/70 px-4 py-3 text-sm leading-7 text-foreground outline-none transition-colors focus:border-primary/40"
                      onChange={(event) => handleEditorChange("edgeCases", event.target.value)}
                      value={editor.edgeCases}
                    />
                  </section>

                  <section className="rounded-3xl border border-border bg-card/90 p-5 shadow-xl shadow-black/10">
                    <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Open questions</p>
                    <p className="mt-2 text-sm text-muted-foreground">One open question per line.</p>
                    <textarea
                      aria-label="Open questions"
                      className="mt-4 min-h-32 w-full rounded-2xl border border-border bg-background/70 px-4 py-3 text-sm leading-7 text-foreground outline-none transition-colors focus:border-primary/40"
                      onChange={(event) => handleEditorChange("openQuestions", event.target.value)}
                      value={editor.openQuestions}
                    />
                  </section>
                </div>
              ) : (
                <div className="space-y-5">
                  <section className="rounded-3xl border border-border bg-card/90 p-6 shadow-xl shadow-black/10">
                    <div className="flex flex-wrap items-start justify-between gap-4">
                      <div className="min-w-0">
                        <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Overview</p>
                        <h2 className="mt-2 text-xl font-medium text-foreground">{workObject.title}</h2>
                        <p className="mt-4 max-w-3xl text-sm leading-7 text-foreground/90">{draftPlan.overview}</p>
                      </div>
                      <p className="rounded-full border border-border bg-background/70 px-3 py-1 text-xs font-medium text-muted-foreground">
                        Updated {formatTimestamp(draftPlan.updatedAt)}
                      </p>
                    </div>
                  </section>

                  <PlanSection title="Implementation steps" items={draftPlan.implementationSteps} ordered />
                  <PlanSection title="Edge cases" items={draftPlan.edgeCases} />
                  <PlanSection title="Open questions" items={draftPlan.openQuestions} />
                </div>
              )}
            </div>
          ) : (
            <EmptyState
              eyebrow="Prototype-to-Plan"
              title="No draft plan yet"
              description="Generate a deterministic draft plan from the current Work Object, prototype reference, and linked Cluster context."
            />
          )}
        </div>

        <aside className="space-y-5">
          <section className="rounded-3xl border border-border bg-card/90 p-5 shadow-xl shadow-black/10">
            <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Work Object</p>
            <h2 className="mt-2 text-lg font-medium text-foreground">{workObject.title}</h2>
            <p className="mt-3 text-sm leading-6 text-muted-foreground">{workObject.summary}</p>
            <p className="mt-4 text-sm text-foreground/90">{workObject.goal}</p>
            <div className="mt-4">
              <Link
                className="inline-flex items-center rounded-full border border-border bg-background/70 px-4 py-2 text-sm font-medium text-foreground transition-colors hover:bg-background"
                to={`/work/${workObject.id}`}
              >
                Open Work Object
              </Link>
            </div>
          </section>

          <section className="rounded-3xl border border-border bg-card/90 p-5 shadow-xl shadow-black/10">
            <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Prototype-to-Plan context</p>
            {prototypeContext ? (
              <div className="mt-4 space-y-4">
                <div>
                  <p className="text-xs font-medium uppercase tracking-[0.16em] text-muted-foreground">Prototype link</p>
                  {workObject.prototypeUrl ? (
                    <a
                      className="mt-2 inline-flex items-center text-sm font-medium text-primary"
                      href={workObject.prototypeUrl}
                      rel="noreferrer"
                      target="_blank"
                    >
                      Open prototype
                    </a>
                  ) : (
                    <p className="mt-2 text-sm text-muted-foreground">No prototype link attached yet.</p>
                  )}
                </div>
                <div>
                  <p className="text-xs font-medium uppercase tracking-[0.16em] text-muted-foreground">Prototype note</p>
                  <p className="mt-2 text-sm leading-6 text-muted-foreground">
                    {workObject.prototypeNote || "No prototype note attached yet."}
                  </p>
                </div>
              </div>
            ) : (
              <div className="mt-4">
                <StatePanel
                  description="No prototype context is attached yet. The draft will fall back to Work Object and linked Cluster context until a prototype reference is added."
                  eyebrow="Prototype-to-Plan"
                  title="No prototype context yet"
                />
              </div>
            )}
          </section>

          <section className="rounded-3xl border border-border bg-card/90 p-5 shadow-xl shadow-black/10">
            <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Linked Cluster</p>
            {linkedCluster ? (
              <div className="mt-4 space-y-3">
                <div className="flex items-center gap-3">
                  <p className="text-sm font-medium text-foreground">{linkedCluster.title}</p>
                  <span className="rounded-full border border-primary/20 bg-primary/10 px-2.5 py-1 text-xs font-medium text-primary">
                    {linkedCluster.severity} severity
                  </span>
                </div>
                <p className="text-sm leading-6 text-muted-foreground">{linkedCluster.summary}</p>
                <Link
                  className="inline-flex items-center rounded-full border border-border bg-background/70 px-4 py-2 text-sm font-medium text-foreground transition-colors hover:bg-background"
                  to={`/inbox/${linkedCluster.id}`}
                >
                  Open Cluster context
                </Link>
              </div>
            ) : (
              <div className="mt-4">
                <StatePanel
                  description="This Work Object can still carry a draft plan, but the upstream Context Inbox link is missing right now."
                  eyebrow="Context origin"
                  title="No linked Cluster"
                />
              </div>
            )}
          </section>
        </aside>
      </div>
    </div>
  );
}
