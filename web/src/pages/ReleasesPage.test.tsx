import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { ReleasesPage } from "./ReleasesPage";

describe("ReleasesPage", () => {
  it("renders the releases scaffold", () => {
    render(<ReleasesPage />);

    expect(screen.getByRole("heading", { name: "Releases" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Release list scaffold" })).toBeInTheDocument();
  });
});
