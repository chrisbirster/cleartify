import { useParams } from "react-router-dom";

import { EmptyState } from "../components/EmptyState";
import { PageHeader } from "../components/PageHeader";
import type { WorkObject } from "../lib/types";

type WorkObjectDetailPageProps = {
  workObjects: WorkObject[];
};

export function WorkObjectDetailPage({ workObjects }: WorkObjectDetailPageProps) {
  const { workObjectId } = useParams();
  const workObject = workObjects.find((entry) => entry.id === workObjectId);

  if (!workObject) {
    return <EmptyState title="Work Object not found" description="Select a Work Object from the list to inspect it." />;
  }

  return (
    <div>
      <PageHeader title={workObject.title} description={workObject.goal} />
      <section className="grid gap-6 lg:grid-cols-2">
        <div className="rounded-3xl border border-border bg-card p-5">
          <h2 className="text-lg font-medium text-foreground">Decisions</h2>
          <div className="mt-4 space-y-3">
            {workObject.decisions.map((decision) => (
              <article key={decision.id}>
                <h3 className="text-sm font-medium text-foreground">{decision.title}</h3>
                <p className="mt-1 text-sm text-muted-foreground">{decision.rationale}</p>
              </article>
            ))}
          </div>
        </div>
        <div className="rounded-3xl border border-border bg-card p-5">
          <h2 className="text-lg font-medium text-foreground">Artifacts</h2>
          <div className="mt-4 space-y-3">
            {workObject.artifacts.map((artifact) => (
              <article key={artifact.id}>
                <h3 className="text-sm font-medium text-foreground">{artifact.title}</h3>
                <p className="mt-1 text-sm text-muted-foreground">{artifact.body}</p>
              </article>
            ))}
          </div>
        </div>
      </section>
    </div>
  );
}
