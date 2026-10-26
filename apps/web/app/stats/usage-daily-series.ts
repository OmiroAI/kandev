import type { UsageStatsDTO } from "@/lib/types/http";

export type UsageDayPoint = { date: string; knownTokens: number; incomplete: number };
const DAY_MS = 86_400_000;

function utcDate(value: string): number {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(value)) return NaN;
  const timestamp = Date.parse(`${value}T00:00:00Z`);
  return Number.isFinite(timestamp) && new Date(timestamp).toISOString().slice(0, 10) === value
    ? timestamp
    : NaN;
}

export function usageDailySeries(
  daily: UsageStatsDTO["daily"],
  from: string,
  until: string,
): UsageDayPoint[] {
  const start = utcDate(from);
  const days = (utcDate(until) - start) / DAY_MS;
  if (!Number.isInteger(days) || days < 1 || days > 365) return [];
  const observations = new Map(daily.map((day) => [day.date, day]));
  return Array.from({ length: days }, (_, index) => {
    const date = new Date(start + index * DAY_MS).toISOString().slice(0, 10);
    const observation = observations.get(date);
    return {
      date,
      knownTokens: observation?.known_tokens_total ?? 0,
      incomplete:
        (observation?.incomplete_run_count ?? 0) +
        (observation?.incomplete_native_event_count ?? 0),
    };
  });
}
