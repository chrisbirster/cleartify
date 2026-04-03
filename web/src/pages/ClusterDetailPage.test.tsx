import { cleanup, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Effect } from "effect";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, describe, expect, it, vi } from "vitest";

import { browserRedirect, ClusterDetailPage, redirectToWorkObject } from "./ClusterDetailPage";
import { appDataFixture } from "../../test/fixtures";

describe("ClusterDetailPage", () => {
  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it("renders the selected cluster with metadata and source links", () => {
    render(
      <MemoryRouter initialEntries={["/inbox/clu_mobile_perf"]}>
        <Routes>
          <Route path="/inbox/:clusterId" element={<ClusterDetailPage clusters={appDataFixture.clusters} />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "Mobile performance improvements" })).toBeInTheDocument();
    expect(screen.getByText("PR #234 needs context from original spec")).toBeInTheDocument();
    expect(screen.getByText("active")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Open source" })).toHaveAttribute("href", "https://github.com/cleartifyhq/cleartify/pull/234");
    expect(screen.getByRole("button", { name: "Create Work Object" })).toBeInTheDocument();
  });

  it("renders a plural Signal count when a Cluster groups multiple Signals", () => {
    const clusters = [
      {
        ...appDataFixture.clusters[0],
        id: "clu_plural_signals",
        signals: [
          appDataFixture.clusters[0].signals[0],
          {
            ...appDataFixture.signals[1],
            id: "sig_plural_extra",
          },
        ],
      },
    ];

    render(
      <MemoryRouter initialEntries={["/inbox/clu_plural_signals"]}>
        <Routes>
          <Route path="/inbox/:clusterId" element={<ClusterDetailPage clusters={clusters} />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByText("2 Signals")).toBeInTheDocument();
  });

  it("omits source links for Signals without a URL", () => {
    const clusters = [
      {
        ...appDataFixture.clusters[0],
        id: "clu_no_url",
        title: "Cluster without source links",
        signals: [
          {
            ...appDataFixture.clusters[0].signals[0],
            id: "sig_no_url",
            title: "Signal without source link",
            url: undefined,
          },
        ],
      },
    ];

    render(
      <MemoryRouter initialEntries={["/inbox/clu_no_url"]}>
        <Routes>
          <Route path="/inbox/:clusterId" element={<ClusterDetailPage clusters={clusters} />} />
        </Routes>
      </MemoryRouter>,
    );

    const releaseSignalCard = screen.getByText("Signal without source link").closest("article");
    expect(releaseSignalCard).not.toBeNull();
    expect(within(releaseSignalCard as HTMLElement).queryByRole("link", { name: "Open source" })).not.toBeInTheDocument();
  });

  it("renders a not found empty state", () => {
    render(
      <MemoryRouter initialEntries={["/inbox/missing"]}>
        <Routes>
          <Route path="/inbox/:clusterId" element={<ClusterDetailPage clusters={appDataFixture.clusters} />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "Cluster not found" })).toBeInTheDocument();
  });

  it("shows a pending state while creating a work object", async () => {
    const user = userEvent.setup();

    render(
      <MemoryRouter initialEntries={["/inbox/clu_mobile_perf"]}>
        <Routes>
          <Route
            path="/inbox/:clusterId"
            element={<ClusterDetailPage clusters={appDataFixture.clusters} createWorkObject={() => Effect.never} />}
          />
        </Routes>
      </MemoryRouter>,
    );

    await user.click(screen.getByRole("button", { name: "Create Work Object" }));
    expect(screen.getByRole("button", { name: "Creating Work Object..." })).toBeDisabled();
  });

  it("shows an inline error state when creation fails", async () => {
    const user = userEvent.setup();

    render(
      <MemoryRouter initialEntries={["/inbox/clu_mobile_perf"]}>
        <Routes>
          <Route
            path="/inbox/:clusterId"
            element={
              <ClusterDetailPage
                clusters={appDataFixture.clusters}
                createWorkObject={() => Effect.fail(new Error("boom"))}
              />
            }
          />
        </Routes>
      </MemoryRouter>,
    );

    await user.click(screen.getByRole("button", { name: "Create Work Object" }));
    expect(await screen.findByText("Unable to create Work Object.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Create Work Object" })).toBeEnabled();
  });

  it("passes the new work object to the success callback after creation", async () => {
    const user = userEvent.setup();
    const onCreateSuccess = vi.fn();
    const createdWorkObject = {
      id: "wo_new",
      title: "Created",
      summary: "Summary",
      goal: "Goal",
      status: "planning",
      priority: "high",
      ownerType: "user",
      ownerId: "usr_demo",
      decisions: [],
      artifacts: [],
      createdAt: "2026-03-29T14:30:00Z",
      updatedAt: "2026-03-29T14:30:00Z",
    };
    const createWorkObject = vi.fn().mockReturnValue(Effect.succeed(createdWorkObject));

    render(
      <MemoryRouter initialEntries={["/inbox/clu_mobile_perf"]}>
        <Routes>
          <Route
            path="/inbox/:clusterId"
            element={
              <ClusterDetailPage
                clusters={appDataFixture.clusters}
                createWorkObject={createWorkObject}
                onCreateSuccess={onCreateSuccess}
              />
            }
          />
        </Routes>
      </MemoryRouter>,
    );

    await user.click(screen.getByRole("button", { name: "Create Work Object" }));

    await waitFor(() => {
      expect(createWorkObject).toHaveBeenCalledWith("clu_mobile_perf");
      expect(onCreateSuccess).toHaveBeenCalledWith(expect.objectContaining({ id: "wo_new" }));
    });
  });

  it("uses a hard redirect helper for the default success path", () => {
    const assignSpy = vi.spyOn(browserRedirect, "assign").mockImplementation(() => undefined);

    redirectToWorkObject({
      id: "wo_new",
      title: "Created",
      summary: "Summary",
      goal: "Goal",
      status: "planning",
      priority: "high",
      ownerType: "user",
      ownerId: "usr_demo",
      decisions: [],
      artifacts: [],
      createdAt: "2026-03-29T14:30:00Z",
      updatedAt: "2026-03-29T14:30:00Z",
    });

    expect(assignSpy).toHaveBeenCalledWith("/work/wo_new");
  });

  it("uses the browser location assign helper", () => {
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => undefined);

    browserRedirect.assign("/work/wo_new");

    expect(errorSpy).toHaveBeenCalled();
  });
});
