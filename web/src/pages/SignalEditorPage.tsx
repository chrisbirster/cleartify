import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { Effect } from "effect";

import { PageHeader } from "../components/PageHeader";
import { StatePanel } from "../components/StatePanel";
import { createSignal, deleteSignal, loadSignal, updateSignal } from "../lib/api";
import { SIGNAL_SOURCE_TYPES, formatSignalSourceType, type SignalInput } from "../lib/types";

type SignalEditorPageProps = {
  mode: "create" | "edit";
  onDataChanged?: () => Promise<unknown>;
};

const emptySignal: SignalInput = {
  sourceType: "feedback",
  sourceId: "",
  title: "",
  body: "",
  url: "",
};

export function SignalEditorPage({ mode, onDataChanged }: SignalEditorPageProps) {
  const navigate = useNavigate();
  const { signalId } = useParams();
  const [form, setForm] = useState<SignalInput>(emptySignal);
  const [state, setState] = useState<{ loading: boolean; saving: boolean; deleting: boolean; error?: string }>({
    loading: mode === "edit",
    saving: false,
    deleting: false,
  });

  useEffect(() => {
    if (mode !== "edit" || !signalId) {
      return;
    }
    let active = true;
    Effect.runPromise(loadSignal(signalId))
      .then((signal) => {
        if (active) {
          setForm({
            sourceType: signal.sourceType as SignalInput["sourceType"],
            sourceId: signal.sourceId,
            title: signal.title,
            body: signal.body,
            url: signal.url ?? "",
          });
          setState((current) => ({ ...current, loading: false }));
        }
      })
      .catch((error) => {
        if (active) {
          setState({ loading: false, saving: false, deleting: false, error: error instanceof Error ? error.message : "Unable to load Signal." });
        }
      });
    return () => {
      active = false;
    };
  }, [mode, signalId]);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setState((current) => ({ ...current, saving: true, error: undefined }));
    try {
      const saved =
        mode === "create"
          ? await Effect.runPromise(createSignal(form))
          : await Effect.runPromise(updateSignal(signalId ?? "", form));
      await onDataChanged?.();
      navigate(`/signals/${saved.id}`);
    } catch (error) {
      setState((current) => ({
        ...current,
        saving: false,
        error: error instanceof Error ? error.message : "Unable to save Signal.",
      }));
    }
  }

  async function handleDelete() {
    if (mode !== "edit" || !signalId) {
      return;
    }
    setState((current) => ({ ...current, deleting: true, error: undefined }));
    try {
      await Effect.runPromise(deleteSignal(signalId));
      await onDataChanged?.();
      navigate("/signals");
    } catch (error) {
      setState((current) => ({
        ...current,
        deleting: false,
        error: error instanceof Error ? error.message : "Unable to delete Signal.",
      }));
    }
  }

  if (state.loading) {
    return <StatePanel surface="page" tone="loading" eyebrow="Signals" title="Loading Signal" description="Preparing the selected Signal for editing." role="status" />;
  }

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow="Signals"
        title={mode === "create" ? "New Signal" : form.title || "Edit Signal"}
        description={mode === "create" ? "Capture a new source of product context." : "Review and update this Signal's source and context."}
      />
      <form className="space-y-5 rounded-3xl border border-border bg-card/90 p-6 shadow-xl" onSubmit={handleSubmit}>
        <div className="grid gap-5 lg:grid-cols-2">
          <label className="space-y-2 text-sm text-foreground">
            <span className="text-muted-foreground">Source type</span>
            <select
              className="w-full rounded-2xl border border-border bg-background px-4 py-3"
              value={form.sourceType}
              onChange={(event) => setForm((current) => ({ ...current, sourceType: event.target.value as SignalInput["sourceType"] }))}
            >
              {SIGNAL_SOURCE_TYPES.map((sourceType) => (
                <option key={sourceType} value={sourceType}>
                  {formatSignalSourceType(sourceType)}
                </option>
              ))}
            </select>
          </label>
          <label className="space-y-2 text-sm text-foreground">
            <span className="text-muted-foreground">Source id</span>
            <input className="w-full rounded-2xl border border-border bg-background px-4 py-3" value={form.sourceId} onChange={(event) => setForm((current) => ({ ...current, sourceId: event.target.value }))} />
          </label>
        </div>
        <label className="block space-y-2 text-sm text-foreground">
          <span className="text-muted-foreground">Title</span>
          <input className="w-full rounded-2xl border border-border bg-background px-4 py-3" value={form.title} onChange={(event) => setForm((current) => ({ ...current, title: event.target.value }))} />
        </label>
        <label className="block space-y-2 text-sm text-foreground">
          <span className="text-muted-foreground">Body</span>
          <textarea className="min-h-36 w-full rounded-2xl border border-border bg-background px-4 py-3" value={form.body} onChange={(event) => setForm((current) => ({ ...current, body: event.target.value }))} />
        </label>
        <label className="block space-y-2 text-sm text-foreground">
          <span className="text-muted-foreground">External link</span>
          <input className="w-full rounded-2xl border border-border bg-background px-4 py-3" value={form.url ?? ""} onChange={(event) => setForm((current) => ({ ...current, url: event.target.value }))} placeholder="https://..." />
        </label>
        {state.error ? <p className="text-sm text-red-300">{state.error}</p> : null}
        <div className="flex flex-wrap items-center gap-3">
          <button className="rounded-full bg-primary px-5 py-2 text-sm font-medium text-primary-foreground" disabled={state.saving || state.deleting} type="submit">
            {state.saving ? "Saving…" : mode === "create" ? "Create Signal" : "Save Signal"}
          </button>
          <Link className="rounded-full border border-border px-5 py-2 text-sm font-medium text-foreground" to="/signals">
            Back to Signals
          </Link>
          {mode === "edit" ? (
            <button className="rounded-full border border-red-400/30 px-5 py-2 text-sm font-medium text-red-200" disabled={state.saving || state.deleting} onClick={handleDelete} type="button">
              {state.deleting ? "Deleting…" : "Delete Signal"}
            </button>
          ) : null}
        </div>
      </form>
    </div>
  );
}
