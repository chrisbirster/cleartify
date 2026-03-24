import { Link } from "react-router";
import { PageContainer } from "#/components/page-layout";

type IconName =
  | "activity"
  | "alert-triangle"
  | "arrow-right"
  | "check-circle"
  | "dollar-sign"
  | "layers"
  | "network"
  | "play"
  | "search"
  | "shield"
  | "terminal"
  | "users";

type BadgeTone = "green" | "orange" | "purple" | "teal" | "yellow";

const orgUnits = [
  { name: "Engineering", systems: "12 systems", healthy: true },
  { name: "Data Science", systems: "8 systems", healthy: true },
  { name: "Product", systems: "5 systems", healthy: false },
  { name: "Marketing", systems: "3 systems", healthy: true },
] as const;

const systemRows = [
  {
    name: "GitHub Copilot",
    meta: "Organization",
    owner: "Platform Team",
    status: "Approved",
    tone: "green" as const,
  },
  {
    name: "OpenAI API",
    meta: "Project: prod-ml",
    owner: "ML Team",
    status: "Approved",
    tone: "green" as const,
  },
  {
    name: "Amazon Bedrock",
    meta: "Account + IAM",
    owner: "Cloud Team",
    status: "Approved",
    tone: "green" as const,
  },
  {
    name: "Cursor",
    meta: "Workspace",
    owner: "Engineering",
    status: "Pending",
    tone: "yellow" as const,
  },
] as const;

const findings = [
  { label: "3 orphaned workspaces found", kind: "Finding", tone: "red" as const },
  { label: "Cursor Enterprise pending approval", kind: "Approval", tone: "yellow" as const },
  { label: "Budget threshold reached: OpenAI", kind: "Alert", tone: "yellow" as const },
  { label: "New system detected: Claude Team", kind: "Discovery", tone: "blue" as const },
] as const;

const keyQuestions = [
  "Which teams are using AI?",
  "Through which tools and access surfaces?",
  "Who owns each one?",
  "What is approved?",
  "What is it costing us?",
  "What needs cleanup right now?",
];

const accessSurfaces = [
  {
    name: "GitHub Copilot",
    badge: "Organization",
    description: "Org admin, seat assignments, policy settings",
    tone: "purple" as const,
  },
  {
    name: "OpenAI API",
    badge: "Project",
    description: "API keys, usage limits, model access",
    tone: "green" as const,
  },
  {
    name: "Amazon Bedrock",
    badge: "Account + IAM Role",
    description: "AWS account, IAM policies, model access",
    tone: "orange" as const,
  },
  {
    name: "Internal RAG App",
    badge: "Application",
    description: "App owner, data sources, access control",
    tone: "teal" as const,
  },
] as const;

const integrations = [
  { name: "OpenAI API", status: "Available" },
  { name: "GitHub Copilot", status: "Available" },
  { name: "Amazon Bedrock", status: "Available" },
  { name: "Internal AI Apps", status: "Available" },
  { name: "ChatGPT Enterprise", status: "Coming soon" },
  { name: "Anthropic Direct", status: "Coming soon" },
] as const;

function cx(...parts: Array<string | false | null | undefined>) {
  return parts.filter(Boolean).join(" ");
}

