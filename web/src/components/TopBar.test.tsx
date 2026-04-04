import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { TopBar } from "./TopBar";

describe("TopBar", () => {
  it("renders the session user", () => {
    render(
      <MemoryRouter>
        <TopBar
          session={{
            authenticated: true,
            needsOnboarding: false,
            user: { id: "u1", email: "demo@cleartify.com", displayName: "Demo User" },
            workspace: { id: "wrk_1", slug: "demo-team", name: "Demo Team" },
          }}
        />
      </MemoryRouter>,
    );

    expect(screen.getByRole("link", { name: "Demo User" })).toHaveAttribute("href", "/me");
    expect(screen.getByRole("link", { name: "Demo Team" })).toHaveAttribute("href", "/team");
    expect(screen.getByRole("link", { name: "Sign out" })).toHaveAttribute("href", "/logout");
  });

  it("renders a signed out fallback", () => {
    render(
      <MemoryRouter>
        <TopBar session={{ authenticated: false, needsOnboarding: false }} />
      </MemoryRouter>,
    );

    expect(screen.getByText("Signed out")).toBeInTheDocument();
  });
});
