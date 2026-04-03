import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { Effect } from "effect";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { appDataFixture } from "../../test/fixtures";
import * as api from "../lib/api";
import { PlanPage } from "./PlanPage";

let loadDraftPlanForWorkObjectMock: any;
let generateDraftPlanForWorkObjectMock: any;
let updateDraftPlanForWorkObjectMock: any;

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
  createdAt: "2026-03-29T21:00:00Z",
  updatedAt: "2026-03-29T21:05:00Z",
};

describe("PlanPage", () => {
  beforeEach(() => {
    loadDraftPlanForWorkObjectMock = vi.spyOn(api, "loadDraftPlanForWorkObject");
    generateDraftPlanForWorkObjectMock = vi.spyOn(api, "generateDraftPlanForWorkObject");
    updateDraftPlanForWorkObjectMock = vi.spyOn(api, "updateDraftPlanForWorkObject");
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it("renders a not found empty state when the Work Object is missing", () => {
    render(
      <MemoryRouter initialEntries={["/plans/missing"]}>
        <Routes>
          <Route path="/plans/:workObjectId" element={<PlanPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "Work Object not found" })).toBeInTheDocument();
  });

  it("renders a persisted draft plan when one already exists", async () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.succeed(persistedDraftPlan));

    render(
      <MemoryRouter initialEntries={["/plans/wo_onboarding_refresh"]}>
        <Routes>
          <Route path="/plans/:workObjectId" element={<PlanPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByText(/This draft plan turns "Guided onboarding refresh"/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Refresh draft plan" })).toBeInTheDocument();
    expect(screen.getByText("Implementation steps")).toBeInTheDocument();
    expect(screen.getByText('Carry forward the linked Cluster context from "Onboarding activation clarity".')).toBeInTheDocument();
    expect(screen.getByText("Open prototype")).toHaveAttribute("href", "https://www.figma.com/proto/cleartify-guided-onboarding-refresh-v1");
    expect(screen.getByText("Onboarding activation clarity")).toBeInTheDocument();
  });

  it("renders a polished loading state while the persisted draft plan is loading", () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.never);

    render(
      <MemoryRouter initialEntries={["/plans/wo_onboarding_refresh"]}>
        <Routes>
          <Route path="/plans/:workObjectId" element={<PlanPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByRole("status")).toHaveTextContent("Loading persisted draft plan...");
    expect(screen.getByText("Loading the persisted draft so the current Work Object, prototype context, and next execution pass stay in one view.")).toBeInTheDocument();
  });

  it("renders an empty state before the first generation pass", async () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.succeed(null));

    render(
      <MemoryRouter initialEntries={["/plans/wo_mobile_perf"]}>
        <Routes>
          <Route path="/plans/:workObjectId" element={<PlanPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByRole("heading", { name: "No draft plan yet" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Generate draft plan" })).toBeInTheDocument();
  });

  it("runs the draft plan generation flow and renders the new persisted plan", async () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.succeed(null));
    generateDraftPlanForWorkObjectMock.mockReturnValue(
      Effect.succeed({
        ...persistedDraftPlan,
        workObjectId: "wo_mobile_perf",
        overview: 'This draft plan turns "Speed up mobile dashboard rendering" into executable work using prototype context.',
      }),
    );

    render(
      <MemoryRouter initialEntries={["/plans/wo_mobile_perf"]}>
        <Routes>
          <Route path="/plans/:workObjectId" element={<PlanPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByRole("heading", { name: "No draft plan yet" })).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Generate draft plan" }));

    await waitFor(() => expect(generateDraftPlanForWorkObjectMock).toHaveBeenCalledWith("wo_mobile_perf"));
    expect(await screen.findByText(/This draft plan turns "Speed up mobile dashboard rendering"/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Refresh draft plan" })).toBeInTheDocument();
  });

  it("shows a pending state while draft plan generation is in flight", async () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.succeed(null));
    generateDraftPlanForWorkObjectMock.mockReturnValue(Effect.never as Effect.Effect<never, Error>);

    render(
      <MemoryRouter initialEntries={["/plans/wo_mobile_perf"]}>
        <Routes>
          <Route path="/plans/:workObjectId" element={<PlanPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByRole("heading", { name: "No draft plan yet" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Generate draft plan" }));

    expect(screen.getByRole("button", { name: "Generating draft plan..." })).toBeDisabled();
  });

  it("shows an inline error when draft plan generation fails", async () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.succeed(null));
    generateDraftPlanForWorkObjectMock.mockReturnValue(Effect.fail(new Error("boom")));

    render(
      <MemoryRouter initialEntries={["/plans/wo_mobile_perf"]}>
        <Routes>
          <Route path="/plans/:workObjectId" element={<PlanPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByRole("heading", { name: "No draft plan yet" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Generate draft plan" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Unable to generate draft plan.");
  });

  it("shows an inline error when loading the persisted draft plan fails", async () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.fail(new Error("boom")));

    render(
      <MemoryRouter initialEntries={["/plans/wo_mobile_perf"]}>
        <Routes>
          <Route path="/plans/:workObjectId" element={<PlanPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByRole("alert")).toHaveTextContent("Unable to load draft plan.");
  });

  it("supports editing and saving a persisted draft plan", async () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.succeed(persistedDraftPlan));
    updateDraftPlanForWorkObjectMock.mockReturnValue(
      Effect.succeed({
        ...persistedDraftPlan,
        overview: "Edited onboarding draft overview.",
        implementationSteps: ["Review the edited prototype.", "Confirm the smallest onboarding slice."],
        edgeCases: ["The edited prototype may still evolve."],
        openQuestions: ["Which onboarding moments are still unresolved?"],
        updatedAt: "2026-03-29T21:30:00Z",
      }),
    );

    render(
      <MemoryRouter initialEntries={["/plans/wo_onboarding_refresh"]}>
        <Routes>
          <Route path="/plans/:workObjectId" element={<PlanPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByRole("button", { name: "Edit plan" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Edit plan" }));

    fireEvent.change(screen.getByLabelText("Overview"), { target: { value: "Edited onboarding draft overview." } });
    fireEvent.change(screen.getByLabelText("Implementation steps"), {
      target: { value: "Review the edited prototype.\nConfirm the smallest onboarding slice." },
    });
    fireEvent.change(screen.getByLabelText("Edge cases"), {
      target: { value: "The edited prototype may still evolve." },
    });
    fireEvent.change(screen.getByLabelText("Open questions"), {
      target: { value: "Which onboarding moments are still unresolved?" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Save plan" }));

    await waitFor(() =>
      expect(updateDraftPlanForWorkObjectMock).toHaveBeenCalledWith("wo_onboarding_refresh", {
        overview: "Edited onboarding draft overview.",
        implementationSteps: ["Review the edited prototype.", "Confirm the smallest onboarding slice."],
        edgeCases: ["The edited prototype may still evolve."],
        openQuestions: ["Which onboarding moments are still unresolved?"],
      }),
    );
    expect(await screen.findByText("Edited onboarding draft overview.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Edit plan" })).toBeInTheDocument();
  });

  it("cancels draft plan editing without saving", async () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.succeed(persistedDraftPlan));

    render(
      <MemoryRouter initialEntries={["/plans/wo_onboarding_refresh"]}>
        <Routes>
          <Route path="/plans/:workObjectId" element={<PlanPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByRole("button", { name: "Edit plan" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Edit plan" }));
    fireEvent.change(screen.getByLabelText("Overview"), { target: { value: "Discarded overview." } });
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));

    expect(updateDraftPlanForWorkObjectMock).not.toHaveBeenCalled();
    expect(screen.queryByLabelText("Overview")).not.toBeInTheDocument();
    expect(screen.getByText(/This draft plan turns "Guided onboarding refresh"/i)).toBeInTheDocument();
  });

  it("shows a pending state while draft plan edits are saving", async () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.succeed(persistedDraftPlan));
    updateDraftPlanForWorkObjectMock.mockReturnValue(Effect.never as Effect.Effect<never, Error>);

    render(
      <MemoryRouter initialEntries={["/plans/wo_onboarding_refresh"]}>
        <Routes>
          <Route path="/plans/:workObjectId" element={<PlanPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByRole("button", { name: "Edit plan" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Edit plan" }));
    fireEvent.click(screen.getByRole("button", { name: "Save plan" }));

    expect(screen.getByRole("button", { name: "Saving plan..." })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Cancel" })).toBeDisabled();
  });

  it("shows an inline error when draft plan edits fail", async () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.succeed(persistedDraftPlan));
    updateDraftPlanForWorkObjectMock.mockReturnValue(Effect.fail(new Error("boom")));

    render(
      <MemoryRouter initialEntries={["/plans/wo_onboarding_refresh"]}>
        <Routes>
          <Route path="/plans/:workObjectId" element={<PlanPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByRole("button", { name: "Edit plan" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Edit plan" }));
    fireEvent.click(screen.getByRole("button", { name: "Save plan" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Unable to save draft plan edits.");
  });
});
