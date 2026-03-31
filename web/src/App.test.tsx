import { cleanup, render, screen } from "@testing-library/react";
import { Effect } from "effect";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { App } from "./App";
import { appDataFixture } from "../test/fixtures";
import * as api from "./lib/api";

let loadAppDataMock: any;
let updateWorkObjectStatusMock: any;
let updateWorkObjectPriorityMock: any;
let updateWorkObjectCoreFieldsMock: any;
let updateWorkObjectPrototypeMock: any;
let createDecisionForWorkObjectMock: any;
let createArtifactForWorkObjectMock: any;
let loadDraftPlanForWorkObjectMock: any;
let generateImplementationArtifactForWorkObjectMock: any;
let generateTaskBreakdownArtifactForWorkObjectMock: any;
let markWorkObjectReadyForReleaseMock: any;

const persistedDraftPlan = {
  workObjectId: "wo_onboarding_refresh",
  overview:
    'This draft plan turns "Prototype-to-Plan onboarding refresh" into executable work using the current summary, goal, and prototype context.',
  implementationSteps: [
    "Review the Work Object goal.",
    "Translate the current summary into the smallest useful implementation slice.",
    'Carry forward the linked Cluster context from "Onboarding refresh".',
    "Use the attached prototype link as the review reference while the first implementation pass is clarified.",
  ],
  edgeCases: [
    "The linked prototype may drift if the reference changes after planning starts.",
    "The prototype note may still leave some interaction details or handoff expectations unresolved.",
    "The linked Cluster is currently marked medium severity, so the implementation scope should stay proportional to that context.",
  ],
  openQuestions: [
    "Which prototype screens or states must ship in the first implementation pass?",
    "What implementation details are still implied, but not stated explicitly, in the prototype note?",
    'Which Signals from "Onboarding refresh" should still shape acceptance criteria during implementation?',
  ],
  createdAt: "2026-03-31T09:00:00Z",
  updatedAt: "2026-03-31T09:05:00Z",
};

