import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { TopBar } from "./TopBar";

describe("TopBar", () => {
  it("renders the session user", () => {
    render(<TopBar session={{ authenticated: true, user: { id: "u1", email: "demo@cleartify.com", displayName: "Demo User" } }} />);

    expect(screen.getByText("Demo User")).toBeInTheDocument();
  });

  it("renders a signed out fallback", () => {
    render(<TopBar session={{ authenticated: false }} />);

    expect(screen.getByText("Signed out")).toBeInTheDocument();
  });
});
