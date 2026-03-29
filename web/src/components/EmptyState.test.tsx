import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { EmptyState } from "./EmptyState";

describe("EmptyState", () => {
  it("renders its content", () => {
    render(<EmptyState title="Phase 0 scaffold" description="More features land next." />);

    expect(screen.getByRole("heading", { name: "Phase 0 scaffold" })).toBeInTheDocument();
    expect(screen.getByText("More features land next.")).toBeInTheDocument();
  });
});
