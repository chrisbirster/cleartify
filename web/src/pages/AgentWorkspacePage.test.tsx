import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { Effect } from "effect";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { appDataFixture } from "../../test/fixtures";
import * as api from "../lib/api";
import { AgentWorkspacePage } from "./AgentWorkspacePage";

const persistedDraftPlan = {
  workObjectId: "wo_onboarding_refresh",
  overview:
    'This draft plan turns "Guided onboarding refresh" into executable work using the current summary, goal, and prototype context.',
  implementationSteps: [
    "Review the Work Object goal.",
    "Translate the current summary into the smallest useful implementation slice.",
    'Carry forward the linked Cluster context from "Onboarding activation clarity".',
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
    'Which Signals from "Onboarding activation clarity" should still shape acceptance criteria during implementation?',
  ],
  createdAt: "2026-03-31T09:00:00Z",
  updatedAt: "2026-03-31T09:05:00Z",
};

describe("AgentWorkspacePage", () => {
  let loadDraftPlanForWorkObjectMock: any;
  let generateImplementationArtifactForWorkObjectMock: any;
  let generateTaskBreakdownArtifactForWorkObjectMock: any;

  beforeEach(() => {
    loadDraftPlanForWorkObjectMock = vi.spyOn(api, "loadDraftPlanForWorkObject");
    generateImplementationArtifactForWorkObjectMock = vi.spyOn(api, "generateImplementationArtifactForWorkObject");
    generateTaskBreakdownArtifactForWorkObjectMock = vi.spyOn(api, "generateTaskBreakdownArtifactForWorkObject");
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it("renders the work object context and persisted draft plan summary", async () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.succeed(persistedDraftPlan));

    render(
      <MemoryRouter initialEntries={["/agents/wo_onboarding_refresh"]}>
        <Routes>
          <Route
            path="/agents/:workObjectId"
            element={<AgentWorkspacePage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "Agent Workspace" })).toBeInTheDocument();
    expect(await screen.findByText(/This draft plan turns "Guided onboarding refresh"/i)).toBeInTheDocument();
    expect(screen.getByText("Onboarding activation clarity")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Generate implementation brief" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Generate task breakdown" })).toBeInTheDocument();
  });

  it("renders a polished loading state while the persisted draft plan is loading", () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.never);

    render(
      <MemoryRouter initialEntries={["/agents/wo_onboarding_refresh"]}>
        <Routes>
          <Route
            path="/agents/:workObjectId"
            element={<AgentWorkspacePage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByRole("status")).toHaveTextContent("Loading persisted draft plan...");
    expect(screen.getByText("Loading the persisted draft so these actions stay grounded in the latest plan.")).toBeInTheDocument();
  });

  it("renders an empty draft plan state when none exists yet", async () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.succeed(null));

    render(
      <MemoryRouter initialEntries={["/agents/wo_mobile_perf"]}>
        <Routes>
          <Route
            path="/agents/:workObjectId"
            element={<AgentWorkspacePage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findAllByText("Persist a draft plan first")).toHaveLength(2);
    expect(screen.getByRole("button", { name: "Generate implementation brief" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Generate task breakdown" })).toBeDisabled();
  });

  it("runs the implementation artifact action and renders the new Artifact immediately", async () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.succeed(persistedDraftPlan));
    generateImplementationArtifactForWorkObjectMock.mockReturnValue(
      Effect.succeed({
        id: "art_impl_1",
        workObjectId: "wo_onboarding_refresh",
        artifactType: "implementation_brief",
        title: "Guided onboarding refresh implementation brief",
        body: "Overview\nDeterministic implementation brief body.",
        createdAt: "2026-03-31T09:15:00Z",
      }),
    );
    const onWorkObjectUpdated = vi.fn();

    render(
      <MemoryRouter initialEntries={["/agents/wo_onboarding_refresh"]}>
        <Routes>
          <Route
            path="/agents/:workObjectId"
            element={
              <AgentWorkspacePage
                workObjects={appDataFixture.workObjects}
                clusters={appDataFixture.clusters}
                onWorkObjectUpdated={onWorkObjectUpdated}
              />
            }
          />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByRole("button", { name: "Generate implementation brief" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Generate implementation brief" }));

    await waitFor(() => expect(generateImplementationArtifactForWorkObjectMock).toHaveBeenCalledWith("wo_onboarding_refresh"));
    expect(await screen.findByText("Guided onboarding refresh implementation brief")).toBeInTheDocument();
    expect(onWorkObjectUpdated).toHaveBeenCalled();
  });

  it("shows a pending state while the implementation artifact action is in flight", async () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.succeed(persistedDraftPlan));
    generateImplementationArtifactForWorkObjectMock.mockReturnValue(Effect.never as Effect.Effect<never, Error>);

    render(
      <MemoryRouter initialEntries={["/agents/wo_onboarding_refresh"]}>
        <Routes>
          <Route
            path="/agents/:workObjectId"
            element={<AgentWorkspacePage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByRole("button", { name: "Generate implementation brief" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Generate implementation brief" }));

    expect(screen.getByRole("button", { name: "Generating implementation brief..." })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Generate task breakdown" })).toBeDisabled();
  });

  it("shows an inline error when implementation artifact generation fails", async () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.succeed(persistedDraftPlan));
    generateImplementationArtifactForWorkObjectMock.mockReturnValue(Effect.fail(new Error("boom")));

    render(
      <MemoryRouter initialEntries={["/agents/wo_onboarding_refresh"]}>
        <Routes>
          <Route
            path="/agents/:workObjectId"
            element={<AgentWorkspacePage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByRole("button", { name: "Generate implementation brief" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Generate implementation brief" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Unable to generate implementation Artifact.");
  });

  it("runs the task-breakdown action and renders the new Artifact immediately", async () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.succeed(persistedDraftPlan));
    generateTaskBreakdownArtifactForWorkObjectMock.mockReturnValue(
      Effect.succeed({
        id: "art_tasks_1",
        workObjectId: "wo_onboarding_refresh",
        artifactType: "task_breakdown",
        title: "Guided onboarding refresh task breakdown",
        body: "Execution tasks\n1. Review the Work Object goal.",
        createdAt: "2026-03-31T09:16:00Z",
      }),
    );
    const onWorkObjectUpdated = vi.fn();

    render(
      <MemoryRouter initialEntries={["/agents/wo_onboarding_refresh"]}>
        <Routes>
          <Route
            path="/agents/:workObjectId"
            element={
              <AgentWorkspacePage
                workObjects={appDataFixture.workObjects}
                clusters={appDataFixture.clusters}
                onWorkObjectUpdated={onWorkObjectUpdated}
              />
            }
          />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByRole("button", { name: "Generate task breakdown" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Generate task breakdown" }));

    await waitFor(() => expect(generateTaskBreakdownArtifactForWorkObjectMock).toHaveBeenCalledWith("wo_onboarding_refresh"));
    expect(await screen.findByText("Guided onboarding refresh task breakdown")).toBeInTheDocument();
    expect(onWorkObjectUpdated).toHaveBeenCalled();
  });

  it("shows a pending state while the task-breakdown action is in flight", async () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.succeed(persistedDraftPlan));
    generateTaskBreakdownArtifactForWorkObjectMock.mockReturnValue(Effect.never as Effect.Effect<never, Error>);

    render(
      <MemoryRouter initialEntries={["/agents/wo_onboarding_refresh"]}>
        <Routes>
          <Route
            path="/agents/:workObjectId"
            element={<AgentWorkspacePage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByRole("button", { name: "Generate task breakdown" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Generate task breakdown" }));

    expect(screen.getByRole("button", { name: "Generating task breakdown..." })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Generate implementation brief" })).toBeDisabled();
  });

  it("shows an inline error when task-breakdown artifact generation fails", async () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.succeed(persistedDraftPlan));
    generateTaskBreakdownArtifactForWorkObjectMock.mockReturnValue(Effect.fail(new Error("boom")));

    render(
      <MemoryRouter initialEntries={["/agents/wo_onboarding_refresh"]}>
        <Routes>
          <Route
            path="/agents/:workObjectId"
            element={<AgentWorkspacePage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByRole("button", { name: "Generate task breakdown" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Generate task breakdown" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Unable to generate task-breakdown Artifact.");
  });

  it("shows an inline error when loading the persisted draft plan fails", async () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.fail(new Error("boom")));

    render(
      <MemoryRouter initialEntries={["/agents/wo_onboarding_refresh"]}>
        <Routes>
          <Route
            path="/agents/:workObjectId"
            element={<AgentWorkspacePage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByRole("alert")).toHaveTextContent("Unable to load draft plan.");
  });
});
