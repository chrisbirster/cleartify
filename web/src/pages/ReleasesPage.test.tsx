import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { appDataFixture } from "../../test/fixtures";
import { ReleasesPage } from "./ReleasesPage";

describe("ReleasesPage", () => {
  it("renders an empty state when no Work Objects are release-ready yet", () => {
    render(
      <MemoryRouter>
        <ReleasesPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />
      </MemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "Releases" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "No release-ready Work Objects yet" })).toBeInTheDocument();
  });

  it("renders a simple list of release-ready Work Objects", () => {
    render(
      <MemoryRouter>
        <ReleasesPage
          workObjects={[
            ...appDataFixture.workObjects.slice(0, 1),
            {
              ...appDataFixture.workObjects[1],
              releaseReadyAt: "2026-03-31T17:00:00Z",
            },
          ]}
          clusters={appDataFixture.clusters}
        />
      </MemoryRouter>,
    );

    expect(screen.getByText("Prototype-to-Plan onboarding refresh")).toBeInTheDocument();
    expect(screen.getByText(/Ready on Mar 31, 2026/i)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Open review" })).toBeInTheDocument();
  });
});
