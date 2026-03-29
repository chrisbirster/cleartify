import { EmptyState } from "../components/EmptyState";
import { PageHeader } from "../components/PageHeader";

export function PrototypeNewPage() {
  return (
    <div>
      <PageHeader title="Attach prototype" description="This route is reserved for the prototype attach flow." />
      <EmptyState
        title="Attach flow coming next"
        description="Phase 0 keeps this URL stable so the Prototype-to-Plan flow can land without routing churn."
      />
    </div>
  );
}
