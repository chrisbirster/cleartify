import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { AppShell } from "./AppShell";

describe("AppShell", () => {
  it("renders the shared shell and children", () => {
    render(
      <MemoryRouter initialEntries={["/"]}>
        <AppShell
          session={{ authenticated: true, user: { id: "u1", email: "demo@cleartify.com", displayName: "Demo User" } }}
          navItems={[{ label: "Today", to: "/" }]}
    >
      <div>Shell child</div>
    </AppShell>
  </MemoryRouter>,
    );

    expect(screen.getByRole("heading", { name: "Cleartify" })).toBeInTheDocument();
    expect(screen.getByText("Shell child")).toBeInTheDocument();
  });
});
