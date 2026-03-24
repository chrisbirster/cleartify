import { Link } from "react-router";
import { PageContainer } from "#/components/page-layout";
import {
  CleartifyIcon,
  MetricCard,
  SectionHeading,
  StatusBadge,
  cx,
} from "#/components/cleartify-ui";

const findings = [
  {
    title: "3 orphaned workspaces still need owners",
    type: "Ownership gap",
    owner: "Platform Team",
    due: "Today",
    tone: "orange" as const,
    dot: "bg-orange-400",
    detail:
      "Shared vendor workspaces were discovered during the last sync, but no department owner or manager mapping is attached yet.",
  },
  {
    title: "Cursor Enterprise access surface pending approval",
    type: "Approval",
    owner: "Engineering",
    due: "Tomorrow",
    tone: "yellow" as const,
    dot: "bg-yellow-400",
    detail:
      "Engineering requested an approved route for IDE usage, but the use case still needs security review and budget sign-off.",
  },
  {
    title: "OpenAI prod-ml project reached 92% of budget",
    type: "Budget alert",
    owner: "ML Team",
    due: "This week",
    tone: "teal" as const,
    dot: "bg-cyan-400",
    detail:
      "Usage is healthy, but spend is tracking ahead of forecast and should be reviewed before new internal pilots are added.",
  },
  {
    title: "Claude Team workspace appeared without a registered use case",
    type: "New discovery",
    owner: "Product Ops",
    due: "This week",
    tone: "purple" as const,
    dot: "bg-violet-400",
    detail:
      "The workspace is active and mapped to a team, but it is missing a business purpose, data classification, and approval trail.",
  },
] as const;

const reviewLanes = [
  {
    title: "Approvals needing review",
    icon: "clipboard-list" as const,
    items: [
      "Cursor Enterprise for engineering copilots",
      "Support assistant expansion to external knowledge sources",
      "Finance content workflow renewal before quarter close",
    ],
  },
  {
    title: "Expiring exceptions",
    icon: "shield" as const,
    items: [
      "Marketing workspace exception expires in 6 days",
      "Internal RAG app data-class override expires in 11 days",
      "Temporary Bedrock budget uplift expires this month",
    ],
  },
  {
    title: "Recently resolved",
    icon: "check-circle" as const,
    items: [
      "GitHub Copilot seat ownership mapped to Platform",
      "Product Claude workspace linked to cost center PD-340",
      "Support bot review packet attached for audit evidence",
    ],
  },
] as const;

const operatingPrinciples = [
  "Route findings to the manager or platform owner who can actually fix them.",
  "Keep approvals, exceptions, and budget watch items in the same operational queue.",
  "Close the loop with evidence so reviews stop depending on screenshots and Slack archaeology.",
] as const;

