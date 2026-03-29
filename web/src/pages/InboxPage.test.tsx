import { cleanup, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, describe, expect, it } from "vitest";

import { InboxPage } from "./InboxPage";
import { appDataFixture } from "../../test/fixtures";

describe("InboxPage", () => {
  afterEach(() => {
    cleanup();
  });

  it("renders seeded clusters", () => {
    render(
      <MemoryRouter>
        <InboxPage clusters={appDataFixture.clusters} />
      </MemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "Context Inbox" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /Mobile performance improvements/i })).toBeInTheDocument();
    expect(screen.getAllByText("active")).toHaveLength(2);
    expect(screen.getByText(/PR #234 needs context from original spec/i)).toBeInTheDocument();
  });

  it("renders an empty state when there are no clusters", () => {
    render(
      <MemoryRouter>
        <InboxPage clusters={[]} />
      </MemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "No Clusters yet" })).toBeInTheDocument();
  });

  it("renders a singular Signal count when a Cluster has one Signal", () => {
    render(
      <MemoryRouter>
        <InboxPage
          clusters={[
            {
              ...appDataFixture.clusters[0],
              id: "clu_single_signal",
              signals: [appDataFixture.clusters[0].signals[0]],
            },
          ]}
        />
      </MemoryRouter>,
    );

    expect(screen.getByText("1 Signal")).toBeInTheDocument();
  });

  it("renders a plural Signal count when a Cluster has multiple Signals", () => {
    render(
      <MemoryRouter>
        <InboxPage
          clusters={[
            {
              ...appDataFixture.clusters[0],
              id: "clu_plural_signal",
              signals: [
                appDataFixture.clusters[0].signals[0],
                {
                  ...appDataFixture.signals[1],
                  id: "sig_plural_extra",
                },
              ],
            },
          ]}
        />
      </MemoryRouter>,
    );

    expect(screen.getByText("2 Signals")).toBeInTheDocument();
  });
});
