import { type FormEvent, useState } from "react";
import { Effect } from "effect";
import { Link, useNavigate } from "react-router-dom";

import { PageHeader } from "../components/PageHeader";
import { createProject } from "../lib/api";

type ProjectNewPageProps = {
  onDataChanged?: () => Promise<unknown>;
};

export function ProjectNewPage({ onDataChanged }: ProjectNewPageProps) {
  const navigate = useNavigate();
  const [name, setName] = useState("");
  const [summary, setSummary] = useState("");
  const [isCreating, setIsCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setIsCreating(true);
    setError(null);

    try {
      const project = await Effect.runPromise(
        createProject({
          name,
          summary,
        }),
      );
      await onDataChanged?.();
      navigate(`/projects/${project.id}`);
    } catch {
      setError("Unable to create Project.");
      setIsCreating(false);
    }
  };

  return (
    <div>
      <PageHeader
        eyebrow="Projects"
        title="New Project"
        description="Create a project boundary so Cleartify can scope the dashboard, Context Inbox, Work Objects, and release loop around real product work."
      />
      <div className="max-w-3xl rounded-3xl border border-border bg-card/95 p-6 shadow-xl shadow-black/10">
        <form className="space-y-5" onSubmit={handleSubmit}>
          <div>
            <label className="text-sm font-medium text-foreground" htmlFor="project-name">
              Project name
            </label>
            <input
              className="mt-2 w-full rounded-2xl border border-border bg-background/70 px-4 py-3 text-sm text-foreground outline-none transition focus:border-primary/40"
              id="project-name"
              onChange={(event) => setName(event.currentTarget.value)}
              placeholder="Guided onboarding refresh"
              value={name}
            />
          </div>

          <div>
            <label className="text-sm font-medium text-foreground" htmlFor="project-summary">
              Summary
            </label>
            <textarea
              className="mt-2 min-h-32 w-full rounded-2xl border border-border bg-background/70 px-4 py-3 text-sm text-foreground outline-none transition focus:border-primary/40"
              id="project-summary"
              onChange={(event) => setSummary(event.currentTarget.value)}
              placeholder="Describe the product outcome this project should own so the rest of the workflow can scope around it."
              value={summary}
            />
          </div>

          {error ? <p className="text-sm text-rose-200">{error}</p> : null}

          <div className="flex flex-wrap gap-3">
            <button
              className="rounded-full bg-primary px-5 py-2 text-sm font-medium text-primary-foreground transition hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-60"
              disabled={isCreating}
              type="submit"
            >
              {isCreating ? "Creating Project..." : "Create Project"}
            </button>
            <Link
              className="rounded-full border border-border px-5 py-2 text-sm font-medium text-foreground transition hover:border-border/70 hover:bg-card/70"
              to="/projects"
            >
              Cancel
            </Link>
          </div>
        </form>
      </div>
    </div>
  );
}
