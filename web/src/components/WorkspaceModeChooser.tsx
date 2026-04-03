import type { WorkspaceMode } from "../lib/types";

type WorkspaceModeChooserProps = {
  workspaceName?: string;
  pendingMode?: WorkspaceMode;
  errorMessage?: string;
  onSelect: (mode: WorkspaceMode) => void;
};

const modeCards: Array<{
  mode: WorkspaceMode;
  eyebrow: string;
  title: string;
  description: string;
  bullets: string[];
  action: string;
}> = [
  {
    mode: "sample",
    eyebrow: "Sample workspace",
    title: "Start with the hero flow",
    description: "Use the polished sample workspace for evaluation, demos, and the end-to-end MVP story.",
    bullets: [
      "Opens the current Context Inbox hero path immediately.",
      "Keeps the sample review, plan, agent, and release loop intact.",
      "Clearly labeled so demo context never looks like live team data.",
    ],
    action: "Open sample workspace",
  },
  {
    mode: "live",
    eyebrow: "Live workspace",
    title: "Start with a clean workspace",
    description: "Begin with an empty live workspace so real teams are not dropped into demo-looking product data.",
    bullets: [
      "Starts clean and setup-ready with no sample hero content by default.",
      "Keeps the app shell available while the workspace is still being configured.",
      "Leaves Projects, connectors, and playbooks for the next live-workspace epics.",
    ],
    action: "Start live workspace",
  },
];

export function WorkspaceModeChooser({ workspaceName, pendingMode, errorMessage, onSelect }: WorkspaceModeChooserProps) {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-6 py-10 text-foreground">
      <div className="w-full max-w-5xl rounded-[2rem] border border-border bg-card/90 p-8 shadow-[0_24px_80px_rgba(0,0,0,0.35)] backdrop-blur">
        <div className="max-w-3xl">
          <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Workspace setup</p>
          <h1 className="mt-3 text-3xl font-semibold tracking-tight">Choose how this workspace should start.</h1>
          <p className="mt-4 text-sm leading-6 text-muted-foreground">
            {workspaceName ? `${workspaceName} is ready for its first mode selection. ` : ""}
            Pick a sample workspace for the Cleartify hero flow, or start live so the workspace stays clean and ready for real
            context.
          </p>
        </div>

        <div className="mt-8 grid gap-4 lg:grid-cols-2">
          {modeCards.map((card) => {
            const isPending = pendingMode === card.mode;

            return (
              <section
                key={card.mode}
                className="rounded-[1.75rem] border border-border/80 bg-black/20 p-6 shadow-[inset_0_1px_0_rgba(255,255,255,0.03)]"
              >
                <p className="text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground">{card.eyebrow}</p>
                <h2 className="mt-3 text-2xl font-semibold text-foreground">{card.title}</h2>
                <p className="mt-3 text-sm leading-6 text-muted-foreground">{card.description}</p>
                <ul className="mt-5 space-y-3 text-sm leading-6 text-foreground/90">
                  {card.bullets.map((bullet) => (
                    <li key={bullet} className="rounded-2xl border border-border/70 bg-background/40 px-4 py-3">
                      {bullet}
                    </li>
                  ))}
                </ul>
                <button
                  type="button"
                  onClick={() => onSelect(card.mode)}
                  disabled={Boolean(pendingMode)}
                  className={[
                    "mt-6 rounded-full px-5 py-2.5 text-sm font-medium transition disabled:cursor-not-allowed disabled:opacity-60",
                    card.mode === "sample"
                      ? "bg-foreground text-background hover:bg-foreground/90"
                      : "border border-border bg-background/70 text-foreground hover:border-border/70 hover:bg-background",
                  ].join(" ")}
                >
                  {isPending ? "Saving workspace mode..." : card.action}
                </button>
              </section>
            );
          })}
        </div>

        {errorMessage ? <p className="mt-6 text-sm text-red-300">{errorMessage}</p> : null}
      </div>
    </div>
  );
}
