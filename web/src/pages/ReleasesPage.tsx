import { EmptyState } from "../components/EmptyState";
import { PageHeader } from "../components/PageHeader";

export function ReleasesPage() {
  return (
    <div>
      <PageHeader title="Releases" description="Release summaries and product memory ship behind this stable Phase 0 route." />
      <EmptyState
        title="Release list scaffold"
        description="The product shell keeps this route ready for release summaries, outcomes, and linked decisions."
      />
    </div>
  );
}
