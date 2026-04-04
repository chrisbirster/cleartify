import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { TodayPage } from "./TodayPage";
import { appDataFixture } from "../../test/fixtures";

describe("TodayPage", () => {
  it("renders the shell overview counts and links", () => {
    render(
      <MemoryRouter>
        <TodayPage
          projects={appDataFixture.projects}
          signals={appDataFixture.signals}
          clusters={appDataFixture.clusters}
          workObjects={appDataFixture.workObjects}
        />
      </MemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "Today" })).toBeInTheDocument();
    expect(screen.getByText("Projects")).toBeInTheDocument();
    expect(screen.getByText("Open Clusters")).toBeInTheDocument();
    expect(screen.getByText("Active Work")).toBeInTheDocument();
    expect(screen.getByText(String(appDataFixture.projects.length))).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Open Projects" })).toBeInTheDocument();
  });
});
