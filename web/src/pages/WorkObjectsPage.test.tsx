import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { WorkObjectsPage } from "./WorkObjectsPage";
import { appDataFixture } from "../../test/fixtures";

describe("WorkObjectsPage", () => {
  it("renders seeded work objects", () => {
    render(
      <MemoryRouter>
        <WorkObjectsPage workObjects={appDataFixture.workObjects} />
      </MemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "Work Objects" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /Speed up mobile dashboard rendering/i })).toBeInTheDocument();
  });
});
