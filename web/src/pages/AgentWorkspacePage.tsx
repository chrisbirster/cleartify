import { EmptyState } from "../components/EmptyState";
import { PageHeader } from "../components/PageHeader";

export function AgentWorkspacePage() {
  return (
    <div>
      <PageHeader title="Agent Workspace" description="Agent execution actions are stubbed behind this route in Phase 0." />
      <EmptyState
        title="Agent execution scaffold"
        description="This route stays stable so task generation, docs, and artifact drafting can land in the next phase."
      />
    </div>
  );
}
