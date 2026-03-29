import { Link } from "react-router-dom";

import { PageHeader } from "../components/PageHeader";
import type { WorkObject } from "../lib/types";

type WorkObjectsPageProps = {
  workObjects: WorkObject[];
};

export function WorkObjectsPage({ workObjects }: WorkObjectsPageProps) {
  return (
    <div>
      <PageHeader title="Work Objects" description="Review seeded execution objects created from clustered context." />
      <div className="space-y-4">
        {workObjects.map((workObject) => (
          <Link
            key={workObject.id}
            to={`/work/${workObject.id}`}
            className="block rounded-3xl border border-border bg-card p-5 transition-colors hover:border-primary/40"
          >
            <div className="flex items-center justify-between gap-4">
              <div>
                <h2 className="text-lg font-medium text-foreground">{workObject.title}</h2>
                <p className="mt-2 text-sm text-muted-foreground">{workObject.summary}</p>
              </div>
              <div className="text-right text-sm text-muted-foreground">
                <div className="capitalize">{workObject.status}</div>
                <div className="capitalize">{workObject.priority}</div>
              </div>
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}
