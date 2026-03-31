import { EmptyState } from "../components/EmptyState";
import { PageHeader } from "../components/PageHeader";

export function PrototypeNewPage() {
  return (
    <div>
      <PageHeader title="Attach prototype" description="Prototype capture now starts from the Work Object detail view." />
      <EmptyState
        title="Bulk import coming later"
        description="This route stays stable for a broader prototype import flow once the lightweight Work Object attach loop is in place."
      />
    </div>
  );
}
