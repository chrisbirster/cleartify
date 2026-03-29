import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { ReviewPage } from "./ReviewPage";

describe("ReviewPage", () => {
  it("renders the review scaffold", () => {
    render(<ReviewPage />);

    expect(screen.getByRole("heading", { name: "Review flow" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Review scaffold" })).toBeInTheDocument();
  });
});
