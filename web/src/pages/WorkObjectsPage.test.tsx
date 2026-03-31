import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { WorkObjectsPage } from "./WorkObjectsPage";
import { appDataFixture } from "../../test/fixtures";

describe("WorkObjectsPage", () => {
  it("renders seeded work objects with clear status and priority styling", () => {
    render(
      <MemoryRouter>
        <WorkObjectsPage workObjects={appDataFixture.workObjects} clusters={appDataFixture.clusters} />
      </MemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "Work Objects" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /Speed up mobile dashboard rendering/i })).toBeInTheDocument();
    expect(screen.getByText("planning")).toBeInTheDocument();
    expect(screen.getByText("high priority")).toBeInTheDocument();
    expect(screen.getByText("in review")).toBeInTheDocument();
    expect(screen.getByText("medium priority")).toBeInTheDocument();
    expect(screen.getByText("Mobile performance improvements")).toBeInTheDocument();
    expect(screen.getByText("high severity")).toBeInTheDocument();
    expect(screen.getByText("Onboarding refresh")).toBeInTheDocument();
    expect(screen.getByText("medium severity")).toBeInTheDocument();
  });
});
