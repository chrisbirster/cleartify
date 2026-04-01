import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { Effect } from "effect";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { appDataFixture } from "../../test/fixtures";
import * as api from "../lib/api";
import { ReviewPage } from "./ReviewPage";

const persistedDraftPlan = {
  workObjectId: "wo_onboarding_refresh",
  overview:
    'This draft plan turns "Guided onboarding refresh" into executable work using the current summary, goal, and prototype context.',
  implementationSteps: [
    "Review the Work Object goal.",
    "Translate the current summary into the smallest useful implementation slice.",
  ],
  edgeCases: ["Prototype drift after review."],
  openQuestions: ["Which prototype states must ship first?"],
  createdAt: "2026-03-31T09:00:00Z",
  updatedAt: "2026-03-31T09:05:00Z",
};

describe("ReviewPage", () => {
  let loadDraftPlanForWorkObjectMock: any;
  let markWorkObjectReadyForReleaseMock: any;

  beforeEach(() => {
    loadDraftPlanForWorkObjectMock = vi.spyOn(api, "loadDraftPlanForWorkObject");
    markWorkObjectReadyForReleaseMock = vi.spyOn(api, "markWorkObjectReadyForRelease");
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it("renders the review summary with persisted draft plan, Decisions, and Artifacts", async () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.succeed(persistedDraftPlan));

    render(
      <MemoryRouter initialEntries={["/review/wo_onboarding_refresh"]}>
        <Routes>
          <Route
            path="/review/:workObjectId"
            element={<ReviewPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "Review" })).toBeInTheDocument();
    expect(await screen.findByText(/This draft plan turns "Guided onboarding refresh"/i)).toBeInTheDocument();
    expect(screen.getByText("Ship the five-screen onboarding slice first")).toBeInTheDocument();
    expect(screen.getByText("Onboarding scope review notes")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Mark release-ready" })).toBeInTheDocument();
  });

  it("renders a polished loading state while the persisted draft plan is loading", () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.never);

    render(
      <MemoryRouter initialEntries={["/review/wo_onboarding_refresh"]}>
        <Routes>
          <Route path="/review/:workObjectId" element={<ReviewPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByRole("status")).toHaveTextContent("Loading persisted draft plan...");
    expect(screen.getByText("Loading the persisted draft so this review reflects the latest planning context.")).toBeInTheDocument();
  });

  it("renders a draft plan prerequisite state when no persisted plan exists yet", async () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.succeed(null));

    render(
      <MemoryRouter initialEntries={["/review/wo_mobile_perf"]}>
        <Routes>
          <Route path="/review/:workObjectId" element={<ReviewPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByText("Persist a draft plan first")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Mark release-ready" })).toBeDisabled();
  });

  it("marks the Work Object ready for release and renders the persisted state immediately", async () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.succeed(persistedDraftPlan));
    markWorkObjectReadyForReleaseMock.mockReturnValue(
      Effect.succeed({
        ...appDataFixture.workObjects[1],
        releaseReadyAt: "2026-03-31T17:00:00Z",
        updatedAt: "2026-03-31T17:00:00Z",
      }),
    );
    const onWorkObjectUpdated = vi.fn();

    render(
      <MemoryRouter initialEntries={["/review/wo_onboarding_refresh"]}>
        <Routes>
          <Route
            path="/review/:workObjectId"
            element={
              <ReviewPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} onWorkObjectUpdated={onWorkObjectUpdated} />
            }
          />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByRole("button", { name: "Mark release-ready" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Mark release-ready" }));

    await waitFor(() => expect(markWorkObjectReadyForReleaseMock).toHaveBeenCalledWith("wo_onboarding_refresh"));
    expect(await screen.findByText(/Release-ready on Mar 31, 2026/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Release-ready" })).toBeDisabled();
    expect(onWorkObjectUpdated).toHaveBeenCalled();
  });

  it("shows a pending state while the release-ready action is in flight", async () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.succeed(persistedDraftPlan));
    markWorkObjectReadyForReleaseMock.mockReturnValue(Effect.never as Effect.Effect<never, Error>);

    render(
      <MemoryRouter initialEntries={["/review/wo_onboarding_refresh"]}>
        <Routes>
          <Route path="/review/:workObjectId" element={<ReviewPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />} />
        </Routes>
      </MemoryRouter>,
    );

    const actionButton = await screen.findByRole("button", { name: "Mark release-ready" });
    await waitFor(() => expect(actionButton).toBeEnabled());
    fireEvent.click(actionButton);

    expect(screen.getByRole("button", { name: "Marking release-ready..." })).toBeDisabled();
  });

  it("shows an inline error when the release-ready action fails", async () => {
    loadDraftPlanForWorkObjectMock.mockReturnValue(Effect.succeed(persistedDraftPlan));
    markWorkObjectReadyForReleaseMock.mockReturnValue(Effect.fail(new Error("boom")));

    render(
      <MemoryRouter initialEntries={["/review/wo_onboarding_refresh"]}>
        <Routes>
          <Route path="/review/:workObjectId" element={<ReviewPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(await screen.findByRole("button", { name: "Mark release-ready" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Mark release-ready" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Unable to mark the Work Object ready for release.");
  });
});
