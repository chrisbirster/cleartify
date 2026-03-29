import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { PageHeader } from "./PageHeader";

describe("PageHeader", () => {
  it("renders a title and description", () => {
    render(<PageHeader title="Context Inbox" description="Review clustered context." />);

    expect(screen.getByRole("heading", { name: "Context Inbox" })).toBeInTheDocument();
    expect(screen.getByText("Review clustered context.")).toBeInTheDocument();
  });
});
