import { Link } from "react-router";
import { PageContainer } from "#/components/page-layout";
import {
  CleartifyIcon,
  MetricCard,
  SectionHeading,
  StatusBadge,
} from "#/components/cleartify-ui";

const systems = [
  {
    name: "GitHub Copilot",
    type: "Vendor tool",
    owner: "Platform Team",
    accessSurface: "Organization",
    approval: "Approved",
    tone: "green" as const,
    detail: "Governed at the org level with seat assignments, model mix, and policy ownership.",
  },
  {
    name: "OpenAI API",
    type: "API project",
    owner: "ML Team",
    accessSurface: "Project: prod-ml",
    approval: "Approved",
    tone: "green" as const,
    detail: "Direct API access with project-scoped budgets, service accounts, and usage rollups.",
  },
  {
    name: "Amazon Bedrock",
    type: "Cloud account",
    owner: "Cloud Team",
    accessSurface: "Account + IAM Role",
    approval: "Approved",
    tone: "green" as const,
    detail: "Provider use routed through AWS accounts and IAM boundaries for production workloads.",
  },
  {
    name: "Cursor",
    type: "Workspace tool",
    owner: "Engineering",
    accessSurface: "Workspace",
    approval: "Pending",
    tone: "yellow" as const,
    detail: "Detected via developer tooling, but still waiting on the owning team’s final approval record.",
  },
  {
    name: "Internal RAG App",
    type: "Internal app",
    owner: "Operations Enablement",
    accessSurface: "Application",
    approval: "Conditional",
    tone: "teal" as const,
    detail: "Approved with a restricted data policy and human-review requirements for customer-facing use.",
  },
  {
    name: "Claude Workspace",
    type: "Vendor workspace",
    owner: "Product Ops",
    accessSurface: "Workspace",
    approval: "Review",
    tone: "orange" as const,
    detail: "Mapped to Product, but data-class review is still incomplete for one shared workspace.",
  },
] as const;

const surfaces = [
  {
    name: "Organization",
    example: "GitHub Copilot",
    description: "Org admin boundary with policy settings, seats, and central reporting.",
    tone: "purple" as const,
  },
  {
    name: "Project",
    example: "OpenAI API",
    description: "Project-scoped keys, budgets, and usage controls.",
    tone: "green" as const,
  },
  {
    name: "Account + IAM Role",
    example: "Amazon Bedrock",
    description: "Cloud account and IAM policy boundary with region-level telemetry.",
    tone: "orange" as const,
  },
  {
    name: "Application",
    example: "Internal RAG App",
    description: "Internal AI app with app-level ownership, data sources, and review coverage.",
    tone: "teal" as const,
  },
] as const;

