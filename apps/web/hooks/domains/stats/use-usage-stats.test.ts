import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { StatsRange } from "@/lib/api/domains/stats-api";
import type { UsageSource, UsageStatsDTO } from "@/lib/types/http";

const fetchUsageStats = vi.hoisted(() => vi.fn());
vi.mock("@/lib/api/domains/stats-api", () => ({ fetchUsageStats }));
import { useUsageStats } from "./use-usage-stats";

function data(known = 12): UsageStatsDTO {
  return {
    totals: {
      known_tokens_total: known,
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
    next_task_cursor: "task:a",
    next_provider_cursor: "provider:a",
    daily_from: "2026-01-01",
    daily_until: "2026-01-03",
  };
}

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (error: Error) => void;
  const promise = new Promise<T>((ok, fail) => {
    resolve = ok;
    reject = fail;
  });
  return { promise, resolve, reject };
}

// Covers AC-COSTS-EXTERNAL-003.2 and AC-COSTS-EXTERNAL-003.3 request ownership.
describe("useUsageStats", () => {
  beforeEach(() => {
    fetchUsageStats.mockReset();
    fetchUsageStats.mockResolvedValue(data());
  });

  it("loads external usage independently of native Stats requests", async () => {
    const { result } = renderHook(() => useUsageStats("ws", "week", "external"));
    await waitFor(() => expect(result.current.data?.totals.known_tokens_total).toBe(12));
    expect(result.current.loading).toBe(false);
    expect(fetchUsageStats).toHaveBeenCalledWith(
      "ws",
      expect.objectContaining({ cache: "no-store" }),
      "week",
      expect.objectContaining({ source: "external" }),
    );
  });

  it.each([
    { key: "workspace", update: { workspace: "new" } },
    { key: "source", update: { source: "external" } },
    { key: "range", update: { range: "week" } },
  ] as const)("discards late results when $key changes", async ({ update }) => {
    const old = deferred<UsageStatsDTO>();
    fetchUsageStats.mockReturnValueOnce(old.promise).mockResolvedValueOnce(data(30));
    const initialProps: { workspace: string; range: StatsRange; source: UsageSource } = {
      workspace: "old",
      range: "month",
      source: "all",
    };
    const { result, rerender } = renderHook(
      ({ workspace, range, source }) => useUsageStats(workspace, range, source),
      { initialProps },
    );
    rerender({ ...initialProps, ...update });
    expect(result.current.data).toBeUndefined();
    await waitFor(() => expect(result.current.data?.totals.known_tokens_total).toBe(30));
    await act(async () => {
      old.resolve(data(99));
      await old.promise;
    });
    expect(result.current.data?.totals.known_tokens_total).toBe(30);
  });

  it("coalesces refreshes, preserves data on failure, and recovers on retry", async () => {
    const { result } = renderHook(() => useUsageStats("ws", "month", "all"));
    await waitFor(() => expect(result.current.data).toBeDefined());
    const refresh = deferred<UsageStatsDTO>();
    fetchUsageStats.mockReturnValueOnce(refresh.promise);
    act(() => {
      void result.current.refresh();
      void result.current.refresh();
    });
    expect(fetchUsageStats).toHaveBeenCalledTimes(2);
    await act(async () => {
      refresh.reject(new Error("offline"));
      await expect(refresh.promise).rejects.toThrow("offline");
    });
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).toBe(true);
    expect(result.current.data?.totals.known_tokens_total).toBe(12);
    fetchUsageStats.mockResolvedValueOnce(data(20));
    await act(() => result.current.refresh());
    expect(result.current.error).toBe(false);
    expect(result.current.data?.totals.known_tokens_total).toBe(20);
  });

  it("retries a failed task page and keeps provider pagination independent", async () => {
    const { result } = renderHook(() => useUsageStats("ws", "all", "external"));
    await waitFor(() => expect(result.current.data).toBeDefined());
    fetchUsageStats.mockRejectedValueOnce(new Error("page unavailable"));
    await act(() => result.current.nextTasks());
    expect(result.current.error).toBe(true);
    expect(result.current.data?.totals.known_tokens_total).toBe(12);
    fetchUsageStats.mockResolvedValueOnce(data(15));
    await act(() => result.current.refresh());
    expect(fetchUsageStats.mock.lastCall?.[3]).toMatchObject({
      source: "external",
      taskCursor: "task:a",
    });
    await act(() => result.current.nextProviders());
    expect(fetchUsageStats.mock.lastCall?.[3]).toMatchObject({
      source: "external",
      taskCursor: "task:a",
      providerCursor: "provider:a",
    });
  });

  it("aborts the owning request on unmount and never reads without a workspace", async () => {
    const pending = deferred<UsageStatsDTO>();
    fetchUsageStats.mockReturnValueOnce(pending.promise);
    const first = renderHook(() => useUsageStats("ws", "month", "native"));
    expect(fetchUsageStats).toHaveBeenCalledTimes(1);
    const signal: AbortSignal = fetchUsageStats.mock.calls[0][1].init.signal;
    first.unmount();
    expect(signal.aborted).toBe(true);
    await act(async () => {
      pending.resolve(data());
      await pending.promise;
    });
    const empty = renderHook(() => useUsageStats(undefined, "month", "all"));
    expect(empty.result.current.loading).toBe(false);
    expect(fetchUsageStats).toHaveBeenCalledTimes(1);
  });
});