function Icon({ name, className }: { name: IconName; className?: string }) {
  return (
    <svg
      className={className}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      {name === "activity" ? (
        <path d="M22 12h-2.5a2 2 0 0 0-1.9 1.45l-2.4 8.1a.25.25 0 0 1-.48 0L9.24 2.2a.25.25 0 0 0-.48 0l-2.35 8.35A2 2 0 0 1 4.5 12H2" />
      ) : null}
      {name === "alert-triangle" ? (
        <>
          <path d="m21.73 18-8-14a2 2 0 0 0-3.46 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3" />
          <path d="M12 9v4" />
          <path d="M12 17h.01" />
        </>
      ) : null}
      {name === "arrow-right" ? (
        <>
          <path d="M5 12h14" />
          <path d="m12 5 7 7-7 7" />
        </>
      ) : null}
      {name === "check-circle" ? (
        <>
          <circle cx="12" cy="12" r="10" />
          <path d="m9 12 2 2 4-4" />
        </>
      ) : null}
      {name === "dollar-sign" ? (
        <>
          <line x1="12" y1="2" x2="12" y2="22" />
          <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
        </>
      ) : null}
      {name === "layers" ? (
        <>
          <path d="M12 2 2 7l10 5 10-5-10-5Z" />
          <path d="m2 12 10 5 10-5" />
          <path d="m2 17 10 5 10-5" />
        </>
      ) : null}
      {name === "network" ? (
        <>
          <rect x="16" y="16" width="6" height="6" rx="1" />
          <rect x="2" y="16" width="6" height="6" rx="1" />
          <rect x="9" y="2" width="6" height="6" rx="1" />
          <path d="M5 16v-3a1 1 0 0 1 1-1h12a1 1 0 0 1 1 1v3" />
          <path d="M12 12V8" />
        </>
      ) : null}
      {name === "play" ? <path d="M6 4.5a2 2 0 0 1 3-1.72l10 5.8a2 2 0 0 1 0 3.44l-10 5.8A2 2 0 0 1 6 16.1Z" /> : null}
      {name === "search" ? (
        <>
          <circle cx="11" cy="11" r="8" />
          <path d="m21 21-4.35-4.35" />
        </>
      ) : null}
      {name === "shield" ? (
        <path d="M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.5 3.8 17 5 19 5a1 1 0 0 1 1 1z" />
      ) : null}
      {name === "terminal" ? (
        <>
          <path d="m4 17 6-6-6-6" />
          <path d="M12 19h8" />
        </>
      ) : null}
      {name === "users" ? (
        <>
          <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
          <circle cx="9" cy="7" r="4" />
          <path d="M22 21v-2a4 4 0 0 0-3-3.87" />
          <path d="M16 3.13a4 4 0 0 1 0 7.75" />
        </>
      ) : null}
    </svg>
  );
}

function StatusBadge({
  children,
  tone,
}: {
  children: string;
  tone: BadgeTone;
}) {
  const toneClass = {
    green:
      "border-[rgba(42,214,143,0.24)] bg-[rgba(42,214,143,0.12)] text-[var(--app-success-text)]",
    orange:
      "border-[rgba(246,196,83,0.24)] bg-[rgba(246,196,83,0.12)] text-[var(--app-warning-text)]",
    purple:
      "border-[rgba(172,124,255,0.24)] bg-[rgba(172,124,255,0.12)] text-[rgb(198,169,255)]",
    teal:
      "border-[rgba(22,211,228,0.24)] bg-[rgba(22,211,228,0.12)] text-[var(--app-accent)]",
    yellow:
      "border-[rgba(246,196,83,0.24)] bg-[rgba(246,196,83,0.12)] text-[var(--app-warning-text)]",
  }[tone];

  return (
    <span
      className={cx(
        "inline-flex items-center justify-center rounded-full border px-2.5 py-1 text-[11px] font-medium tracking-[0.02em]",
        toneClass,
      )}
    >
      {children}
    </span>
  );
}

function SectionTitle({
  title,
  description,
  align = "center",
}: {
  title: string;
  description?: string;
  align?: "center" | "left";
}) {
  return (
    <div className={cx("max-w-3xl", align === "center" && "mx-auto text-center")}>
      <h2 className="text-3xl font-semibold tracking-tight text-[var(--app-text)] sm:text-4xl">
        {title}
      </h2>
      {description ? (
        <p className="mt-4 text-base leading-7 text-[var(--app-text-soft)] sm:text-lg">
          {description}
        </p>
      ) : null}
    </div>
  );
}

function MetricCard({
  label,
  value,
  detail,
  icon,
}: {
  label: string;
  value: string;
  detail: string;
  icon: IconName;
}) {
  return (
    <div className="rounded-[1.75rem] border border-[var(--app-line)] bg-[var(--app-card)] p-5 shadow-[0_18px_45px_rgba(0,0,0,0.16)]">
      <div className="flex items-center gap-2">
        <Icon name={icon} className="h-4 w-4 text-[var(--app-accent)]" />
        <span className="text-xs uppercase tracking-[0.2em] text-[var(--app-muted)]">{label}</span>
      </div>
      <p className="mt-4 text-3xl font-semibold tracking-tight text-[var(--app-text)]">{value}</p>
      <p className="mt-2 text-sm leading-6 text-[var(--app-text-soft)]">{detail}</p>
    </div>
  );
}

