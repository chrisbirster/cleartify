import type { ReactNode } from "react";

export type CleartifyIconName =
  | "activity"
  | "alert-triangle"
  | "arrow-right"
  | "check-circle"
  | "clipboard-list"
  | "dollar-sign"
  | "file-warning"
  | "layers"
  | "map-pin"
  | "network"
  | "search"
  | "shield"
  | "terminal"
  | "users";

export type BadgeTone =
  | "blue"
  | "green"
  | "orange"
  | "purple"
  | "slate"
  | "teal"
  | "yellow";

export function cx(...parts: Array<string | false | null | undefined>) {
  return parts.filter(Boolean).join(" ");
}

export function CleartifyIcon({
  name,
  className,
}: {
  name: CleartifyIconName;
  className?: string;
}) {
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
      {name === "clipboard-list" ? (
        <>
          <rect x="8" y="2" width="8" height="4" rx="1" />
          <path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2" />
          <path d="M12 11h4" />
          <path d="M12 16h4" />
          <path d="M8 11h.01" />
          <path d="M8 16h.01" />
        </>
      ) : null}
      {name === "dollar-sign" ? (
        <>
          <line x1="12" y1="2" x2="12" y2="22" />
          <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
        </>
      ) : null}
      {name === "file-warning" ? (
        <>
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
          <path d="M14 2v6h6" />
          <path d="M12 11v4" />
          <path d="M12 18h.01" />
        </>
      ) : null}
      {name === "layers" ? (
        <>
          <path d="M12 2 2 7l10 5 10-5-10-5Z" />
          <path d="m2 12 10 5 10-5" />
          <path d="m2 17 10 5 10-5" />
        </>
      ) : null}
      {name === "map-pin" ? (
        <>
          <path d="M20 10c0 5-5.54 10.2-7.4 11.8a1 1 0 0 1-1.2 0C9.54 20.2 4 15 4 10a8 8 0 0 1 16 0Z" />
          <circle cx="12" cy="10" r="3" />
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

export function StatusBadge({
  children,
  tone,
}: {
  children: ReactNode;
  tone: BadgeTone;
}) {
  const toneClass = {
    blue: "border-[rgba(64,169,255,0.24)] bg-[rgba(64,169,255,0.12)] text-sky-300",
    green:
      "border-[rgba(42,214,143,0.24)] bg-[rgba(42,214,143,0.12)] text-[var(--app-success-text)]",
    orange:
      "border-[rgba(255,166,92,0.24)] bg-[rgba(255,166,92,0.12)] text-[var(--app-warning-text)]",
    purple: "border-[rgba(172,124,255,0.24)] bg-[rgba(172,124,255,0.12)] text-[rgb(198,169,255)]",
    slate: "border-[var(--app-line)] bg-[var(--app-card)] text-[var(--app-muted)]",
    teal: "border-[rgba(22,211,228,0.24)] bg-[rgba(22,211,228,0.12)] text-[var(--app-accent)]",
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

export function SectionHeading({
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
      <h2 className="text-3xl font-semibold tracking-tight text-[var(--app-text)] sm:text-4xl">{title}</h2>
      {description ? (
        <p className="mt-4 text-base leading-7 text-[var(--app-text-soft)] sm:text-lg">{description}</p>
      ) : null}
    </div>
  );
}

export function MetricCard({
  label,
  value,
  detail,
  icon,
}: {
  label: string;
  value: string;
  detail: string;
  icon: CleartifyIconName;
}) {
  return (
    <div className="rounded-[1.75rem] border border-[var(--app-line)] bg-[var(--app-card)] p-5 shadow-[0_18px_45px_rgba(0,0,0,0.16)]">
      <div className="flex items-center gap-2">
        <CleartifyIcon name={icon} className="h-4 w-4 text-[var(--app-accent)]" />
        <span className="text-xs uppercase tracking-[0.2em] text-[var(--app-muted)]">{label}</span>
      </div>
      <p className="mt-4 text-3xl font-semibold tracking-tight text-[var(--app-text)]">{value}</p>
      <p className="mt-2 text-sm leading-6 text-[var(--app-text-soft)]">{detail}</p>
    </div>
  );
}