describe("App", () => {
  beforeEach(() => {
    loadAppDataMock = vi.spyOn(api, "loadAppData");
    updateWorkObjectStatusMock = vi.spyOn(api, "updateWorkObjectStatus");
    updateWorkObjectPriorityMock = vi.spyOn(api, "updateWorkObjectPriority");
    updateWorkObjectCoreFieldsMock = vi.spyOn(api, "updateWorkObjectCoreFields");
    updateWorkObjectPrototypeMock = vi.spyOn(api, "updateWorkObjectPrototype");
    createDecisionForWorkObjectMock = vi.spyOn(api, "createDecisionForWorkObject");
    createArtifactForWorkObjectMock = vi.spyOn(api, "createArtifactForWorkObject");
    loadDraftPlanForWorkObjectMock = vi.spyOn(api, "loadDraftPlanForWorkObject");
    generateImplementationArtifactForWorkObjectMock = vi.spyOn(api, "generateImplementationArtifactForWorkObject");
    generateTaskBreakdownArtifactForWorkObjectMock = vi.spyOn(api, "generateTaskBreakdownArtifactForWorkObject");
    markWorkObjectReadyForReleaseMock = vi.spyOn(api, "markWorkObjectReadyForRelease");
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
    window.history.pushState({}, "", "/");
  });

  it("renders a loading state while bootstrap data is in flight", () => {
    loadAppDataMock.mockReturnValue(Effect.never);

    render(<App />);

    expect(screen.getByRole("status")).toHaveTextContent("Loading workspace");
  });

  it("renders an error state when bootstrap loading fails", async () => {
    loadAppDataMock.mockReturnValue(Effect.fail(new Error("boom")));

    render(<App />);

    expect(await screen.findByRole("heading", { name: "Unable to load Cleartify" })).toBeInTheDocument();
    expect(screen.getByText("Unable to load app data.")).toBeInTheDocument();
  });

  it("renders an auth boundary fallback when the session is unauthenticated", async () => {
    loadAppDataMock.mockReturnValue(Effect.succeed({
      ...appDataFixture,
      session: { authenticated: false },
    }));

    render(<App />);

    expect(await screen.findByRole("heading", { name: "Authentication required" })).toBeInTheDocument();
  });

  it("renders the app router after bootstrap data loads", async () => {
    window.history.pushState({}, "", "/app");
    loadAppDataMock.mockReturnValue(Effect.succeed(appDataFixture));

    render(<App />);

    expect(await screen.findByRole("heading", { name: "Today" })).toBeInTheDocument();
    expect(screen.getByText("Demo User")).toBeInTheDocument();
  });

  it("keeps the Work Object list in sync after a successful status change", async () => {
    const user = userEvent.setup();
    window.history.pushState({}, "", "/app/work/wo_mobile_perf");
    loadAppDataMock.mockReturnValue(Effect.succeed(appDataFixture));
    updateWorkObjectStatusMock.mockReturnValue(
      Effect.succeed({
        ...appDataFixture.workObjects[0],
        status: "shipped",
        updatedAt: "2026-03-29T17:00:00Z",
      }),
    );

    render(<App />);

    expect(await screen.findByRole("heading", { name: "Speed up mobile dashboard rendering" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "shipped" }));

    expect(await screen.findByRole("button", { name: "shipped" })).toHaveAttribute("aria-pressed", "true");

    await user.click(screen.getByRole("link", { name: "Work Objects" }));

    expect(await screen.findByRole("heading", { name: "Work Objects" })).toBeInTheDocument();
    expect(screen.getByText("shipped")).toBeInTheDocument();
  });

  it("keeps the Work Object list in sync after a successful priority change", async () => {
    const user = userEvent.setup();
    window.history.pushState({}, "", "/app/work/wo_mobile_perf");
    loadAppDataMock.mockReturnValue(Effect.succeed(appDataFixture));
    updateWorkObjectPriorityMock.mockReturnValue(
      Effect.succeed({
        ...appDataFixture.workObjects[0],
        priority: "urgent",
        updatedAt: "2026-03-29T17:05:00Z",
      }),
    );

    render(<App />);

    expect(await screen.findByRole("heading", { name: "Speed up mobile dashboard rendering" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "urgent" }));

    expect(await screen.findByRole("button", { name: "urgent" })).toHaveAttribute("aria-pressed", "true");

    await user.click(screen.getByRole("link", { name: "Work Objects" }));

    expect(await screen.findByRole("heading", { name: "Work Objects" })).toBeInTheDocument();
    expect(screen.getByText("urgent priority")).toBeInTheDocument();
  });

  it("keeps detail and list views in sync after a successful core field edit", async () => {
    const user = userEvent.setup();
    window.history.pushState({}, "", "/app/work/wo_mobile_perf");
    loadAppDataMock.mockReturnValue(Effect.succeed(appDataFixture));
    updateWorkObjectCoreFieldsMock.mockReturnValue(
      Effect.succeed({
        ...appDataFixture.workObjects[0],
        title: "Sharper mobile dashboard rendering",
        summary: "A tighter summary for the seeded work object.",
        goal: "Keep the mobile dashboard crisp under real data load.",
        updatedAt: "2026-03-29T17:10:00Z",
      }),
    );

    render(<App />);

    expect(await screen.findByRole("heading", { name: "Speed up mobile dashboard rendering" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Edit details" }));
    await user.clear(screen.getByLabelText("Title"));
    await user.type(screen.getByLabelText("Title"), "Sharper mobile dashboard rendering");
    await user.clear(screen.getByLabelText("Summary"));
    await user.type(screen.getByLabelText("Summary"), "A tighter summary for the seeded work object.");
    await user.clear(screen.getByLabelText("Goal"));
    await user.type(screen.getByLabelText("Goal"), "Keep the mobile dashboard crisp under real data load.");
    await user.click(screen.getByRole("button", { name: "Save changes" }));

    expect(await screen.findByRole("heading", { name: "Sharper mobile dashboard rendering" })).toBeInTheDocument();

    await user.click(screen.getByRole("link", { name: "Work Objects" }));

    expect(await screen.findByRole("heading", { name: "Work Objects" })).toBeInTheDocument();
    expect(screen.getByText("Sharper mobile dashboard rendering")).toBeInTheDocument();
    expect(screen.getByText("A tighter summary for the seeded work object.")).toBeInTheDocument();
    expect(screen.getByText("Mobile performance improvements")).toBeInTheDocument();
    expect(screen.getByText("high severity")).toBeInTheDocument();
  }, 10_000);

  it("keeps the Decisions panel in sync after creating a new Decision and navigating away and back", async () => {
    const user = userEvent.setup();
    window.history.pushState({}, "", "/app/work/wo_mobile_perf");
    loadAppDataMock.mockReturnValue(Effect.succeed(appDataFixture));
    createDecisionForWorkObjectMock.mockReturnValue(
      Effect.succeed({
        id: "dec_nav_sync",
        workObjectId: "wo_mobile_perf",
        title: "Keep the mobile performance pass tight",
        summary: "Capture the smallest product judgment that still improves responsiveness.",
        rationale: "That keeps the Work Object moving without broadening scope too early.",
        createdAt: "2026-03-29T18:45:00Z",
      }),
    );

    render(<App />);

    expect(await screen.findByRole("heading", { name: "Speed up mobile dashboard rendering" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Add Decision" }));
    await user.type(screen.getByLabelText("Decision title"), "Keep the mobile performance pass tight");
    await user.type(screen.getByLabelText("Summary"), "Capture the smallest product judgment that still improves responsiveness.");
    await user.type(screen.getByLabelText("Rationale"), "That keeps the Work Object moving without broadening scope too early.");
    await user.click(screen.getByRole("button", { name: "Save Decision" }));

    expect(await screen.findByText("Keep the mobile performance pass tight")).toBeInTheDocument();

    await user.click(screen.getByRole("link", { name: "Work Objects" }));
    expect(await screen.findByRole("heading", { name: "Work Objects" })).toBeInTheDocument();

    await user.click(screen.getByRole("link", { name: /Speed up mobile dashboard rendering/i }));
    expect(await screen.findByText("Keep the mobile performance pass tight")).toBeInTheDocument();
  }, 10_000);

  it("keeps prototype context in sync after a successful prototype attach update", async () => {
    const user = userEvent.setup();
    window.history.pushState({}, "", "/app/work/wo_mobile_perf");
    loadAppDataMock.mockReturnValue(Effect.succeed(appDataFixture));
    updateWorkObjectPrototypeMock.mockReturnValue(
      Effect.succeed({
        ...appDataFixture.workObjects[0],
        prototypeUrl: "https://example.com/prototype/mobile-perf-v2",
        prototypeNote: "Updated prototype note for the first Prototype-to-Plan pass.",
        updatedAt: "2026-03-29T20:15:00Z",
      }),
    );

    render(<App />);

    expect(await screen.findByRole("heading", { name: "Speed up mobile dashboard rendering" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Edit prototype" }));
    await user.clear(screen.getByLabelText("Prototype link"));
    await user.type(screen.getByLabelText("Prototype link"), "https://example.com/prototype/mobile-perf-v2");
    await user.clear(screen.getByLabelText("Prototype note"));
    await user.type(screen.getByLabelText("Prototype note"), "Updated prototype note for the first Prototype-to-Plan pass.");
    await user.click(screen.getByRole("button", { name: "Save prototype" }));

    expect(await screen.findByText("Updated prototype note for the first Prototype-to-Plan pass.")).toBeInTheDocument();

    await user.click(screen.getByRole("link", { name: "Prototypes" }));

    expect(await screen.findByRole("heading", { name: "Prototypes" })).toBeInTheDocument();
    expect(screen.getByText("Updated prototype note for the first Prototype-to-Plan pass.")).toBeInTheDocument();
  }, 10_000);

  it("keeps the Artifacts panel in sync after creating a new Artifact and navigating away and back", async () => {
    const user = userEvent.setup();
    window.history.pushState({}, "", "/app/work/wo_mobile_perf");
    loadAppDataMock.mockReturnValue(Effect.succeed(appDataFixture));
    createArtifactForWorkObjectMock.mockReturnValue(
      Effect.succeed({
        id: "art_nav_sync",
        workObjectId: "wo_mobile_perf",
        artifactType: "note",
        title: "Mobile rollout output note",
        body: "Capture the first output directly on the Work Object.",
        createdAt: "2026-03-29T19:45:00Z",
      }),
    );

    render(<App />);

    expect(await screen.findByRole("heading", { name: "Speed up mobile dashboard rendering" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Add Artifact" }));
    await user.type(screen.getByLabelText("Artifact type"), "note");
    await user.type(screen.getByLabelText("Artifact title"), "Mobile rollout output note");
    await user.type(screen.getByLabelText("Body"), "Capture the first output directly on the Work Object.");
    await user.click(screen.getByRole("button", { name: "Save Artifact" }));

    expect(await screen.findByText("Mobile rollout output note")).toBeInTheDocument();

    await user.click(screen.getByRole("link", { name: "Work Objects" }));
    expect(await screen.findByRole("heading", { name: "Work Objects" })).toBeInTheDocument();

    await user.click(screen.getByRole("link", { name: /Speed up mobile dashboard rendering/i }));
    expect(await screen.findByText("Mobile rollout output note")).toBeInTheDocument();
  });

  it("keeps generated implementation Artifacts in sync after running the mock agent action and opening the Work Object", async () => {
    const user = userEvent.setup();
    window.history.pushState({}, "", "/app/agents/wo_onboarding_refresh");
    loadAppDataMock.mockReturnValue(Effect.succeed(appDataFixture));
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.succeed(persistedDraftPlan));
    generateImplementationArtifactForWorkObjectMock.mockReturnValue(
      Effect.succeed({
        id: "art_impl_sync",
        workObjectId: "wo_onboarding_refresh",
        artifactType: "implementation_brief",
        title: "Prototype-to-Plan onboarding refresh implementation brief",
        body: "Overview\nDeterministic implementation brief body.",
        createdAt: "2026-03-31T09:25:00Z",
      }),
    );

    render(<App />);

    expect(await screen.findByRole("button", { name: "Generate implementation Artifact" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Generate implementation Artifact" }));

    expect(await screen.findByText("Prototype-to-Plan onboarding refresh implementation brief")).toBeInTheDocument();

    await user.click(screen.getByRole("link", { name: "Open Work Object" }));

    expect(await screen.findByRole("heading", { name: "Prototype-to-Plan onboarding refresh" })).toBeInTheDocument();
    expect(screen.getByText("Prototype-to-Plan onboarding refresh implementation brief")).toBeInTheDocument();
  });

  it("keeps generated task-breakdown Artifacts in sync after running the mock agent action and opening the Work Object", async () => {
    const user = userEvent.setup();
    window.history.pushState({}, "", "/app/agents/wo_onboarding_refresh");
    loadAppDataMock.mockReturnValue(Effect.succeed(appDataFixture));
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.succeed(persistedDraftPlan));
    generateTaskBreakdownArtifactForWorkObjectMock.mockReturnValue(
      Effect.succeed({
        id: "art_tasks_sync",
        workObjectId: "wo_onboarding_refresh",
        artifactType: "task_breakdown",
        title: "Prototype-to-Plan onboarding refresh task breakdown",
        body: "Execution tasks\n1. Review the Work Object goal.",
        createdAt: "2026-03-31T09:30:00Z",
      }),
    );

    render(<App />);

    expect(await screen.findByRole("button", { name: "Generate task-breakdown Artifact" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Generate task-breakdown Artifact" }));

    expect(await screen.findByText("Prototype-to-Plan onboarding refresh task breakdown")).toBeInTheDocument();

    await user.click(screen.getByRole("link", { name: "Open Work Object" }));

    expect(await screen.findByRole("heading", { name: "Prototype-to-Plan onboarding refresh" })).toBeInTheDocument();
    expect(screen.getByText("Prototype-to-Plan onboarding refresh task breakdown")).toBeInTheDocument();
  });

  it("keeps the Releases list in sync after marking a Work Object ready for release", async () => {
    const user = userEvent.setup();
    window.history.pushState({}, "", "/app/review/wo_onboarding_refresh");
    loadAppDataMock.mockReturnValue(Effect.succeed(appDataFixture));
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.succeed(persistedDraftPlan));
    markWorkObjectReadyForReleaseMock.mockReturnValue(
      Effect.succeed({
        ...appDataFixture.workObjects[1],
        releaseReadyAt: "2026-03-31T17:00:00Z",
        updatedAt: "2026-03-31T17:00:00Z",
      }),
    );

    render(<App />);

    expect(await screen.findByRole("button", { name: "Mark ready for release" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Mark ready for release" }));

    expect(await screen.findByText(/Ready for release on Mar 31, 2026/i)).toBeInTheDocument();

    await user.click(screen.getByRole("link", { name: "Releases" }));

    expect(await screen.findByRole("heading", { name: "Releases" })).toBeInTheDocument();
    expect(screen.getByText("Prototype-to-Plan onboarding refresh")).toBeInTheDocument();
    expect(screen.getByText(/Ready on Mar 31, 2026/i)).toBeInTheDocument();
  });
});
