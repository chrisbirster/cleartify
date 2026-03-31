import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { StatePanel } from "./StatePanel";

describe("StatePanel", () => {
  it("renders a section state with eyebrow, title, and description", () => {
    render(<StatePanel eyebrow="Prototype-to-Plan" title="No draft plan yet" description="Generate a first pass." />);

    expect(screen.getByText("Prototype-to-Plan")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "No draft plan yet" })).toBeInTheDocument();
    expect(screen.getByText("Generate a first pass.")).toBeInTheDocument();
  });

  it("renders a loading state with status semantics", () => {
    render(
      <StatePanel
        description="Bringing the latest planning context into view."
        role="status"
        title="Loading persisted draft plan..."
        tone="loading"
      />,
    );

    expect(screen.getByRole("status")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Loading persisted draft plan..." })).toBeInTheDocument();
  });
});
