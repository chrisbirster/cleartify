import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { Effect } from "effect";

import { PageHeader } from "../components/PageHeader";
import { StatePanel } from "../components/StatePanel";
import { createArtifact, deleteArtifact, loadArtifact, updateArtifact } from "../lib/api";
import { ARTIFACT_LINK_TYPES, formatArtifactLinkType, type ArtifactInput, type WorkObject } from "../lib/types";

type ArtifactEditorPageProps = {
  mode: "create" | "edit";
  workObjects: WorkObject[];
  onDataChanged?: () => Promise<unknown>;
};

const emptyArtifact = (workObjectID: string): ArtifactInput => ({
  workObjectId: workObjectID,
  type: "note",
  title: "",
  body: "",
  linkType: "",
  url: "",
});

export function ArtifactEditorPage({ mode, workObjects, onDataChanged }: ArtifactEditorPageProps) {
  const navigate = useNavigate();
  const { artifactId } = useParams();
  const defaultWorkObjectID = workObjects[0]?.id ?? "";
  const [form, setForm] = useState<ArtifactInput>(emptyArtifact(defaultWorkObjectID));
  const [state, setState] = useState<{ loading: boolean; saving: boolean; deleting: boolean; error?: string }>({
    loading: mode === "edit",
    saving: false,
    deleting: false,
  });

  useEffect(() => {
    if (mode !== "edit" || !artifactId) {
      return;
    }
    let active = true;
    Effect.runPromise(loadArtifact(artifactId))
      .then((artifact) => {
        if (active) {
          setForm({
            workObjectId: artifact.workObjectId,
            type: artifact.artifactType,
            title: artifact.title,
            body: artifact.body,
            linkType: (artifact.linkType ?? "") as ArtifactInput["linkType"],
            url: artifact.url ?? "",
          });
          setState((current) => ({ ...current, loading: false }));
        }
      })
      .catch((error) => {
        if (active) {
          setState({ loading: false, saving: false, deleting: false, error: error instanceof Error ? error.message : "Unable to load Artifact." });
        }
      });
    return () => {
      active = false;
    };
  }, [artifactId, mode]);

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setState((current) => ({ ...current, saving: true, error: undefined }));
    try {
      const saved =
        mode === "create"
          ? await Effect.runPromise(createArtifact(form))
          : await Effect.runPromise(updateArtifact(artifactId ?? "", form));
      await onDataChanged?.();
      navigate(`/artifacts/${saved.id}`);
    } catch (error) {
      setState((current) => ({
        ...current,
        saving: false,
        error: error instanceof Error ? error.message : "Unable to save Artifact.",
      }));
    }
  }

  async function handleDelete() {
    if (mode !== "edit" || !artifactId) {
      return;
    }
    setState((current) => ({ ...current, deleting: true, error: undefined }));
    try {
      await Effect.runPromise(deleteArtifact(artifactId));
      await onDataChanged?.();
      navigate("/artifacts");
    } catch (error) {
      setState((current) => ({
        ...current,
        deleting: false,
        error: error instanceof Error ? error.message : "Unable to delete Artifact.",
      }));
    }
  }

  if (state.loading) {
    return <StatePanel surface="page" tone="loading" eyebrow="Artifacts" title="Loading Artifact" description="Preparing the selected Artifact for editing." role="status" />;
  }

  return (
    <div className="space-y-6">
      <PageHeader eyebrow="Artifacts" title={mode === "create" ? "New Artifact" : form.title || "Edit Artifact"} description="Capture or refine persisted output and typed external references tied to a Work Object." />
      <form className="space-y-5 rounded-3xl border border-border bg-card/90 p-6 shadow-xl" onSubmit={handleSubmit}>
        <label className="block space-y-2 text-sm text-foreground">
          <span className="text-muted-foreground">Work Object</span>
          <select className="w-full rounded-2xl border border-border bg-background px-4 py-3" value={form.workObjectId ?? ""} onChange={(event) => setForm((current) => ({ ...current, workObjectId: event.target.value }))}>
            {workObjects.map((workObject) => (
              <option key={workObject.id} value={workObject.id}>
                {workObject.title}
              </option>
            ))}
          </select>
        </label>
        <div className="grid gap-5 lg:grid-cols-2">
          <label className="space-y-2 text-sm text-foreground">
            <span className="text-muted-foreground">Artifact type</span>
            <input className="w-full rounded-2xl border border-border bg-background px-4 py-3" value={form.type} onChange={(event) => setForm((current) => ({ ...current, type: event.target.value }))} />
          </label>
          <label className="space-y-2 text-sm text-foreground">
            <span className="text-muted-foreground">External link type</span>
            <select className="w-full rounded-2xl border border-border bg-background px-4 py-3" value={form.linkType ?? ""} onChange={(event) => setForm((current) => ({ ...current, linkType: event.target.value as ArtifactInput["linkType"] }))}>
              <option value="">No external link</option>
              {ARTIFACT_LINK_TYPES.map((linkType) => (
                <option key={linkType} value={linkType}>
                  {formatArtifactLinkType(linkType)}
                </option>
              ))}
            </select>
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
          <span className="text-muted-foreground">External link URL</span>
          <input className="w-full rounded-2xl border border-border bg-background px-4 py-3" value={form.url ?? ""} onChange={(event) => setForm((current) => ({ ...current, url: event.target.value }))} placeholder="https://..." />
        </label>
        {state.error ? <p className="text-sm text-red-300">{state.error}</p> : null}
        <div className="flex flex-wrap items-center gap-3">
          <button className="rounded-full bg-primary px-5 py-2 text-sm font-medium text-primary-foreground" disabled={state.saving || state.deleting} type="submit">
            {state.saving ? "Saving…" : mode === "create" ? "Create Artifact" : "Save Artifact"}
          </button>
          <Link className="rounded-full border border-border px-5 py-2 text-sm font-medium text-foreground" to="/artifacts">
            Back to Artifacts
          </Link>
          {mode === "edit" ? (
            <button className="rounded-full border border-red-400/30 px-5 py-2 text-sm font-medium text-red-200" disabled={state.saving || state.deleting} onClick={handleDelete} type="button">
              {state.deleting ? "Deleting…" : "Delete Artifact"}
            </button>
          ) : null}
        </div>
      </form>
    </div>
  );
}
