import type { AppData } from "../src/lib/types";

export const appDataFixture: AppData = {
  session: {
    authenticated: true,
    user: {
      id: "usr_demo",
      email: "demo@cleartify.com",
      displayName: "Demo User",
    },
  },
  signals: [
    {
      id: "sig_pr_context",
      sourceType: "github",
      sourceId: "pr-234",
      title: "PR #234 needs context from original spec",
      body: "Code review is missing the original rationale for loading-state behavior.",
      url: "https://example.com/pr/234",
      createdAt: "2026-03-01T09:10:00Z",
    },
    {
      id: "sig_mobile_feedback",
      sourceType: "feedback",
      sourceId: "fb-100",
      title: "Mobile performance issues reported",
      body: "Dashboard feels slow on mobile during chart rendering.",
      createdAt: "2026-03-01T09:00:00Z",
    },
  ],
  clusters: [
    {
      id: "clu_mobile_perf",
      title: "Mobile performance improvements",
      summary: "Signals point to chart-heavy flows feeling slow on mobile devices.",
      severity: "high",
      status: "active",
      createdAt: "2026-03-01T09:00:00Z",
      updatedAt: "2026-03-01T09:30:00Z",
      signals: [
        {
          id: "sig_pr_context",
          sourceType: "github",
          sourceId: "pr-234",
          title: "PR #234 needs context from original spec",
          body: "Code review is missing the original rationale for loading-state behavior.",
          url: "https://example.com/pr/234",
          createdAt: "2026-03-01T09:10:00Z",
        },
      ],
    },
    {
      id: "clu_onboarding_refresh",
      title: "Onboarding refresh",
      summary: "Design review and prototype signals suggest a lighter onboarding flow.",
      severity: "medium",
      status: "active",
      createdAt: "2026-03-01T09:00:00Z",
      updatedAt: "2026-03-01T09:30:00Z",
      signals: [
        {
          id: "sig_mobile_feedback",
          sourceType: "feedback",
          sourceId: "fb-100",
          title: "Mobile performance issues reported",
          body: "Dashboard feels slow on mobile during chart rendering.",
          createdAt: "2026-03-01T09:00:00Z",
        },
      ],
    },
  ],
  workObjects: [
    {
      id: "wo_mobile_perf",
      clusterId: "clu_mobile_perf",
      title: "Speed up mobile dashboard rendering",
      summary: "Improve dashboard responsiveness on small screens without losing analytical depth.",
      goal: "Reduce perceived latency for mobile dashboard visits.",
      status: "planning",
      priority: "high",
      ownerType: "user",
      ownerId: "usr_demo",
      prototypeUrl: "https://example.com/prototype/mobile-perf",
      createdAt: "2026-03-01T09:00:00Z",
      updatedAt: "2026-03-01T09:45:00Z",
      decisions: [
        {
          id: "dec_mobile_virtualize",
          workObjectId: "wo_mobile_perf",
          title: "Virtualize heavy dashboard sections",
          summary: "Prioritize virtualization before a full chart rewrite.",
          rationale: "It improves mobile responsiveness quickly and reduces scope risk.",
          createdAt: "2026-03-01T09:50:00Z",
        },
      ],
      artifacts: [
        {
          id: "art_mobile_plan",
          workObjectId: "wo_mobile_perf",
          artifactType: "plan",
          title: "Mobile optimization implementation plan",
          body: "Draft plan generated from mobile performance context.",
          createdAt: "2026-03-01T09:55:00Z",
        },
      ],
    },
  ],
};
