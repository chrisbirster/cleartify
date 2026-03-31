import { Effect } from "effect";
import { describe, expect, it, vi, afterEach } from "vitest";

import {
  createArtifactForWorkObject,
  createDecisionForWorkObject,
  createWorkObjectFromCluster,
  generateImplementationArtifactForWorkObject,
  generateTaskBreakdownArtifactForWorkObject,
  generateDraftPlanForWorkObject,
  jsonFetch,
  loadDraftPlanForWorkObject,
  loadAppData,
  markWorkObjectReadyForRelease,
  updateDraftPlanForWorkObject,
  updateWorkObjectCoreFields,
  updateWorkObjectPriority,
  updateWorkObjectPrototype,
  updateWorkObjectStatus,
} from "./api";

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

  it("loadDraftPlanForWorkObject returns null on 404 so persisted empty states can render", async () => {
    vi.stubGlobal(
      "fetch",
      fetchMock.mockResolvedValue({
        ok: false,
        status: 404,
      }),
    );

    await expect(Effect.runPromise(loadDraftPlanForWorkObject("wo_mobile_perf"))).resolves.toBeNull();
    expect(fetchMock).toHaveBeenCalledWith("/api/work-objects/wo_mobile_perf/plan", {
      headers: {
        Accept: "application/json",
      },
    });
  });

  it("generateDraftPlanForWorkObject posts to the work object draft plan endpoint", async () => {
    vi.stubGlobal(
      "fetch",
      fetchMock.mockResolvedValue({
        ok: true,
        json: async () => ({ workObjectId: "wo_mobile_perf", overview: "Generated overview" }),
      }),
    );

    await expect(Effect.runPromise(generateDraftPlanForWorkObject("wo_mobile_perf"))).resolves.toEqual({
      workObjectId: "wo_mobile_perf",
      overview: "Generated overview",
    });
    expect(fetchMock).toHaveBeenCalledWith("/api/work-objects/wo_mobile_perf/plan", {
      method: "POST",
      headers: expect.any(Headers),
    });
  });

  it("updateDraftPlanForWorkObject puts to the work object draft plan endpoint", async () => {
    vi.stubGlobal(
      "fetch",
      fetchMock.mockResolvedValue({
        ok: true,
        json: async () => ({ workObjectId: "wo_mobile_perf", overview: "Edited overview" }),
      }),
    );

    await expect(
      Effect.runPromise(
        updateDraftPlanForWorkObject("wo_mobile_perf", {
          overview: "Edited overview",
          implementationSteps: ["Edited step"],
          edgeCases: ["Edited edge case"],
          openQuestions: ["Edited question"],
        }),
      ),
    ).resolves.toEqual({
      workObjectId: "wo_mobile_perf",
      overview: "Edited overview",
    });
    expect(fetchMock).toHaveBeenCalledWith("/api/work-objects/wo_mobile_perf/plan", {
      method: "PUT",
      headers: expect.any(Headers),
      body: JSON.stringify({
        overview: "Edited overview",
        implementationSteps: ["Edited step"],
        edgeCases: ["Edited edge case"],
        openQuestions: ["Edited question"],
      }),
    });
  });

  it("updateWorkObjectStatus posts to the work object status endpoint", async () => {
    vi.stubGlobal(
      "fetch",
      fetchMock.mockResolvedValue({
        ok: true,
        json: async () => ({ id: "wo_mobile_perf", status: "ready" }),
      }),
    );

    await expect(Effect.runPromise(updateWorkObjectStatus("wo_mobile_perf", "ready"))).resolves.toEqual({
      id: "wo_mobile_perf",
      status: "ready",
    });
    expect(fetchMock).toHaveBeenCalledWith("/api/work-objects/wo_mobile_perf/status", {
      method: "POST",
      headers: expect.any(Headers),
      body: JSON.stringify({ status: "ready" }),
    });
  });

  it("updateWorkObjectCoreFields posts to the work object core endpoint", async () => {
    vi.stubGlobal(
      "fetch",
      fetchMock.mockResolvedValue({
        ok: true,
        json: async () => ({ id: "wo_mobile_perf", title: "Updated title" }),
      }),
    );

    await expect(
      Effect.runPromise(
        updateWorkObjectCoreFields("wo_mobile_perf", {
          title: "Updated title",
          summary: "Updated summary",
          goal: "Updated goal",
        }),
      ),
    ).resolves.toEqual({
      id: "wo_mobile_perf",
      title: "Updated title",
    });
    expect(fetchMock).toHaveBeenCalledWith("/api/work-objects/wo_mobile_perf/core", {
      method: "POST",
      headers: expect.any(Headers),
      body: JSON.stringify({
        title: "Updated title",
        summary: "Updated summary",
        goal: "Updated goal",
      }),
    });
  });

  it("updateWorkObjectPriority posts to the work object priority endpoint", async () => {
    vi.stubGlobal(
      "fetch",
      fetchMock.mockResolvedValue({
        ok: true,
        json: async () => ({ id: "wo_mobile_perf", priority: "urgent" }),
      }),
    );

    await expect(Effect.runPromise(updateWorkObjectPriority("wo_mobile_perf", "urgent"))).resolves.toEqual({
      id: "wo_mobile_perf",
      priority: "urgent",
    });
    expect(fetchMock).toHaveBeenCalledWith("/api/work-objects/wo_mobile_perf/priority", {
      method: "POST",
      headers: expect.any(Headers),
      body: JSON.stringify({ priority: "urgent" }),
    });
  });

  it("markWorkObjectReadyForRelease posts to the work object release-ready endpoint", async () => {
    vi.stubGlobal(
      "fetch",
      fetchMock.mockResolvedValue({
        ok: true,
        json: async () => ({ id: "wo_onboarding_refresh", releaseReadyAt: "2026-03-31T17:00:00Z" }),
      }),
    );

    await expect(Effect.runPromise(markWorkObjectReadyForRelease("wo_onboarding_refresh"))).resolves.toEqual({
      id: "wo_onboarding_refresh",
      releaseReadyAt: "2026-03-31T17:00:00Z",
    });
    expect(fetchMock).toHaveBeenCalledWith("/api/work-objects/wo_onboarding_refresh/release-ready", {
      method: "POST",
      headers: expect.any(Headers),
    });
  });

  it("updateWorkObjectPrototype posts to the work object prototype endpoint", async () => {
    vi.stubGlobal(
      "fetch",
      fetchMock.mockResolvedValue({
        ok: true,
        json: async () => ({ id: "wo_mobile_perf", prototypeUrl: "https://example.com/prototype/mobile-perf-v2" }),
      }),
    );

    await expect(
      Effect.runPromise(
        updateWorkObjectPrototype("wo_mobile_perf", {
          prototypeUrl: "https://example.com/prototype/mobile-perf-v2",
          prototypeNote: "Updated prototype note",
        }),
      ),
    ).resolves.toEqual({
      id: "wo_mobile_perf",
      prototypeUrl: "https://example.com/prototype/mobile-perf-v2",
    });
    expect(fetchMock).toHaveBeenCalledWith("/api/work-objects/wo_mobile_perf/prototype", {
      method: "POST",
      headers: expect.any(Headers),
      body: JSON.stringify({
        prototypeUrl: "https://example.com/prototype/mobile-perf-v2",
        prototypeNote: "Updated prototype note",
      }),
    });
  });

  it("createDecisionForWorkObject posts to the work object decisions endpoint", async () => {
    vi.stubGlobal(
      "fetch",
      fetchMock.mockResolvedValue({
        ok: true,
        json: async () => ({ id: "dec_1", title: "New Decision" }),
      }),
    );

    await expect(
      Effect.runPromise(
        createDecisionForWorkObject("wo_mobile_perf", {
          title: "New Decision",
          summary: "Keep the first scope small.",
          rationale: "That protects speed while product context is still fresh.",
        }),
      ),
    ).resolves.toEqual({
      id: "dec_1",
      title: "New Decision",
    });
    expect(fetchMock).toHaveBeenCalledWith("/api/work-objects/wo_mobile_perf/decisions", {
      method: "POST",
      headers: expect.any(Headers),
      body: JSON.stringify({
        title: "New Decision",
        summary: "Keep the first scope small.",
        rationale: "That protects speed while product context is still fresh.",
      }),
    });
  });

  it("generateImplementationArtifactForWorkObject posts to the work object implementation artifact endpoint", async () => {
    vi.stubGlobal(
      "fetch",
      fetchMock.mockResolvedValue({
        ok: true,
        json: async () => ({ id: "art_impl_1", artifactType: "implementation_brief" }),
      }),
    );

    await expect(Effect.runPromise(generateImplementationArtifactForWorkObject("wo_onboarding_refresh"))).resolves.toEqual({
      id: "art_impl_1",
      artifactType: "implementation_brief",
    });
    expect(fetchMock).toHaveBeenCalledWith("/api/work-objects/wo_onboarding_refresh/implementation-artifact", {
      method: "POST",
      headers: expect.any(Headers),
    });
  });

  it("generateTaskBreakdownArtifactForWorkObject posts to the work object task-breakdown artifact endpoint", async () => {
    vi.stubGlobal(
      "fetch",
      fetchMock.mockResolvedValue({
        ok: true,
        json: async () => ({ id: "art_tasks_1", artifactType: "task_breakdown" }),
      }),
    );

    await expect(Effect.runPromise(generateTaskBreakdownArtifactForWorkObject("wo_onboarding_refresh"))).resolves.toEqual({
      id: "art_tasks_1",
      artifactType: "task_breakdown",
    });
    expect(fetchMock).toHaveBeenCalledWith("/api/work-objects/wo_onboarding_refresh/task-breakdown-artifact", {
      method: "POST",
      headers: expect.any(Headers),
    });
  });

  it("createArtifactForWorkObject posts to the work object artifacts endpoint", async () => {
    vi.stubGlobal(
      "fetch",
      fetchMock.mockResolvedValue({
        ok: true,
        json: async () => ({ id: "art_1", artifactType: "note" }),
      }),
    );

    await expect(
      Effect.runPromise(
        createArtifactForWorkObject("wo_mobile_perf", {
          type: "note",
          title: "Mobile rollout note",
          body: "Document the first pass output clearly.",
        }),
      ),
    ).resolves.toEqual({
      id: "art_1",
      artifactType: "note",
    });
    expect(fetchMock).toHaveBeenCalledWith("/api/work-objects/wo_mobile_perf/artifacts", {
      method: "POST",
      headers: expect.any(Headers),
      body: JSON.stringify({
        type: "note",
        title: "Mobile rollout note",
        body: "Document the first pass output clearly.",
      }),
    });
  });
});
