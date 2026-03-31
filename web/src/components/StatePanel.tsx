import type { ReactNode } from "react";

type StatePanelProps = {
  title: string;
  description: string;
  eyebrow?: string;
  action?: ReactNode;
  role?: "status" | "alert";
  surface?: "page" | "section";
  tone?: "empty" | "loading";
};

export function StatePanel({
  title,
  description,
  eyebrow,
  action,
  role,
  surface = "section",
  tone = "empty",
}: StatePanelProps) {
  const isPageSurface = surface === "page";
  const containerClassName = isPageSurface
    ? "rounded-3xl border border-border bg-card/95 p-8 shadow-xl shadow-black/10"
    : tone === "loading"
      ? "rounded-2xl border border-border/80 bg-background/55 px-4 py-5"
      : "rounded-2xl border border-dashed border-border/80 bg-background/40 px-4 py-5";

  return (
    <div className={containerClassName} role={role}>
      {eyebrow ? <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">{eyebrow}</p> : null}
      <h2 className={eyebrow ? "mt-2 text-xl font-medium text-foreground" : "text-xl font-medium text-foreground"}>{title}</h2>
      <p className="mt-3 text-sm leading-6 text-muted-foreground">{description}</p>
      {action ? <div className="mt-4">{action}</div> : null}
    </div>
  );
}
