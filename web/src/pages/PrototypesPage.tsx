import { EmptyState } from "../components/EmptyState";
import { PageHeader } from "../components/PageHeader";

export function PrototypesPage() {
  return (
    <div>
      <PageHeader title="Prototypes" description="Prototype review and import land in the next phase." />
      <EmptyState
        title="Prototype-to-Plan scaffold"
        description="Phase 0 keeps the route stable and reserves this space for prototype imports."
      />
    </div>
  );
}
