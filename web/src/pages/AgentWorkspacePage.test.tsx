import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { AgentWorkspacePage } from "./AgentWorkspacePage";

describe("AgentWorkspacePage", () => {
  it("renders the agent workspace scaffold", () => {
    render(<AgentWorkspacePage />);

    expect(screen.getByRole("heading", { name: "Agent Workspace" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Agent execution scaffold" })).toBeInTheDocument();
  });
});
