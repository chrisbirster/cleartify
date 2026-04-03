type OnboardingWalkthroughProps = {
  workspaceName?: string;
  pending: boolean;
  errorMessage?: string;
  onComplete: () => void;
};

const walkthroughSteps = [
  {
    title: "Context Inbox",
    description: "Start with Signals and Clusters so the reason for the work stays attached from the beginning.",
  },
  {
    title: "Work Objects",
    description: "Turn the highest-signal Cluster into a Work Object with a clear goal, priority, Decisions, and Artifacts.",
  },
  {
    title: "Prototype-to-Plan",
    description: "Attach a prototype link or note, then generate and refine a persisted draft plan before execution starts.",
  },
  {
    title: "Agent Workspace",
    description: "Use deterministic agent actions to turn the Work Object and draft plan into execution-ready Artifacts.",
  },
  {
    title: "Review and Releases",
    description: "Review the full execution context, then mark the Work Object release-ready so it enters the release queue.",
  },
];

export function OnboardingWalkthrough({
  workspaceName,
  pending,
  errorMessage,
  onComplete,
}: OnboardingWalkthroughProps) {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-6 py-10 text-foreground">
      <div className="w-full max-w-4xl rounded-[2rem] border border-border bg-card/90 p-8 shadow-[0_24px_80px_rgba(0,0,0,0.35)] backdrop-blur">
        <div className="max-w-2xl">
          <p className="text-xs uppercase tracking-[0.18em] text-muted-foreground">Welcome</p>
          <h1 className="mt-3 text-3xl font-semibold tracking-tight">Cleartify keeps product context attached to execution.</h1>
          <p className="mt-4 text-sm leading-6 text-muted-foreground">
            {workspaceName ? `${workspaceName} is ready to start with the hero flow. ` : ""}
            This walkthrough is the shortest path through the product: move from structured work, to a draft plan, to
            agent output, to review, to the release queue.
          </p>
        </div>

        <div className="mt-8 grid gap-4 md:grid-cols-2 xl:grid-cols-5">
          {walkthroughSteps.map((step, index) => (
            <section
              key={step.title}
              className="rounded-3xl border border-border/80 bg-black/20 p-5 shadow-[inset_0_1px_0_rgba(255,255,255,0.03)]"
            >
              <p className="text-[11px] uppercase tracking-[0.18em] text-muted-foreground">Step {index + 1}</p>
              <h2 className="mt-3 text-base font-semibold text-foreground">{step.title}</h2>
              <p className="mt-3 text-sm leading-6 text-muted-foreground">{step.description}</p>
            </section>
          ))}
        </div>

        <div className="mt-8 flex flex-col gap-3 border-t border-border/80 pt-6 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <p className="text-sm font-medium text-foreground">Start with the onboarding hero path.</p>
            <p className="text-sm text-muted-foreground">You’ll land back in the app shell with the walkthrough dismissed for this user.</p>
          </div>
          <div className="flex flex-col items-stretch gap-2 sm:items-end">
            <button
              type="button"
              onClick={onComplete}
              disabled={pending}
              className="rounded-full bg-foreground px-5 py-2 text-sm font-medium text-background transition hover:bg-foreground/90 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {pending ? "Finishing walkthrough..." : "Enter Cleartify"}
            </button>
            {errorMessage ? <p className="text-sm text-red-300">{errorMessage}</p> : null}
          </div>
        </div>
      </div>
    </div>
  );
}
