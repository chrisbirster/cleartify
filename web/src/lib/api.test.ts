import { Effect } from "effect";
import { describe, expect, it, vi, afterEach } from "vitest";

import { createWorkObjectFromCluster, jsonFetch, loadAppData } from "./api";

const fetchMock = vi.fn();

describe("api helpers", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    fetchMock.mockReset();
  });

  it("jsonFetch returns parsed json", async () => {
    vi.stubGlobal("fetch", fetchMock.mockResolvedValue({
      ok: true,
      json: async () => ({ ok: true }),
    }));

    await expect(Effect.runPromise(jsonFetch<{ ok: boolean }>("/api/health"))).resolves.toEqual({ ok: true });
  });

  it("jsonFetch rejects on non-ok responses", async () => {
    vi.stubGlobal("fetch", fetchMock.mockResolvedValue({
      ok: false,
      status: 500,
    }));

    await expect(Effect.runPromise(jsonFetch("/api/health"))).rejects.toThrow("Request failed: 500");
  });

  it("loadAppData fetches the phase zero shell endpoints", async () => {
    vi.stubGlobal(
      "fetch",
      fetchMock
        .mockResolvedValueOnce({ ok: true, json: async () => ({ authenticated: true }) })
        .mockResolvedValueOnce({ ok: true, json: async () => [{ id: "sig_1" }] })
        .mockResolvedValueOnce({ ok: true, json: async () => [{ id: "clu_1" }] })
        .mockResolvedValueOnce({ ok: true, json: async () => [{ id: "wo_1" }] }),
    );

    await expect(Effect.runPromise(loadAppData())).resolves.toEqual({
      session: { authenticated: true },
      signals: [{ id: "sig_1" }],
      clusters: [{ id: "clu_1" }],
      workObjects: [{ id: "wo_1" }],
    });
  });

  it("createWorkObjectFromCluster posts to the cluster create endpoint", async () => {
    vi.stubGlobal(
      "fetch",
      fetchMock.mockResolvedValue({
        ok: true,
        json: async () => ({ id: "wo_new" }),
      }),
    );

    await expect(Effect.runPromise(createWorkObjectFromCluster("clu_mobile_perf"))).resolves.toEqual({ id: "wo_new" });
    expect(fetchMock).toHaveBeenCalledWith("/api/clusters/clu_mobile_perf/create-work-object", {
      method: "POST",
      headers: expect.any(Headers),
    });
  });
});
