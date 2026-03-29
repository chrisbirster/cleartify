import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { SideNav } from "./SideNav";

describe("SideNav", () => {
  it("renders nav items and highlights the active route", () => {
    render(
      <MemoryRouter initialEntries={["/work"]}>
        <SideNav items={[{ label: "Today", to: "/" }, { label: "Work Objects", to: "/work" }]} />
      </MemoryRouter>,
    );

    expect(screen.getByRole("link", { name: "Today" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Work Objects" })).toHaveAttribute("aria-current", "page");
  });
});
