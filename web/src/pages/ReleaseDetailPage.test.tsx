import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { ReleaseDetailPage } from "./ReleaseDetailPage";

describe("ReleaseDetailPage", () => {
  it("renders the release detail scaffold", () => {
    render(<ReleaseDetailPage />);

    expect(screen.getByRole("heading", { name: "Release detail" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Release detail scaffold" })).toBeInTheDocument();
  });
});
