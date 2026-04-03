import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { Effect } from "effect";

import { PageHeader } from "../components/PageHeader";
import { StatePanel } from "../components/StatePanel";
import { createDecision, deleteDecision, loadDecision, updateDecision } from "../lib/api";
import type { DecisionRecordInput, WorkObject } from "../lib/types";

type DecisionEditorPageProps = {
  mode: "create" | "edit";
  workObjects: WorkObject[];
  onDataChanged?: () => Promise<unknown>;
};

const emptyDecision = (workObjectID: string): DecisionRecordInput => ({
  workObjectId: workObjectID,
  title: "",
  summary: "",
  rationale: "",
});

export function DecisionEditorPage({ mode, workObjects, onDataChanged }: DecisionEditorPageProps) {
  const navigate = useNavigate();
  const { decisionId } = useParams();
  const defaultWorkObjectID = workObjects[0]?.id ?? "";
  const [form, setForm] = useState<DecisionRecordInput>(emptyDecision(defaultWorkObjectID));
  const [state, setState] = useState<{ loading: boolean; saving: boolean; deleting: boolean; error?: string }>({
    loading: mode === "edit",
    saving: false,
    deleting: false,
  });

  useEffect(() => {
    if (mode !== "edit" || !decisionId) {
      return;
    }
    let active = true;
    Effect.runPromise(loadDecision(decisionId))
      .then((decision) => {
        if (active) {
          setForm({
            workObjectId: decision.workObjectId,
            title: decision.title,
            summary: decision.summary,
            rationale: decision.rationale,
          });
          setState((current) => ({ ...current, loading: false }));
        }
      })
      .catch((error) => {
        if (active) {
          setState({ loading: false, saving: false, deleting: false, error: error instanceof Error ? error.message : "Unable to load Decision." });
        }
      });
    return () => {
      active = false;
    };
  }, [decisionId, mode]);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setState((current) => ({ ...current, saving: true, error: undefined }));
    try {
      const saved =
        mode === "create"
          ? await Effect.runPromise(createDecision(form))
          : await Effect.runPromise(updateDecision(decisionId ?? "", form));
      await onDataChanged?.();
      navigate(`/decisions/${saved.id}`);
    } catch (error) {
      setState((current) => ({
        ...current,
        saving: false,
        error: error instanceof Error ? error.message : "Unable to save Decision.",
      }));
    }
  }

  async function handleDelete() {
    if (mode !== "edit" || !decisionId) {
      return;
    }
    setState((current) => ({ ...current, deleting: true, error: undefined }));
    try {
      await Effect.runPromise(deleteDecision(decisionId));
      await onDataChanged?.();
      navigate("/decisions");
    } catch (error) {
      setState((current) => ({
        ...current,
        deleting: false,
        error: error instanceof Error ? error.message : "Unable to delete Decision.",
      }));
    }
  }

  if (state.loading) {
    return <StatePanel surface="page" tone="loading" eyebrow="Decisions" title="Loading Decision" description="Preparing the selected Decision for editing." role="status" />;
  }

  return (
    <div className="space-y-6">
      <PageHeader eyebrow="Decisions" title={mode === "create" ? "New Decision" : form.title || "Edit Decision"} description="Capture or refine a persisted product judgment tied to a Work Object." />
      <form className="space-y-5 rounded-3xl border border-border bg-card/90 p-6 shadow-xl" onSubmit={handleSubmit}>
        <label className="block space-y-2 text-sm text-foreground">
          <span className="text-muted-foreground">Work Object</span>
          <select className="w-full rounded-2xl border border-border bg-background px-4 py-3" value={form.workObjectId} onChange={(event) => setForm((current) => ({ ...current, workObjectId: event.target.value }))}>
            {workObjects.map((workObject) => (
              <option key={workObject.id} value={workObject.id}>
                {workObject.title}
              </option>
            ))}
          </select>
        </label>
        <label className="block space-y-2 text-sm text-foreground">
          <span className="text-muted-foreground">Title</span>
          <input className="w-full rounded-2xl border border-border bg-background px-4 py-3" value={form.title} onChange={(event) => setForm((current) => ({ ...current, title: event.target.value }))} />
        </label>
        <label className="block space-y-2 text-sm text-foreground">
          <span className="text-muted-foreground">Summary</span>
          <textarea className="min-h-24 w-full rounded-2xl border border-border bg-background px-4 py-3" value={form.summary} onChange={(event) => setForm((current) => ({ ...current, summary: event.target.value }))} />
        </label>
        <label className="block space-y-2 text-sm text-foreground">
          <span className="text-muted-foreground">Rationale</span>
          <textarea className="min-h-36 w-full rounded-2xl border border-border bg-background px-4 py-3" value={form.rationale} onChange={(event) => setForm((current) => ({ ...current, rationale: event.target.value }))} />
        </label>
        {state.error ? <p className="text-sm text-red-300">{state.error}</p> : null}
        <div className="flex flex-wrap items-center gap-3">
          <button className="rounded-full bg-primary px-5 py-2 text-sm font-medium text-primary-foreground" disabled={state.saving || state.deleting} type="submit">
            {state.saving ? "Saving…" : mode === "create" ? "Create Decision" : "Save Decision"}
          </button>
          <Link className="rounded-full border border-border px-5 py-2 text-sm font-medium text-foreground" to="/decisions">
            Back to Decisions
          </Link>
          {mode === "edit" ? (
            <button className="rounded-full border border-red-400/30 px-5 py-2 text-sm font-medium text-red-200" disabled={state.saving || state.deleting} onClick={handleDelete} type="button">
              {state.deleting ? "Deleting…" : "Delete Decision"}
            </button>
          ) : null}
        </div>
      </form>
    </div>
  );
}
