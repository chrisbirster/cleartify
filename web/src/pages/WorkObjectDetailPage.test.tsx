import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { Effect } from "effect";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { WorkObjectDetailPage } from "./WorkObjectDetailPage";
import { appDataFixture } from "../../test/fixtures";
import * as api from "../lib/api";

let updateWorkObjectStatusMock: any;
let updateWorkObjectPriorityMock: any;
let updateWorkObjectCoreFieldsMock: any;
let updateWorkObjectPrototypeMock: any;
let createDecisionForWorkObjectMock: any;
let createArtifactForWorkObjectMock: any;

describe("WorkObjectDetailPage", () => {
  beforeEach(() => {
    updateWorkObjectStatusMock = vi.spyOn(api, "updateWorkObjectStatus");
    updateWorkObjectPriorityMock = vi.spyOn(api, "updateWorkObjectPriority");
    updateWorkObjectCoreFieldsMock = vi.spyOn(api, "updateWorkObjectCoreFields");
    updateWorkObjectPrototypeMock = vi.spyOn(api, "updateWorkObjectPrototype");
    createDecisionForWorkObjectMock = vi.spyOn(api, "createDecisionForWorkObject");
    createArtifactForWorkObjectMock = vi.spyOn(api, "createArtifactForWorkObject");
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it("renders the selected work object with linked Cluster context, evidence, Decisions, and Artifacts", () => {
    render(
      <MemoryRouter initialEntries={["/work/wo_mobile_perf"]}>
        <Routes>
          <Route
            path="/work/:workObjectId"
            element={<WorkObjectDetailPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "Speed up mobile dashboard rendering" })).toBeInTheDocument();
    expect(screen.getAllByText("Improve dashboard responsiveness on small screens without losing analytical depth.")).toHaveLength(2);
    expect(screen.getByText("Work Object details")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Edit Work Object" })).toBeInTheDocument();
    expect(screen.getByText("Prototype context")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Edit prototype context" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Open prototype" })).toHaveAttribute("href", "https://www.figma.com/proto/cleartify-mobile-dashboard-optimization-v1");
    expect(screen.getByText("Status")).toBeInTheDocument();
    expect(screen.getByText("Priority")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "planning" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("button", { name: "high" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByText("Linked Cluster")).toBeInTheDocument();
    expect(screen.getByText("Mobile performance improvements")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Open Context Inbox" })).toHaveAttribute("href", "/inbox/clu_mobile_perf");
    expect(screen.getByText("Signal evidence")).toBeInTheDocument();
    expect(screen.getByText("PR #234 needs context from original spec")).toBeInTheDocument();
    expect(screen.getByText("Virtualize heavy dashboard sections")).toBeInTheDocument();
    expect(screen.getByText("Prioritize virtualization before a full chart rewrite.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Add Decision" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Add Artifact" })).toBeInTheDocument();
    expect(screen.getByText("Mobile optimization implementation plan")).toBeInTheDocument();
    expect(screen.getByText("Draft plan generated from mobile performance context.")).toBeInTheDocument();
  });

  it("renders a not found empty state", () => {
    render(
      <MemoryRouter initialEntries={["/work/missing"]}>
        <Routes>
          <Route
            path="/work/:workObjectId"
            element={<WorkObjectDetailPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "Work Object not found" })).toBeInTheDocument();
  });

  it("renders graceful empty states when there is no linked Cluster and no attached context panels", () => {
    render(
      <MemoryRouter initialEntries={["/work/wo_orphaned"]}>
        <Routes>
          <Route
            path="/work/:workObjectId"
            element={
              <WorkObjectDetailPage
                clusters={appDataFixture.clusters}
                workObjects={[
                  ...appDataFixture.workObjects,
                  {
                    id: "wo_orphaned",
                    title: "Standalone Work Object",
                    summary: "A detached seeded work object.",
                    goal: "Clarify the fallback states for orphaned execution work.",
                    status: "planning",
                    priority: "low",
                    ownerType: "user",
                    ownerId: "usr_demo",
                    decisions: null as unknown as [],
                    artifacts: null as unknown as [],
                    createdAt: "2026-03-01T09:00:00Z",
                    updatedAt: "2026-03-01T09:45:00Z",
                  },
                ]}
              />
            }
          />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByText("No linked Cluster")).toBeInTheDocument();
    expect(screen.getByText("No prototype attached")).toBeInTheDocument();
    expect(screen.getByText("No supporting Signals yet")).toBeInTheDocument();
    expect(screen.getByText("No Decisions yet")).toBeInTheDocument();
    expect(screen.getByText("No Artifacts yet")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "planning" })).toHaveClass("bg-zinc-400/16");
    expect(screen.getByRole("button", { name: "low" })).toHaveClass("bg-slate-300/16");
  });

  it("runs the core field mutation and reports the updated Work Object", async () => {
    const onWorkObjectUpdated = vi.fn();
    updateWorkObjectCoreFieldsMock.mockReturnValue(
      Effect.succeed({
        ...appDataFixture.workObjects[0],
        title: "Sharper mobile dashboard rendering",
        summary: "A tighter summary for the seeded work object.",
        goal: "Keep the mobile dashboard crisp under real data load.",
        updatedAt: "2026-03-29T16:20:00Z",
      }),
    );

    render(
      <MemoryRouter initialEntries={["/work/wo_mobile_perf"]}>
        <Routes>
          <Route
            path="/work/:workObjectId"
            element={
              <WorkObjectDetailPage
                workObjects={appDataFixture.workObjects}
                clusters={appDataFixture.clusters}
                onWorkObjectUpdated={onWorkObjectUpdated}
              />
            }
          />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Edit Work Object" }));
    fireEvent.change(screen.getByLabelText("Title"), { target: { value: "Sharper mobile dashboard rendering" } });
    fireEvent.change(screen.getByLabelText("Summary"), { target: { value: "A tighter summary for the seeded work object." } });
    fireEvent.change(screen.getByLabelText("Goal"), { target: { value: "Keep the mobile dashboard crisp under real data load." } });
    fireEvent.click(screen.getByRole("button", { name: "Save Work Object" }));

    await waitFor(() =>
      expect(updateWorkObjectCoreFieldsMock).toHaveBeenCalledWith("wo_mobile_perf", {
        title: "Sharper mobile dashboard rendering",
        summary: "A tighter summary for the seeded work object.",
        goal: "Keep the mobile dashboard crisp under real data load.",
      }),
    );
    await waitFor(() =>
      expect(onWorkObjectUpdated).toHaveBeenCalledWith(
        expect.objectContaining({
          id: "wo_mobile_perf",
          title: "Sharper mobile dashboard rendering",
          summary: "A tighter summary for the seeded work object.",
          goal: "Keep the mobile dashboard crisp under real data load.",
        }),
      ),
    );
  });

  it("cancels core field editing without saving", () => {
    render(
      <MemoryRouter initialEntries={["/work/wo_mobile_perf"]}>
        <Routes>
          <Route
            path="/work/:workObjectId"
            element={<WorkObjectDetailPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Edit Work Object" }));
    fireEvent.change(screen.getByLabelText("Title"), { target: { value: "Discarded title" } });
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));

    expect(updateWorkObjectCoreFieldsMock).not.toHaveBeenCalled();
    expect(screen.queryByLabelText("Title")).not.toBeInTheDocument();
    expect(screen.getAllByText("Speed up mobile dashboard rendering")).toHaveLength(2);
  });

  it("runs the status mutation and reports the updated Work Object", async () => {
    const onWorkObjectUpdated = vi.fn();
    updateWorkObjectStatusMock.mockReturnValue(
      Effect.succeed({
        ...appDataFixture.workObjects[0],
        status: "ready",
        updatedAt: "2026-03-29T16:30:00Z",
      }),
    );

    render(
      <MemoryRouter initialEntries={["/work/wo_mobile_perf"]}>
        <Routes>
          <Route
            path="/work/:workObjectId"
            element={
              <WorkObjectDetailPage
                workObjects={appDataFixture.workObjects}
                clusters={appDataFixture.clusters}
                onWorkObjectUpdated={onWorkObjectUpdated}
              />
            }
          />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole("button", { name: "ready" }));

    await waitFor(() => expect(updateWorkObjectStatusMock).toHaveBeenCalledWith("wo_mobile_perf", "ready"));
    await waitFor(() =>
      expect(onWorkObjectUpdated).toHaveBeenCalledWith(expect.objectContaining({ id: "wo_mobile_perf", status: "ready" })),
    );
  });

  it("shows a pending state while the status mutation is in flight", async () => {
    updateWorkObjectStatusMock.mockReturnValue(Effect.never as Effect.Effect<never, Error>);

    render(
      <MemoryRouter initialEntries={["/work/wo_mobile_perf"]}>
        <Routes>
          <Route
            path="/work/:workObjectId"
            element={<WorkObjectDetailPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole("button", { name: "ready" }));

    expect(screen.getByText("Saving status...")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "shipped" })).toBeDisabled();
  });

  it("shows a pending state while the core field mutation is in flight", () => {
    updateWorkObjectCoreFieldsMock.mockReturnValue(Effect.never as Effect.Effect<never, Error>);

    render(
      <MemoryRouter initialEntries={["/work/wo_mobile_perf"]}>
        <Routes>
          <Route
            path="/work/:workObjectId"
            element={<WorkObjectDetailPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Edit Work Object" }));
    fireEvent.click(screen.getByRole("button", { name: "Save Work Object" }));

    expect(screen.getByRole("button", { name: "Saving Work Object..." })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Cancel" })).toBeDisabled();
  });

  it("shows an error state when the status mutation fails", async () => {
    updateWorkObjectStatusMock.mockReturnValue(Effect.fail(new Error("boom")));

    render(
      <MemoryRouter initialEntries={["/work/wo_mobile_perf"]}>
        <Routes>
          <Route
            path="/work/:workObjectId"
            element={<WorkObjectDetailPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole("button", { name: "ready" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Unable to update Work Object status.");
  });

  it("shows an error state when the core field mutation fails", async () => {
    updateWorkObjectCoreFieldsMock.mockReturnValue(Effect.fail(new Error("boom")));

    render(
      <MemoryRouter initialEntries={["/work/wo_mobile_perf"]}>
        <Routes>
          <Route
            path="/work/:workObjectId"
            element={<WorkObjectDetailPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Edit Work Object" }));
    fireEvent.click(screen.getByRole("button", { name: "Save Work Object" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Unable to update Work Object details.");
  });

  it("runs the prototype mutation and reports the updated Work Object", async () => {
    const onWorkObjectUpdated = vi.fn();
    updateWorkObjectPrototypeMock.mockReturnValue(
      Effect.succeed({
        ...appDataFixture.workObjects[0],
        prototypeUrl: "https://www.figma.com/proto/cleartify-mobile-dashboard-optimization-v2",
        prototypeNote: "Updated prototype note for the first Prototype-to-Plan pass.",
        updatedAt: "2026-03-29T16:32:00Z",
      }),
    );

    render(
      <MemoryRouter initialEntries={["/work/wo_mobile_perf"]}>
        <Routes>
          <Route
            path="/work/:workObjectId"
            element={
              <WorkObjectDetailPage
                workObjects={appDataFixture.workObjects}
                clusters={appDataFixture.clusters}
                onWorkObjectUpdated={onWorkObjectUpdated}
              />
            }
          />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Edit prototype context" }));
    fireEvent.change(screen.getByLabelText("Prototype link"), { target: { value: "https://www.figma.com/proto/cleartify-mobile-dashboard-optimization-v2" } });
    fireEvent.change(screen.getByLabelText("Prototype note"), {
      target: { value: "Updated prototype note for the first Prototype-to-Plan pass." },
    });
    fireEvent.click(screen.getByRole("button", { name: "Save prototype context" }));

    await waitFor(() =>
      expect(updateWorkObjectPrototypeMock).toHaveBeenCalledWith("wo_mobile_perf", {
        prototypeUrl: "https://www.figma.com/proto/cleartify-mobile-dashboard-optimization-v2",
        prototypeNote: "Updated prototype note for the first Prototype-to-Plan pass.",
      }),
    );
    await waitFor(() =>
      expect(onWorkObjectUpdated).toHaveBeenCalledWith(
        expect.objectContaining({
          id: "wo_mobile_perf",
          prototypeUrl: "https://www.figma.com/proto/cleartify-mobile-dashboard-optimization-v2",
          prototypeNote: "Updated prototype note for the first Prototype-to-Plan pass.",
        }),
      ),
    );
  });

  it("cancels prototype editing without saving", () => {
    render(
      <MemoryRouter initialEntries={["/work/wo_mobile_perf"]}>
        <Routes>
          <Route
            path="/work/:workObjectId"
            element={<WorkObjectDetailPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Edit prototype context" }));
    fireEvent.change(screen.getByLabelText("Prototype note"), { target: { value: "Discarded note" } });
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));

    expect(updateWorkObjectPrototypeMock).not.toHaveBeenCalled();
    expect(screen.queryByLabelText("Prototype link")).not.toBeInTheDocument();
    expect(screen.queryByText("Discarded note")).not.toBeInTheDocument();
  });

  it("shows a pending state while the prototype mutation is in flight", () => {
    updateWorkObjectPrototypeMock.mockReturnValue(Effect.never as Effect.Effect<never, Error>);

    render(
      <MemoryRouter initialEntries={["/work/wo_mobile_perf"]}>
        <Routes>
          <Route
            path="/work/:workObjectId"
            element={<WorkObjectDetailPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Edit prototype context" }));
    fireEvent.click(screen.getByRole("button", { name: "Save prototype context" }));

    expect(screen.getByRole("button", { name: "Saving prototype context..." })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Cancel" })).toBeDisabled();
  });

  it("shows an error state when the prototype mutation fails", async () => {
    updateWorkObjectPrototypeMock.mockReturnValue(Effect.fail(new Error("boom")));

    render(
      <MemoryRouter initialEntries={["/work/wo_mobile_perf"]}>
        <Routes>
          <Route
            path="/work/:workObjectId"
            element={<WorkObjectDetailPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Edit prototype context" }));
    fireEvent.click(screen.getByRole("button", { name: "Save prototype context" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Unable to update Work Object prototype context.");
  });

  it("runs the priority mutation and reports the updated Work Object", async () => {
    const onWorkObjectUpdated = vi.fn();
    updateWorkObjectPriorityMock.mockReturnValue(
      Effect.succeed({
        ...appDataFixture.workObjects[0],
        priority: "urgent",
        updatedAt: "2026-03-29T16:35:00Z",
      }),
    );

    render(
      <MemoryRouter initialEntries={["/work/wo_mobile_perf"]}>
        <Routes>
          <Route
            path="/work/:workObjectId"
            element={
              <WorkObjectDetailPage
                workObjects={appDataFixture.workObjects}
                clusters={appDataFixture.clusters}
                onWorkObjectUpdated={onWorkObjectUpdated}
              />
            }
          />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole("button", { name: "urgent" }));

    await waitFor(() => expect(updateWorkObjectPriorityMock).toHaveBeenCalledWith("wo_mobile_perf", "urgent"));
    await waitFor(() =>
      expect(onWorkObjectUpdated).toHaveBeenCalledWith(expect.objectContaining({ id: "wo_mobile_perf", priority: "urgent" })),
    );
  });

  it("shows a pending state while the priority mutation is in flight", async () => {
    updateWorkObjectPriorityMock.mockReturnValue(Effect.never as Effect.Effect<never, Error>);

    render(
      <MemoryRouter initialEntries={["/work/wo_mobile_perf"]}>
        <Routes>
          <Route
            path="/work/:workObjectId"
            element={<WorkObjectDetailPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole("button", { name: "urgent" }));

    expect(screen.getByText("Saving priority…")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "medium" })).toBeDisabled();
  });

  it("shows an error state when the priority mutation fails", async () => {
    updateWorkObjectPriorityMock.mockReturnValue(Effect.fail(new Error("boom")));

    render(
      <MemoryRouter initialEntries={["/work/wo_mobile_perf"]}>
        <Routes>
          <Route
            path="/work/:workObjectId"
            element={<WorkObjectDetailPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole("button", { name: "urgent" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Unable to update Work Object priority.");
  });

  it("creates a new Decision and renders it immediately in the Decisions panel", async () => {
    const onWorkObjectUpdated = vi.fn();
    createDecisionForWorkObjectMock.mockReturnValue(
      Effect.succeed({
        id: "dec_new_scope",
        workObjectId: "wo_mobile_perf",
        title: "Keep the mobile optimization pass narrow",
        summary: "Start with the smallest high-signal change that improves responsiveness.",
        rationale: "That keeps the Work Object grounded in the clearest product judgment first.",
        createdAt: "2026-03-29T18:20:00Z",
      }),
    );

    render(
      <MemoryRouter initialEntries={["/work/wo_mobile_perf"]}>
        <Routes>
          <Route
            path="/work/:workObjectId"
            element={
              <WorkObjectDetailPage
                workObjects={appDataFixture.workObjects}
                clusters={appDataFixture.clusters}
                onWorkObjectUpdated={onWorkObjectUpdated}
              />
            }
          />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Add Decision" }));
    fireEvent.change(screen.getByLabelText("Decision title"), { target: { value: "Keep the mobile optimization pass narrow" } });
    fireEvent.change(screen.getByLabelText("Summary"), {
      target: { value: "Start with the smallest high-signal change that improves responsiveness." },
    });
    fireEvent.change(screen.getByLabelText("Rationale"), {
      target: { value: "That keeps the Work Object grounded in the clearest product judgment first." },
    });
    fireEvent.click(screen.getByRole("button", { name: "Save Decision" }));

    await waitFor(() =>
      expect(createDecisionForWorkObjectMock).toHaveBeenCalledWith("wo_mobile_perf", {
        title: "Keep the mobile optimization pass narrow",
        summary: "Start with the smallest high-signal change that improves responsiveness.",
        rationale: "That keeps the Work Object grounded in the clearest product judgment first.",
      }),
    );

    expect(await screen.findByText("Keep the mobile optimization pass narrow")).toBeInTheDocument();
    expect(screen.getByText("That keeps the Work Object grounded in the clearest product judgment first.")).toBeInTheDocument();
    await waitFor(() =>
      expect(onWorkObjectUpdated).toHaveBeenCalledWith(
        expect.objectContaining({
          id: "wo_mobile_perf",
          decisions: expect.arrayContaining([
            expect.objectContaining({
              id: "dec_new_scope",
              title: "Keep the mobile optimization pass narrow",
            }),
          ]),
        }),
      ),
    );
  });

  it("cancels Decision creation without saving", () => {
    render(
      <MemoryRouter initialEntries={["/work/wo_mobile_perf"]}>
        <Routes>
          <Route
            path="/work/:workObjectId"
            element={<WorkObjectDetailPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Add Decision" }));
    fireEvent.change(screen.getByLabelText("Decision title"), { target: { value: "Discarded Decision" } });
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));

    expect(createDecisionForWorkObjectMock).not.toHaveBeenCalled();
    expect(screen.queryByLabelText("Decision title")).not.toBeInTheDocument();
    expect(screen.queryByText("Discarded Decision")).not.toBeInTheDocument();
  });

  it("shows a pending state while the Decision create mutation is in flight", () => {
    createDecisionForWorkObjectMock.mockReturnValue(Effect.never as Effect.Effect<never, Error>);

    render(
      <MemoryRouter initialEntries={["/work/wo_mobile_perf"]}>
        <Routes>
          <Route
            path="/work/:workObjectId"
            element={<WorkObjectDetailPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Add Decision" }));
    fireEvent.change(screen.getByLabelText("Decision title"), { target: { value: "Pending Decision" } });
    fireEvent.change(screen.getByLabelText("Summary"), { target: { value: "Pending summary" } });
    fireEvent.change(screen.getByLabelText("Rationale"), { target: { value: "Pending rationale" } });
    fireEvent.click(screen.getByRole("button", { name: "Save Decision" }));

    expect(screen.getByRole("button", { name: "Saving Decision..." })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Cancel" })).toBeDisabled();
  });

  it("shows an error state when the Decision create mutation fails", async () => {
    createDecisionForWorkObjectMock.mockReturnValue(Effect.fail(new Error("boom")));

    render(
      <MemoryRouter initialEntries={["/work/wo_mobile_perf"]}>
        <Routes>
          <Route
            path="/work/:workObjectId"
            element={<WorkObjectDetailPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Add Decision" }));
    fireEvent.change(screen.getByLabelText("Decision title"), { target: { value: "Decision failure" } });
    fireEvent.change(screen.getByLabelText("Summary"), { target: { value: "Failure summary" } });
    fireEvent.change(screen.getByLabelText("Rationale"), { target: { value: "Failure rationale" } });
    fireEvent.click(screen.getByRole("button", { name: "Save Decision" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Unable to add Decision.");
  });

  it("creates a new Artifact and renders it immediately in the Artifacts panel", async () => {
    const onWorkObjectUpdated = vi.fn();
    createArtifactForWorkObjectMock.mockReturnValue(
      Effect.succeed({
        id: "art_new_note",
        workObjectId: "wo_mobile_perf",
        artifactType: "note",
        title: "Capture the loading-state output",
        body: "Keep the first pass Artifact short and tied to the Work Object outcome.",
        createdAt: "2026-03-29T19:20:00Z",
      }),
    );

    render(
      <MemoryRouter initialEntries={["/work/wo_mobile_perf"]}>
        <Routes>
          <Route
            path="/work/:workObjectId"
            element={
              <WorkObjectDetailPage
                workObjects={appDataFixture.workObjects}
                clusters={appDataFixture.clusters}
                onWorkObjectUpdated={onWorkObjectUpdated}
              />
            }
          />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Add Artifact" }));
    fireEvent.change(screen.getByLabelText("Artifact type"), { target: { value: "note" } });
    fireEvent.change(screen.getByLabelText("Artifact title"), { target: { value: "Capture the loading-state output" } });
    fireEvent.change(screen.getByLabelText("Body"), {
      target: { value: "Keep the first pass Artifact short and tied to the Work Object outcome." },
    });
    fireEvent.click(screen.getByRole("button", { name: "Save Artifact" }));

    await waitFor(() =>
      expect(createArtifactForWorkObjectMock).toHaveBeenCalledWith("wo_mobile_perf", {
        type: "note",
        title: "Capture the loading-state output",
        body: "Keep the first pass Artifact short and tied to the Work Object outcome.",
      }),
    );

    expect(await screen.findByText("Capture the loading-state output")).toBeInTheDocument();
    expect(screen.getByText("Keep the first pass Artifact short and tied to the Work Object outcome.")).toBeInTheDocument();
    await waitFor(() =>
      expect(onWorkObjectUpdated).toHaveBeenCalledWith(
        expect.objectContaining({
          id: "wo_mobile_perf",
          artifacts: expect.arrayContaining([
            expect.objectContaining({
              id: "art_new_note",
              title: "Capture the loading-state output",
            }),
          ]),
        }),
      ),
    );
  });

  it("cancels Artifact creation without saving", () => {
    render(
      <MemoryRouter initialEntries={["/work/wo_mobile_perf"]}>
        <Routes>
          <Route
            path="/work/:workObjectId"
            element={<WorkObjectDetailPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Add Artifact" }));
    fireEvent.change(screen.getByLabelText("Artifact title"), { target: { value: "Discarded Artifact" } });
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));

    expect(createArtifactForWorkObjectMock).not.toHaveBeenCalled();
    expect(screen.queryByLabelText("Artifact title")).not.toBeInTheDocument();
    expect(screen.queryByText("Discarded Artifact")).not.toBeInTheDocument();
  });

  it("shows a pending state while the Artifact create mutation is in flight", () => {
    createArtifactForWorkObjectMock.mockReturnValue(Effect.never as Effect.Effect<never, Error>);

    render(
      <MemoryRouter initialEntries={["/work/wo_mobile_perf"]}>
        <Routes>
          <Route
            path="/work/:workObjectId"
            element={<WorkObjectDetailPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Add Artifact" }));
    fireEvent.change(screen.getByLabelText("Artifact type"), { target: { value: "note" } });
    fireEvent.change(screen.getByLabelText("Artifact title"), { target: { value: "Pending Artifact" } });
    fireEvent.change(screen.getByLabelText("Body"), { target: { value: "Pending body" } });
    fireEvent.click(screen.getByRole("button", { name: "Save Artifact" }));

    expect(screen.getByRole("button", { name: "Saving Artifact..." })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Cancel" })).toBeDisabled();
  });

  it("shows an error state when the Artifact create mutation fails", async () => {
    createArtifactForWorkObjectMock.mockReturnValue(Effect.fail(new Error("boom")));

    render(
      <MemoryRouter initialEntries={["/work/wo_mobile_perf"]}>
        <Routes>
          <Route
            path="/work/:workObjectId"
            element={<WorkObjectDetailPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Add Artifact" }));
    fireEvent.change(screen.getByLabelText("Artifact type"), { target: { value: "note" } });
    fireEvent.change(screen.getByLabelText("Artifact title"), { target: { value: "Artifact failure" } });
    fireEvent.change(screen.getByLabelText("Body"), { target: { value: "Failure body" } });
    fireEvent.click(screen.getByRole("button", { name: "Save Artifact" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Unable to add Artifact.");
  });
});
