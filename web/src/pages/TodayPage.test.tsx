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
          signals={appDataFixture.signals}
          clusters={appDataFixture.clusters}
          workObjects={appDataFixture.workObjects}
        />
      </MemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "Today" })).toBeInTheDocument();
    expect(screen.getAllByText(String(appDataFixture.signals.length))).toHaveLength(2);
    expect(screen.getByRole("link", { name: /Open Mobile performance improvements/i })).toBeInTheDocument();
  });
});