export function NotesPage() {
  return (
    <PageContainer className="space-y-8">
      <section
        className="rounded-[2rem] border border-[var(--app-line)] bg-[var(--app-card-strong)] px-6 py-8 shadow-[0_24px_80px_rgba(0,0,0,0.28)] md:px-8 md:py-10"
        style={{
          backgroundImage:
            "radial-gradient(circle_at_top,color-mix(in_srgb,var(--app-accent)_10%,transparent),transparent_34%),linear-gradient(180deg,color-mix(in_srgb,var(--app-card-strong)_92%,transparent),var(--app-card-strong))",
        }}
      >
        <div className="max-w-4xl">
          <p className="text-[11px] uppercase tracking-[0.28em] text-[var(--app-accent)]">AI Systems</p>
          <h1 className="mt-4 text-4xl font-semibold tracking-tight text-[var(--app-text)] md:text-5xl">
            Keep a governed registry of every AI system and access surface.
          </h1>
          <p className="mt-4 max-w-2xl text-base leading-7 text-[var(--app-text-soft)] md:text-lg">
            Cleartify separates what model was used from how it was accessed, so the registry reflects
            real governance boundaries instead of a single merged AI blob.
          </p>
        </div>
      </section>

      <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard
          label="Tracked Systems"
          value="38"
          detail="Across vendor tools, cloud accounts, internal apps, and API projects."
          icon="layers"
        />
        <MetricCard
          label="Approved"
          value="29"
          detail="Systems that already have approval records and active owners."
          icon="check-circle"
        />
        <MetricCard
          label="Pending Review"
          value="6"
          detail="Systems that have been discovered but still need ownership or policy review."
          icon="search"
        />
        <MetricCard
          label="Conditional"
          value="3"
          detail="Systems that are approved with exception or review requirements."
          icon="shield"
        />
      </section>

      <section className="grid gap-8 rounded-[2rem] border border-[var(--app-line)] bg-[var(--app-card)] px-6 py-8 shadow-[0_18px_55px_rgba(0,0,0,0.18)] lg:grid-cols-[1.15fr_0.85fr] md:px-8 md:py-10">
        <div>
          <SectionHeading
            title="Systems registry"
            description="Every record keeps the owner, approval state, and access surface explicit, so teams can govern the route instead of just the model brand."
            align="left"
          />
          <div className="mt-8 space-y-4">
            {systems.map((system) => (
              <div
                key={system.name}
                className="rounded-[1.5rem] border border-[var(--app-line)] bg-[var(--app-muted-surface)] p-5"
              >
                <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                  <div className="min-w-0">
                    <div className="flex flex-wrap items-center gap-3">
                      <h3 className="text-lg font-medium text-[var(--app-text)]">{system.name}</h3>
                      <StatusBadge tone={system.tone}>{system.approval}</StatusBadge>
                    </div>
                    <p className="mt-2 text-sm text-[var(--app-text-soft)]">
                      {system.type} · Owned by {system.owner}
                    </p>
                  </div>
                  <div className="rounded-full border border-[var(--app-line)] bg-[var(--app-card)] px-3 py-1 text-[11px] uppercase tracking-[0.16em] text-[var(--app-muted)]">
                    {system.accessSurface}
                  </div>
                </div>
                <p className="mt-4 text-sm leading-6 text-[var(--app-text-soft)]">{system.detail}</p>
              </div>
            ))}
          </div>
        </div>

        <div className="space-y-6">
          <div className="rounded-[1.75rem] border border-[var(--app-line)] bg-[var(--app-muted-surface)] p-5">
            <div className="flex items-center gap-2">
              <CleartifyIcon name="network" className="h-4 w-4 text-[var(--app-accent)]" />
              <span className="text-sm font-medium text-[var(--app-text)]">Access surfaces</span>
            </div>
            <div className="mt-5 space-y-3">
              {surfaces.map((surface) => (
                <div
                  key={`${surface.example}-${surface.name}`}
                  className="rounded-2xl border border-[var(--app-line)] bg-[var(--app-card)] p-4"
                >
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <p className="font-medium text-[var(--app-text)]">{surface.example}</p>
                      <p className="mt-1 text-sm text-[var(--app-text-soft)]">{surface.description}</p>
                    </div>
                    <StatusBadge tone={surface.tone}>{surface.name}</StatusBadge>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="rounded-[1.75rem] border border-[var(--app-line)] bg-[var(--app-muted-surface)] p-5">
            <div className="flex items-center gap-2">
              <CleartifyIcon name="terminal" className="h-4 w-4 text-[var(--app-accent)]" />
              <span className="text-sm font-medium text-[var(--app-text)]">Next step</span>
            </div>
            <p className="mt-4 text-sm leading-7 text-[var(--app-text-soft)]">
              Once a system is in the registry, the next actions are always the same: assign the owner,
              confirm the access surface, attach the use case, and close any open findings.
            </p>
            <Link
              to="/team"
              className="mt-6 inline-flex items-center gap-2 rounded-2xl border border-[var(--app-line-strong)] px-4 py-2.5 text-sm font-medium text-[var(--app-text)] transition hover:border-[var(--app-accent)] hover:bg-[var(--app-card)]"
            >
              Review open findings
              <CleartifyIcon name="arrow-right" className="h-4 w-4" />
            </Link>
          </div>
        </div>
      </section>
    </PageContainer>
  );
}
