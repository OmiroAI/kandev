import { describe, expect, it } from "vitest";
import type { UsageStatsDTO } from "@/lib/types/http";
import { usageDailySeries } from "./usage-daily-series";

const FIRST_DAY = "2026-01-01";

function day(date: string, known: number, incomplete: number): UsageStatsDTO["daily"][number] {
  return {
    date,
    known_tokens_total: known,
    tokens_total: incomplete ? null : known,
    run_count: 1,
    native_event_count: 0,
    incomplete_run_count: incomplete,
    incomplete_native_event_count: 0,
    review_iterations: 0,
    qa_runs: 0,
    worker_duration_ms: 0,
  };
}

// Covers AC-COSTS-EXTERNAL-002.1 and AC-COSTS-EXTERNAL-003.1 UTC chart semantics.
describe("usageDailySeries", () => {
  it("fills UTC calendar gaps without inventing missing measurements", () => {
    expect(
      usageDailySeries([day(FIRST_DAY, 12, 1), day("2026-01-03", 0, 2)], FIRST_DAY, "2026-01-04"),
    ).toEqual([
      { date: FIRST_DAY, knownTokens: 12, incomplete: 1 },
      { date: "2026-01-02", knownTokens: 0, incomplete: 0 },
      { date: "2026-01-03", knownTokens: 0, incomplete: 2 },
    ]);
  });
  it.each([
    ["2024-01-01", FIRST_DAY],
    ["2026-02-30", "2026-03-04"],
    ["2026-01-02", FIRST_DAY],
    ["bad-date", FIRST_DAY],
  ])("rejects invalid or unbounded windows %s to %s", (from, until) => {
    expect(usageDailySeries([], from, until)).toEqual([]);
  });
});
