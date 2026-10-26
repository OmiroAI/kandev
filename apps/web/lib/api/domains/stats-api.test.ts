import { beforeEach, describe, expect, it, vi } from "vitest";

const fetchJsonMock = vi.hoisted(() => vi.fn());

vi.mock("../client", () => ({
  fetchJson: fetchJsonMock,
}));

import { fetchModelUsage, fetchUsageStats } from "./stats-api";

describe("fetchModelUsage", () => {
  beforeEach(() => fetchJsonMock.mockReset());

  it("requests the model-usage endpoint with the selected range", async () => {
    const response = [{ model: "opus", session_count: 2, turn_count: 4, total_duration_ms: 100 }];
    const options = { cache: "no-store" as const };
    fetchJsonMock.mockResolvedValue(response);

    await expect(fetchModelUsage("workspace-1", options, "month")).resolves.toEqual(response);
    expect(fetchJsonMock).toHaveBeenCalledWith(
      "/api/v1/workspaces/workspace-1/stats/model-usage?range=month",
      options,
    );
  });
});

// Covers AC-COSTS-EXTERNAL-003.2 and independent detail cursors.
describe("fetchUsageStats", () => {
  beforeEach(() => fetchJsonMock.mockReset());
  it("encodes workspace, range, source, and both independent cursors", async () => {
    const response = {
      totals: {
        known_tokens_total: 12,
        tokens_total: null,
        run_count: 2,
        native_event_count: 0,
        incomplete_run_count: 1,
        incomplete_native_event_count: 0,
        review_iterations: 2,
        qa_runs: 1,
        worker_duration_ms: 10,
      },
      daily: [],
      tasks: [],
      providers: [],
      next_task_cursor: "",
      next_provider_cursor: "",
      daily_from: "2026-01-01",
      daily_until: "2026-01-03",
    };
    const options = { cache: "no-store" as const };
    fetchJsonMock.mockResolvedValue(response);
    await expect(
      fetchUsageStats("workspace/a", options, "all", {
        source: "external",
        limit: 25,
        taskCursor: "task:a/b",
        providerCursor: "encoded cursor",
      }),
    ).resolves.toEqual(response);
    expect(fetchJsonMock).toHaveBeenCalledWith(
      "/api/v1/workspaces/workspace%2Fa/stats/usage?range=all&source=external&limit=25&task_cursor=task%3Aa%2Fb&provider_cursor=encoded+cursor",
      options,
    );
  });
});
