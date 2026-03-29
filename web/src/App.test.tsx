import { render, screen } from "@testing-library/react";
import { Effect } from "effect";
import { afterEach, describe, expect, it, vi } from "vitest";

import { App } from "./App";
import { appDataFixture } from "../test/fixtures";
import { loadAppData } from "./lib/api";

vi.mock("./lib/api", () => ({
  loadAppData: vi.fn(),
}));

const loadAppDataMock = vi.mocked(loadAppData);

describe("App", () => {
  afterEach(() => {
    loadAppDataMock.mockReset();
    window.history.pushState({}, "", "/");
  });

  it("renders a loading state while bootstrap data is in flight", () => {
    loadAppDataMock.mockReturnValue(Effect.never);

    render(<App />);

    expect(screen.getByRole("status")).toHaveTextContent("Loading workspace");
  });

  it("renders an error state when bootstrap loading fails", async () => {
    loadAppDataMock.mockReturnValue(Effect.fail(new Error("boom")));

    render(<App />);

    expect(await screen.findByRole("heading", { name: "Unable to load Cleartify" })).toBeInTheDocument();
    expect(screen.getByText("Unable to load app data.")).toBeInTheDocument();
  });

  it("renders an auth boundary fallback when the session is unauthenticated", async () => {
    loadAppDataMock.mockReturnValue(Effect.succeed({
      ...appDataFixture,
      session: { authenticated: false },
    }));

    render(<App />);

    expect(await screen.findByRole("heading", { name: "Authentication required" })).toBeInTheDocument();
  });

  it("renders the app router after bootstrap data loads", async () => {
    window.history.pushState({}, "", "/app");
    loadAppDataMock.mockReturnValue(Effect.succeed(appDataFixture));

    render(<App />);

    expect(await screen.findByRole("heading", { name: "Today" })).toBeInTheDocument();
    expect(screen.getByText("Demo User")).toBeInTheDocument();
  });
});