export function TeamPage() {
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
          <p className="text-[11px] uppercase tracking-[0.28em] text-[var(--app-accent)]">Findings</p>
          <h1 className="mt-4 text-4xl font-semibold tracking-tight text-[var(--app-text)] md:text-5xl">
            Work the queue that turns discovered AI usage into governed AI usage.
          </h1>
          <p className="mt-4 max-w-2xl text-base leading-7 text-[var(--app-text-soft)] md:text-lg">
            Cleartify brings approvals, ownership gaps, expiring exceptions, and budget alerts into
            one place so platform and security teams know what needs attention next.
          </p>
          <div className="mt-8 flex flex-wrap gap-3">
            <Link
              to="/notes/view"
              className="inline-flex min-h-11 items-center justify-center rounded-2xl bg-[var(--app-accent)] px-5 text-sm font-semibold text-[var(--app-accent-ink)]"
            >
              Review systems
            </Link>
            <Link
              to="/stats"
              className="inline-flex min-h-11 items-center justify-center rounded-2xl border border-[var(--app-line)] bg-[var(--app-card)] px-5 text-sm font-semibold text-[var(--app-text)]"
            >
              Check usage
            </Link>
          </div>
        </div>
      </section>

      <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard
          label="Open Findings"
          value="12"
          detail="Issues across ownership, approvals, and newly discovered access surfaces."
          icon="alert-triangle"
        />
        <MetricCard
          label="Pending Approvals"
          value="5"
          detail="Use cases and tools that still need review before they are fully governed."
          icon="clipboard-list"
        />
        <MetricCard
          label="Expiring Exceptions"
          value="4"
          detail="Exception records that need renewal or cleanup before their review date passes."
          icon="shield"
        />
        <MetricCard
          label="Resolved This Month"
          value="19"
          detail="Closed findings with evidence attached for future review readiness."
          icon="check-circle"
        />
      </section>

      <section className="grid gap-8 rounded-[2rem] border border-[var(--app-line)] bg-[var(--app-card)] px-6 py-8 shadow-[0_18px_55px_rgba(0,0,0,0.18)] lg:grid-cols-[1.06fr_0.94fr] md:px-8 md:py-10">
        <div>
          <SectionHeading
            title="Priority queue"
            description="These are the issues a governance manager should be able to triage immediately after opening the app."
            align="left"
          />
          <div className="mt-8 space-y-4">
            {findings.map((finding) => (
              <article
                key={finding.title}
                className="rounded-[1.5rem] border border-[var(--app-line)] bg-[var(--app-muted-surface)] p-5"
              >
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div className="max-w-2xl">
                    <div className="flex items-center gap-3">
                      <span className={cx("mt-0.5 h-2.5 w-2.5 rounded-full", finding.dot)} />
                      <h3 className="text-lg font-medium text-[var(--app-text)]">{finding.title}</h3>
                    </div>
                    <p className="mt-3 text-sm leading-6 text-[var(--app-text-soft)]">{finding.detail}</p>
                  </div>
                  <StatusBadge tone={finding.tone}>{finding.type}</StatusBadge>
                </div>
                <div className="mt-5 flex flex-wrap gap-3 text-sm text-[var(--app-text-soft)]">
                  <span className="rounded-full border border-[var(--app-line)] bg-[var(--app-card)] px-3 py-1.5">
                    Owner: {finding.owner}
                  </span>
                  <span className="rounded-full border border-[var(--app-line)] bg-[var(--app-card)] px-3 py-1.5">
                    Due: {finding.due}
                  </span>
                </div>
              </article>
            ))}
          </div>
        </div>

        <div className="space-y-6">
          {reviewLanes.map((lane) => (
            <section
              key={lane.title}
              className="rounded-[1.75rem] border border-[var(--app-line)] bg-[var(--app-muted-surface)] p-5"
            >
              <div className="flex items-center gap-2">
                <CleartifyIcon name={lane.icon} className="h-4 w-4 text-[var(--app-accent)]" />
                <span className="text-sm font-medium text-[var(--app-text)]">{lane.title}</span>
              </div>
              <div className="mt-5 space-y-3">
                {lane.items.map((item) => (
                  <div
                    key={item}
                    className="rounded-2xl border border-[var(--app-line)] bg-[var(--app-card)] px-4 py-3 text-sm leading-6 text-[var(--app-text-soft)]"
                  >
                    {item}
                  </div>
                ))}
              </div>
            </section>
          ))}
        </div>
      </section>

      <section className="grid gap-8 rounded-[2rem] border border-[var(--app-line)] bg-[var(--app-card)] px-6 py-8 shadow-[0_18px_55px_rgba(0,0,0,0.18)] lg:grid-cols-[0.95fr_1.05fr] md:px-8 md:py-10">
        <div>
          <SectionHeading
            title="How this queue should operate"
            description="Findings are only useful if they turn into clear actions, clean ownership, and durable review evidence."
            align="left"
          />
          <div className="mt-8 space-y-4">
            {operatingPrinciples.map((principle) => (
              <div
                key={principle}
                className="flex gap-3 rounded-[1.5rem] border border-[var(--app-line)] bg-[var(--app-muted-surface)] p-5"
              >
                <span className="mt-2 h-2 w-2 rounded-full bg-[var(--app-accent)]" />
                <p className="text-sm leading-6 text-[var(--app-text-soft)]">{principle}</p>
              </div>
            ))}
          </div>
        </div>

        <div className="rounded-[1.75rem] border border-[var(--app-line)] bg-[var(--app-muted-surface)] p-5">
          <div className="flex items-center gap-2">
            <CleartifyIcon name="search" className="h-4 w-4 text-[var(--app-accent)]" />
            <span className="text-sm font-medium text-[var(--app-text)]">Remediation flow</span>
          </div>
          <div className="mt-5 space-y-4">
            {[
              {
                step: "1. Confirm the surface",
                detail: "Make sure the discovered workspace, project, or app is a real governed access surface and not duplicate noise.",
              },
              {
                step: "2. Assign the right owner",
                detail: "Map the record to the responsible team, manager, or platform group before starting approval work.",
              },
              {
                step: "3. Capture approval state",
                detail: "Register the use case, data class, exception status, and review date directly on the record.",
              },
              {
                step: "4. Attach evidence and close",
                detail: "Resolve the finding only once the owner, review history, and supporting evidence are all present.",
              },
            ].map((step) => (
              <div
                key={step.step}
                className="rounded-2xl border border-[var(--app-line)] bg-[var(--app-card)] p-4"
              >
                <p className="font-medium text-[var(--app-text)]">{step.step}</p>
                <p className="mt-2 text-sm leading-6 text-[var(--app-text-soft)]">{step.detail}</p>
              </div>
            ))}
          </div>
        </div>
      </section>
    </PageContainer>
  );
}
