import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { PlanPage } from "./PlanPage";

describe("PlanPage", () => {
  it("renders the generated plan scaffold", () => {
    render(<PlanPage />);

    expect(screen.getByRole("heading", { name: "Generated plan" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Plan generation scaffold" })).toBeInTheDocument();
  });
});