export function HomePage() {
  return (
    <PageContainer className="space-y-8">
      <section
        className="relative overflow-hidden rounded-[2rem] border border-[var(--app-line)] bg-[var(--app-card-strong)] px-6 py-8 shadow-[0_24px_80px_rgba(0,0,0,0.28)] md:px-8 md:py-10"
        style={{
          backgroundImage:
            "radial-gradient(circle at top, color-mix(in srgb, var(--app-accent) 12%, transparent), transparent 35%), linear-gradient(180deg, color-mix(in srgb, var(--app-card-strong) 92%, transparent), var(--app-card-strong))",
        }}
      >
        <div className="relative max-w-4xl">
          <div className="inline-flex items-center gap-3 rounded-full border border-[var(--app-line)] bg-[rgba(10,20,27,0.72)] px-4 py-2 text-sm text-[var(--app-text-soft)]">
            <span className="h-2 w-2 rounded-full bg-[var(--app-accent)] shadow-[0_0_18px_rgba(22,211,228,0.9)]" />
            Built for IT and platform teams managing AI across vendor tools, cloud accounts, and internal AI apps
          </div>

          <h1 className="mt-8 text-4xl font-semibold tracking-tight text-[var(--app-text)] md:text-5xl xl:text-6xl">
            Governed AI usage,
            <br />
            <span className="bg-[linear-gradient(90deg,var(--app-accent),var(--app-accent-strong))] bg-clip-text text-transparent">
              all in one place
            </span>
          </h1>

          <p className="mt-5 max-w-2xl text-base leading-7 text-[var(--app-text-soft)] md:text-lg">
            Track which teams use AI, through which access surfaces, under which owners, approvals,
            and budgets.
          </p>

          <div className="mt-8 flex flex-wrap gap-3">
            <a
              href="#findings"
              className="inline-flex items-center gap-2 rounded-2xl bg-[var(--app-accent)] px-4 py-2.5 text-sm font-medium text-[var(--app-accent-ink)] transition hover:brightness-105"
            >
              Review findings
              <Icon name="arrow-right" className="h-4 w-4" />
            </a>
            <a
              href="#surfaces"
              className="inline-flex items-center gap-2 rounded-2xl border border-[var(--app-line-strong)] px-4 py-2.5 text-sm font-medium text-[var(--app-text)] transition hover:border-[var(--app-accent)] hover:bg-[var(--app-card)]"
            >
              <Icon name="play" className="h-4 w-4" />
              Explore access surfaces
            </a>
            <Link
              to="/settings"
              className="inline-flex items-center rounded-2xl border border-[var(--app-line-strong)] px-4 py-2.5 text-sm font-medium text-[var(--app-text)] transition hover:border-[var(--app-accent)] hover:bg-[var(--app-card)]"
            >
              Open settings
            </Link>
          </div>
        </div>
      </section>

      <section className="grid gap-6 xl:grid-cols-[1.05fr_1fr_0.88fr]">
        <div className="rounded-[1.75rem] border border-[var(--app-line)] bg-[var(--app-card)] p-5 shadow-[0_18px_50px_rgba(0,0,0,0.18)]">
          <div className="mb-4 flex items-center gap-2">
            <Icon name="users" className="h-4 w-4 text-[var(--app-accent)]" />
            <span className="text-sm font-medium text-[var(--app-text)]">Org Map</span>
          </div>
          <div className="space-y-3">
            {orgUnits.map((unit) => (
              <div
                key={unit.name}
                className="flex items-center justify-between rounded-2xl border border-[var(--app-line)] bg-[var(--app-muted-surface)] px-4 py-3"
              >
                <div className="flex items-center gap-2">
                  <span className="h-2 w-2 rounded-full bg-[var(--app-accent)]" />
                  <span className="text-sm text-[var(--app-text)]">{unit.name}</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-xs text-[var(--app-muted)]">{unit.systems}</span>
                  <Icon
                    name={unit.healthy ? "check-circle" : "alert-triangle"}
                    className={cx(
                      "h-3.5 w-3.5",
                      unit.healthy ? "text-[var(--app-success-text)]" : "text-[var(--app-warning-text)]",
                    )}
                  />
                </div>
              </div>
            ))}
          </div>
        </div>

        <div id="systems" className="rounded-[1.75rem] border border-[var(--app-line)] bg-[var(--app-card)] p-5 shadow-[0_18px_50px_rgba(0,0,0,0.18)]">
          <div className="mb-4 flex items-center gap-2">
            <Icon name="layers" className="h-4 w-4 text-[var(--app-accent)]" />
            <span className="text-sm font-medium text-[var(--app-text)]">AI Systems Registry</span>
          </div>
          <div className="space-y-3">
            {systemRows.map((row) => (
              <div
                key={row.name}
                className="rounded-2xl border border-[var(--app-line)] bg-[var(--app-muted-surface)] p-4"
              >
                <div className="flex items-center justify-between gap-3">
                  <span className="text-sm font-medium text-[var(--app-text)]">{row.name}</span>
                  <StatusBadge tone={row.tone}>{row.status}</StatusBadge>
                </div>
                <div className="mt-2 flex items-center justify-between text-xs text-[var(--app-muted)]">
                  <span>{row.meta}</span>
                  <span>{row.owner}</span>
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="space-y-4">
          <MetricCard
            label="Monthly Spend"
            value="$47.2k"
            detail="Across four approved providers and internal surfaces."
            icon="dollar-sign"
          />
          <MetricCard
            label="Compliance"
            value="94%"
            detail="Governance coverage across mapped AI systems and workspaces."
            icon="shield"
          />
          <div
            id="findings"
            className="rounded-[1.75rem] border border-[var(--app-line)] bg-[var(--app-card)] p-5 shadow-[0_18px_50px_rgba(0,0,0,0.18)]"
          >
            <div className="mb-4 flex items-center gap-2">
              <Icon name="alert-triangle" className="h-4 w-4 text-[var(--app-warning-text)]" />
              <span className="text-sm font-medium text-[var(--app-text)]">Findings</span>
            </div>
            <div className="space-y-3">
              {findings.map((finding) => (
                <div
                  key={finding.label}
                  className="rounded-2xl border border-[var(--app-line)] bg-[var(--app-muted-surface)] px-4 py-3"
                >
                  <div className="flex items-start gap-3">
                    <span
                      className={cx(
                        "mt-1.5 h-2 w-2 rounded-full",
                        finding.tone === "red" && "bg-[var(--app-danger-text)]",
                        finding.tone === "yellow" && "bg-[var(--app-warning-text)]",
                        finding.tone === "blue" && "bg-[var(--app-accent)]",
                      )}
                    />
                    <div>
                      <p className="text-sm font-medium text-[var(--app-text)]">{finding.label}</p>
                      <p className="mt-1 text-xs text-[var(--app-muted)]">{finding.kind}</p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>

      <section className="grid gap-10 rounded-[2rem] border border-[var(--app-line)] bg-[var(--app-card)] px-6 py-8 shadow-[0_18px_55px_rgba(0,0,0,0.18)] lg:grid-cols-[0.95fr_1.05fr] md:px-8 md:py-10">
        <SectionTitle
          title="What Cleartify helps you answer"
          description="These are the questions that keep IT and platform leaders up at night. Cleartify gives you the answers."
          align="left"
        />
        <div className="grid gap-4 sm:grid-cols-2">
          {keyQuestions.map((question) => (
            <div
              key={question}
              className="flex items-start gap-3 rounded-2xl border border-[var(--app-line)] bg-[var(--app-muted-surface)] px-4 py-4"
            >
              <Icon name="check-circle" className="mt-0.5 h-5 w-5 shrink-0 text-[var(--app-accent)]" />
              <span className="text-sm leading-6 text-[var(--app-text)]">{question}</span>
            </div>
          ))}
        </div>
      </section>

      <section
        id="surfaces"
        className="grid gap-8 rounded-[2rem] border border-[var(--app-line)] bg-[var(--app-card)] px-6 py-8 shadow-[0_18px_55px_rgba(0,0,0,0.18)] lg:grid-cols-[1fr_0.92fr] md:px-8 md:py-10"
      >
        <div>
          <span className="inline-flex rounded-full border border-[rgba(22,211,228,0.22)] bg-[rgba(22,211,228,0.12)] px-3 py-1 text-[11px] font-medium uppercase tracking-[0.14em] text-[var(--app-accent)]">
            Core concept
          </span>
          <h2 className="mt-6 text-3xl font-semibold tracking-tight text-[var(--app-text)] sm:text-4xl">
            The real problem is not just which model
          </h2>
          <p className="mt-4 text-xl font-medium text-[var(--app-accent)]">
            It is how AI is being accessed and governed.
          </p>
          <div className="mt-6 space-y-4 text-[var(--app-text-soft)]">
            <p className="leading-7">
              An <span className="font-medium text-[var(--app-text)]">access surface</span> is the specific
              route through which AI is accessed and governed. The same model family can appear
              through multiple access surfaces, each with different owners, logs, budgets, and
              policies.
            </p>
            <p className="leading-7">
              Understanding access surfaces is the foundation of effective AI governance. Without
              this clarity, you are governing blindly.
            </p>
          </div>

          <div className="mt-8 rounded-[1.5rem] border border-[var(--app-line)] bg-[var(--app-muted-surface)] p-5">
            <div className="mb-4 flex items-center gap-3">
              <div className="rounded-xl bg-[var(--app-card)] px-3 py-1.5 text-sm font-medium text-[var(--app-text)]">
                GPT-4
              </div>
              <span className="text-sm text-[var(--app-muted)]">accessed through</span>
            </div>
            <div className="grid gap-2 sm:grid-cols-2">
              {["OpenAI API", "GitHub Copilot", "Internal Gateway", "Team Workspace"].map((item) => (
                <div
                  key={item}
                  className="flex items-center gap-2 rounded-xl border border-[var(--app-line)] bg-[var(--app-card)] px-3 py-2 text-sm text-[var(--app-text-soft)]"
                >
                  <span className="h-1.5 w-1.5 rounded-full bg-[var(--app-accent)]" />
                  {item}
                </div>
              ))}
            </div>
            <p className="mt-4 text-xs text-[var(--app-muted)]">
              Each surface has different owners, budgets, and governance boundaries
            </p>
          </div>
        </div>

        <div>
          <div className="mb-6 flex items-center gap-2">
            <Icon name="network" className="h-4 w-4 text-[var(--app-accent)]" />
            <span className="text-sm font-medium text-[var(--app-text)]">Example Access Surfaces</span>
          </div>
          <div className="space-y-4">
            {accessSurfaces.map((surface) => (
              <div
                key={surface.name}
                className="flex items-start justify-between gap-4 rounded-2xl border border-[var(--app-line)] bg-[var(--app-muted-surface)] p-4"
              >
                <div>
                  <div className="flex flex-wrap items-center gap-3">
                    <h3 className="font-medium text-[var(--app-text)]">{surface.name}</h3>
                    <StatusBadge tone={surface.tone}>{surface.badge}</StatusBadge>
                  </div>
                  <p className="mt-2 text-sm text-[var(--app-text-soft)]">{surface.description}</p>
                </div>
                <span className="mt-2 h-2 w-2 shrink-0 rounded-full bg-[var(--app-accent)]" />
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="rounded-[2rem] border border-[var(--app-line)] bg-[var(--app-card)] px-6 py-8 shadow-[0_18px_55px_rgba(0,0,0,0.18)] md:px-8 md:py-10">
        <SectionTitle
          title="Start with the systems you already use"
          description="Cleartify complements vendor admin consoles. It does not replace them."
        />
        <div className="mt-12 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {integrations.map((integration) => (
            <div
              key={integration.name}
              className="rounded-[1.5rem] border border-[var(--app-line)] bg-[var(--app-muted-surface)] p-5"
            >
              <div className="mb-5 flex items-start justify-between gap-4">
                <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-[var(--app-card)] text-[var(--app-accent)]">
                  <Icon name="layers" className="h-5 w-5" />
                </div>
                <span
                  className={cx(
                    "inline-flex rounded-full border px-2.5 py-1 text-[11px] font-medium tracking-[0.02em]",
                    integration.status === "Available"
                      ? "border-[rgba(42,214,143,0.24)] bg-[rgba(42,214,143,0.12)] text-[var(--app-success-text)]"
                      : "border-[var(--app-line)] bg-[var(--app-card)] text-[var(--app-muted)]",
                  )}
                >
                  {integration.status}
                </span>
              </div>
              <h3 className="text-lg font-medium text-[var(--app-text)]">{integration.name}</h3>
              <p className="mt-2 text-sm leading-7 text-[var(--app-text-soft)]">
                {integration.status === "Available"
                  ? "Connected and ready for governed rollout."
                  : "Planned for a future integration milestone."}
              </p>
            </div>
          ))}
        </div>
      </section>
    </PageContainer>
  );
}
