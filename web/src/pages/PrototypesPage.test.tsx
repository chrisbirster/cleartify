import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { PrototypesPage } from "./PrototypesPage";

describe("PrototypesPage", () => {
  it("renders the phase zero scaffold", () => {
    render(<PrototypesPage />);

    expect(screen.getByRole("heading", { name: "Prototypes" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Prototype-to-Plan scaffold" })).toBeInTheDocument();
  });
});
