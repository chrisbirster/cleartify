import { render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";

import { AppMemoryRouter, AppRouter } from "./router";
import { appDataFixture } from "../test/fixtures";

describe("router", () => {
  afterEach(() => {
    window.history.pushState({}, "", "/");
  });

  it("renders browser routes from the app origin root", () => {
    window.history.pushState({}, "", "/work");

    render(<AppRouter data={appDataFixture} />);

    expect(screen.getByRole("heading", { name: "Work Objects" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Context Inbox" })).toBeInTheDocument();
  });

  it("redirects legacy /app paths back to the root route shape", () => {
    render(<AppMemoryRouter data={appDataFixture} initialEntries={["/app/work"]} />);

    expect(screen.getAllByText("Speed up mobile dashboard rendering").length).toBeGreaterThan(0);
  });

  it("renders placeholder routes inside the shared shell", () => {
    render(<AppMemoryRouter data={appDataFixture} initialEntries={["/agents/wo_mobile_perf"]} />);

    expect(screen.getByRole("heading", { name: "Agent Workspace" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Summarize the next implementation pass" })).toBeInTheDocument();
  });

  it("renders the release detail route", () => {
    render(<AppMemoryRouter data={appDataFixture} initialEntries={["/releases/release-1"]} />);

    expect(screen.getByRole("heading", { name: "Release detail" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Release detail scaffold" })).toBeInTheDocument();
  });
});
