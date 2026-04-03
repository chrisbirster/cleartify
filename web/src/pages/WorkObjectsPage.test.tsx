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
    expect(screen.getByText("planning")).toHaveClass("bg-zinc-400/16");
    expect(screen.getByText("high priority")).toBeInTheDocument();
    expect(screen.getByText("in review")).toBeInTheDocument();
    expect(screen.getByText("medium priority")).toBeInTheDocument();
    expect(screen.getByText("Mobile performance improvements")).toBeInTheDocument();
    expect(screen.getByText("high severity")).toBeInTheDocument();
    expect(screen.getByText("Onboarding activation clarity")).toBeInTheDocument();
    expect(screen.getByText("medium severity")).toBeInTheDocument();
  });

  it("renders low priority with a visible filled treatment", () => {
    render(
      <MemoryRouter>
        <WorkObjectsPage
          workObjects={[
            ...appDataFixture.workObjects,
            {
              ...appDataFixture.workObjects[0],
              id: "wo_low_priority",
              title: "Low-priority cleanup",
              priority: "low",
            },
          ]}
          clusters={appDataFixture.clusters}
        />
      </MemoryRouter>,
    );

    expect(screen.getByText("low priority")).toHaveClass("bg-slate-300/16");
  });
});
