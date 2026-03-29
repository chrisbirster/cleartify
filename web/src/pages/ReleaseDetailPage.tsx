import { EmptyState } from "../components/EmptyState";
import { PageHeader } from "../components/PageHeader";

export function ReleaseDetailPage() {
  return (
    <div>
      <PageHeader title="Release detail" description="Release-level product memory and outcomes will live on this route." />
      <EmptyState
        title="Release detail scaffold"
        description="Phase 0 keeps this route stable so release notes, outcomes, and linked artifacts can arrive without router churn."
      />
    </div>
  );
}
