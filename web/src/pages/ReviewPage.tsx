import { EmptyState } from "../components/EmptyState";
import { PageHeader } from "../components/PageHeader";

export function ReviewPage() {
  return (
    <div>
      <PageHeader title="Review flow" description="Human review and approval are intentionally reserved behind this route." />
      <EmptyState
        title="Review scaffold"
        description="Phase 0 locks the review URL in place so release checks and judgment-heavy decisions can layer in next."
      />
    </div>
  );
}
