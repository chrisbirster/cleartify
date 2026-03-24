import { PageContainer } from "#/components/page-layout";
import {
  CleartifyIcon,
  MetricCard,
  SectionHeading,
  StatusBadge,
  cx,
} from "#/components/cleartify-ui";

const providerSpend = [
  { name: "OpenAI", amount: "$18.9k", pct: 82, tone: "bg-[var(--app-accent)]" },
  { name: "Amazon Bedrock", amount: "$14.2k", pct: 63, tone: "bg-[var(--app-accent-strong)]" },
  { name: "GitHub Copilot", amount: "$9.1k", pct: 41, tone: "bg-cyan-400/85" },
  { name: "Internal Apps", amount: "$5.0k", pct: 24, tone: "bg-sky-400/85" },
] as const;

const surfaceMix = [
  { label: "API Projects", detail: "OpenAI and Anthropic direct usage", amount: "41%" },
  { label: "Cloud Accounts", detail: "Bedrock and managed provider routing", amount: "30%" },
  { label: "Vendor Tools", detail: "Copilot, Cursor, and workspace seats", amount: "19%" },
  { label: "Internal Apps", detail: "RAG bots, gateways, and company assistants", amount: "10%" },
] as const;

const budgetAlerts = [
  { label: "OpenAI / prod-ml", owner: "ML Team", state: "92% of monthly budget", tone: "yellow" as const },
  { label: "Bedrock / us-east-1", owner: "Cloud Team", state: "71% of monthly budget", tone: "teal" as const },
  { label: "Copilot / Engineering", owner: "Platform Team", state: "Stable seat utilization", tone: "green" as const },
] as const;

export function StatsPage() {
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
          <p className="text-[11px] uppercase tracking-[0.28em] text-[var(--app-accent)]">Usage Explorer</p>
          <h1 className="mt-4 text-4xl font-semibold tracking-tight text-[var(--app-text)] md:text-5xl">
            Track spend, activity, and budget pressure across governed AI usage.
          </h1>
          <p className="mt-4 max-w-2xl text-base leading-7 text-[var(--app-text-soft)] md:text-lg">
            Cleartify normalizes usage by provider, team, and access surface so finance, platform,
            and security can work from the same system of record.
          </p>
        </div>
      </section>

      <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard
          label="Monthly Spend"
          value="$47.2k"
          detail="Aggregated across approved provider accounts and internal surfaces."
          icon="dollar-sign"
        />
        <MetricCard
          label="Active Teams"
          value="18"
          detail="Teams with meaningful AI usage in the current reporting window."
          icon="users"
        />
        <MetricCard
          label="Active Surfaces"
          value="27"
          detail="Governed access surfaces with usage in the last 30 days."
          icon="network"
        />
        <MetricCard
          label="Budget Alerts"
          value="3"
          detail="Budgets or limits that need review before the month closes."
          icon="alert-triangle"
        />
      </section>

      <section className="grid gap-8 rounded-[2rem] border border-[var(--app-line)] bg-[var(--app-card)] px-6 py-8 shadow-[0_18px_55px_rgba(0,0,0,0.18)] lg:grid-cols-[1.08fr_0.92fr] md:px-8 md:py-10">
        <div>
          <SectionHeading
            title="Provider spend"
            description="See where governed spend is actually concentrated before you optimize budgets, route usage, or add new controls."
            align="left"
          />
          <div className="mt-8 space-y-4">
            {providerSpend.map((provider) => (
              <div
                key={provider.name}
                className="rounded-[1.5rem] border border-[var(--app-line)] bg-[var(--app-muted-surface)] p-5"
              >
                <div className="flex items-center justify-between gap-4">
                  <div>
                    <p className="text-lg font-medium text-[var(--app-text)]">{provider.name}</p>
                    <p className="mt-1 text-sm text-[var(--app-text-soft)]">Monthly governed spend</p>
                  </div>
                  <span className="text-xl font-semibold text-[var(--app-text)]">{provider.amount}</span>
                </div>
                <div className="mt-4 h-3 rounded-full bg-[var(--app-card)]">
                  <div className={cx("h-3 rounded-full", provider.tone)} style={{ width: `${provider.pct}%` }} />
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="space-y-6">
          <div className="rounded-[1.75rem] border border-[var(--app-line)] bg-[var(--app-muted-surface)] p-5">
            <div className="flex items-center gap-2">
              <CleartifyIcon name="activity" className="h-4 w-4 text-[var(--app-accent)]" />
              <span className="text-sm font-medium text-[var(--app-text)]">Access surface mix</span>
            </div>
            <div className="mt-5 space-y-3">
              {surfaceMix.map((surface) => (
                <div
                  key={surface.label}
                  className="rounded-2xl border border-[var(--app-line)] bg-[var(--app-card)] p-4"
                >
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <p className="font-medium text-[var(--app-text)]">{surface.label}</p>
                      <p className="mt-1 text-sm text-[var(--app-text-soft)]">{surface.detail}</p>
                    </div>
                    <span className="text-sm font-semibold text-[var(--app-accent)]">{surface.amount}</span>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="rounded-[1.75rem] border border-[var(--app-line)] bg-[var(--app-muted-surface)] p-5">
            <div className="flex items-center gap-2">
              <CleartifyIcon name="search" className="h-4 w-4 text-[var(--app-accent)]" />
              <span className="text-sm font-medium text-[var(--app-text)]">Budget watchlist</span>
            </div>
            <div className="mt-5 space-y-3">
              {budgetAlerts.map((alert) => (
                <div
                  key={alert.label}
                  className="rounded-2xl border border-[var(--app-line)] bg-[var(--app-card)] p-4"
                >
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <p className="font-medium text-[var(--app-text)]">{alert.label}</p>
                      <p className="mt-1 text-sm text-[var(--app-text-soft)]">{alert.owner}</p>
                    </div>
                    <StatusBadge tone={alert.tone}>{alert.state}</StatusBadge>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>
    </PageContainer>
  );
}
