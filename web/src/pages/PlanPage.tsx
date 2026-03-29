import { EmptyState } from "../components/EmptyState";
import { PageHeader } from "../components/PageHeader";

export function PlanPage() {
  return (
    <div>
      <PageHeader title="Generated plan" description="Phase 0 reserves this route for Prototype-to-Plan output linked to a Work Object." />
      <EmptyState
        title="Plan generation scaffold"
        description="The shell keeps this URL stable so implementation plans can land without routing churn in the next phase."
      />
    </div>
  );
}
