import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MemoryRouter } from "react-router-dom";

import { PrototypesPage } from "./PrototypesPage";
import { appDataFixture } from "../../test/fixtures";

describe("PrototypesPage", () => {
  it("renders attached prototype context from Work Objects", () => {
    render(
      <MemoryRouter>
        <PrototypesPage workObjects={appDataFixture.workObjects} />
      </MemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "Prototypes" })).toBeInTheDocument();
    expect(screen.getByText("Speed up mobile dashboard rendering")).toBeInTheDocument();
    expect(screen.getAllByRole("link", { name: "Open prototype" })).toHaveLength(2);
    expect(screen.getByText("The approved prototype covers five screens: welcome, workspace setup, teammate invite, empty state, and success confirmation.")).toBeInTheDocument();
  });

  it("renders an empty state when no Work Objects have prototype context", () => {
    render(
      <MemoryRouter>
        <PrototypesPage workObjects={[{ ...appDataFixture.workObjects[0], prototypeUrl: "", prototypeNote: "" }]} />
      </MemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "No prototypes attached" })).toBeInTheDocument();
  });
});
