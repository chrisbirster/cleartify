import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { PrototypeNewPage } from "./PrototypeNewPage";

describe("PrototypeNewPage", () => {
  it("renders the future bulk import scaffold", () => {
    render(<PrototypeNewPage />);

    expect(screen.getByRole("heading", { name: "Attach prototype" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Bulk import coming later" })).toBeInTheDocument();
  });
});
