import { beforeEach, describe, expect, it, vi } from "vitest";

const renderMock = vi.fn();
const createRootMock = vi.fn(() => ({
  render: renderMock,
}));

vi.mock("react-dom/client", () => ({
  createRoot: createRootMock,
}));

describe("main entrypoint", () => {
  beforeEach(() => {
    document.body.innerHTML = '<div id="root"></div>';
    renderMock.mockReset();
    createRootMock.mockClear();
    vi.resetModules();
  });

  it("mounts the app into the root element", async () => {
    await import("./main");

    expect(createRootMock).toHaveBeenCalledWith(document.getElementById("root"));
    expect(renderMock).toHaveBeenCalledTimes(1);
  });

  it("throws when renderApp receives a missing root", async () => {
    const module = await import("./main");

    expect(() => module.renderApp(null)).toThrow("Root element not found");
  });
});
