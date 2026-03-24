import { PageContainer } from "#/components/page-layout";
import {
  CleartifyIcon,
  MetricCard,
  SectionHeading,
  StatusBadge,
} from "#/components/cleartify-ui";

const orgUnits = [
  {
    name: "Engineering",
    owner: "Platform leadership",
    costCenter: "ENG-100",
    systems: 12,
    approved: 11,
    finding: "1 pending approval in developer tooling",
    tone: "green" as const,
  },
  {
    name: "Data Science",
    owner: "ML enablement",
    costCenter: "DS-220",
    systems: 8,
    approved: 8,
    finding: "All mapped to approved surfaces",
    tone: "green" as const,
  },
  {
    name: "Product",
    owner: "Product ops",
    costCenter: "PD-340",
    systems: 5,
    approved: 4,
    finding: "Workspace owner missing on 1 team tool",
    tone: "yellow" as const,
  },
  {
    name: "Marketing",
    owner: "Revenue systems",
    costCenter: "MKT-120",
    systems: 3,
    approved: 3,
    finding: "No active findings",
    tone: "green" as const,
  },
  {
    name: "Finance",
    owner: "Business systems",
    costCenter: "FIN-410",
    systems: 4,
    approved: 3,
    finding: "External content approval workflow still manual",
    tone: "orange" as const,
  },
  {
    name: "Support",
    owner: "Operations enablement",
    costCenter: "OPS-280",
    systems: 6,
    approved: 5,
    finding: "Internal bot review expires in 9 days",
    tone: "teal" as const,
  },
] as const;

const managers = [
  { name: "Engineering / Platform", owner: "Priya S.", mappedUsers: 126, budget: "$18.4k" },
  { name: "Data Science / Applied ML", owner: "Marcus N.", mappedUsers: 41, budget: "$11.2k" },
  { name: "Product / Core", owner: "Talia R.", mappedUsers: 27, budget: "$6.1k" },
] as const;

export function DecksPage() {
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
          <p className="text-[11px] uppercase tracking-[0.28em] text-[var(--app-accent)]">Org Map</p>
          <h1 className="mt-4 text-4xl font-semibold tracking-tight text-[var(--app-text)] md:text-5xl">
            Map AI usage back to the teams that own it.
          </h1>
          <p className="mt-4 max-w-2xl text-base leading-7 text-[var(--app-text-soft)] md:text-lg">
            Cleartify ties systems, workspaces, projects, and access surfaces back to departments,
            managers, and cost centers so governance has a real owner.
          </p>
        </div>
      </section>

      <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard
          label="Mapped Org Units"
          value="24"
          detail="Departments and subteams actively connected to governed AI usage."
          icon="users"
        />
        <MetricCard
          label="Assigned Owners"
          value="87%"
          detail="AI records with a clear business or technical owner."
          icon="shield"
        />
        <MetricCard
          label="Unmapped Seats"
          value="14"
          detail="Accounts or seats still waiting on team attribution."
          icon="alert-triangle"
        />
        <MetricCard
          label="Budget Coverage"
          value="91%"
          detail="Provider spend tied back to cost centers and review owners."
          icon="dollar-sign"
        />
      </section>

      <section className="grid gap-8 rounded-[2rem] border border-[var(--app-line)] bg-[var(--app-card)] px-6 py-8 shadow-[0_18px_55px_rgba(0,0,0,0.18)] lg:grid-cols-[1.1fr_0.9fr] md:px-8 md:py-10">
        <div>
          <SectionHeading
            title="Team coverage by org unit"
            description="See which departments are already governed and which ones still need ownership, approval, or policy cleanup."
            align="left"
          />
          <div className="mt-8 grid gap-4 md:grid-cols-2">
            {orgUnits.map((unit) => (
              <div
                key={unit.name}
                className="rounded-[1.5rem] border border-[var(--app-line)] bg-[var(--app-muted-surface)] p-5"
              >
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <h3 className="text-lg font-medium text-[var(--app-text)]">{unit.name}</h3>
                    <p className="mt-2 text-sm text-[var(--app-text-soft)]">
                      {unit.owner} · {unit.costCenter}
                    </p>
                  </div>
                  <StatusBadge tone={unit.tone}>
                    {unit.approved}/{unit.systems} approved
                  </StatusBadge>
                </div>
                <div className="mt-5 grid grid-cols-2 gap-3">
                  <div className="rounded-2xl border border-[var(--app-line)] bg-[var(--app-card)] p-3">
                    <p className="text-[11px] uppercase tracking-[0.18em] text-[var(--app-muted)]">Systems</p>
                    <p className="mt-2 text-xl font-semibold text-[var(--app-text)]">{unit.systems}</p>
                  </div>
                  <div className="rounded-2xl border border-[var(--app-line)] bg-[var(--app-card)] p-3">
                    <p className="text-[11px] uppercase tracking-[0.18em] text-[var(--app-muted)]">Approved</p>
                    <p className="mt-2 text-xl font-semibold text-[var(--app-text)]">{unit.approved}</p>
                  </div>
                </div>
                <p className="mt-4 text-sm leading-6 text-[var(--app-text-soft)]">{unit.finding}</p>
              </div>
            ))}
          </div>
        </div>

        <div className="space-y-6">
          <div className="rounded-[1.75rem] border border-[var(--app-line)] bg-[var(--app-muted-surface)] p-5">
            <div className="flex items-center gap-2">
              <CleartifyIcon name="map-pin" className="h-4 w-4 text-[var(--app-accent)]" />
              <span className="text-sm font-medium text-[var(--app-text)]">Manager coverage</span>
            </div>
            <div className="mt-5 space-y-3">
              {managers.map((manager) => (
                <div
                  key={manager.name}
                  className="rounded-2xl border border-[var(--app-line)] bg-[var(--app-card)] p-4"
                >
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <p className="font-medium text-[var(--app-text)]">{manager.name}</p>
                      <p className="mt-1 text-sm text-[var(--app-text-soft)]">{manager.owner}</p>
                    </div>
                    <StatusBadge tone="teal">{manager.budget}</StatusBadge>
                  </div>
                  <p className="mt-3 text-sm text-[var(--app-text-soft)]">
                    {manager.mappedUsers} mapped users across approved and pending surfaces.
                  </p>
                </div>
              ))}
            </div>
          </div>

          <div className="rounded-[1.75rem] border border-[var(--app-line)] bg-[var(--app-muted-surface)] p-5">
            <div className="flex items-center gap-2">
              <CleartifyIcon name="users" className="h-4 w-4 text-[var(--app-accent)]" />
              <span className="text-sm font-medium text-[var(--app-text)]">Why this matters</span>
            </div>
            <ul className="mt-5 space-y-3 text-sm leading-6 text-[var(--app-text-soft)]">
              <li className="flex gap-3">
                <span className="mt-2 h-2 w-2 rounded-full bg-[var(--app-accent)]" />
                Budget owners can finally see which teams are actually using each governed surface.
              </li>
              <li className="flex gap-3">
                <span className="mt-2 h-2 w-2 rounded-full bg-[var(--app-accent)]" />
                Security reviewers can route findings to the right manager instead of chasing shared inboxes.
              </li>
              <li className="flex gap-3">
                <span className="mt-2 h-2 w-2 rounded-full bg-[var(--app-accent)]" />
                Platform teams can see where AI adoption is concentrated before adding new controls or integrations.
              </li>
            </ul>
          </div>
        </div>
      </section>
    </PageContainer>
  );
}
