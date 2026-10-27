import { fetchJson, type ApiRequestOptions } from "../client";
import type {
  UsageStatsDTO,
  UsageSource,
  ModelUsageDTO,
  CompletedTaskActivityDTO,
  DailyActivityDTO,
  GitStatsDTO,
  GlobalStatsDTO,
  RepositoryStatsDTO,
  TaskStatsDTO,
} from "@/lib/types/http";

export type StatsRange = "week" | "month" | "all";

export type TaskStatsResponse = {
  task_stats: TaskStatsDTO[];
  task_stats_has_more: boolean;
};

function rangeQuery(range?: StatsRange): string {
  return range ? `?range=${encodeURIComponent(range)}` : "";
}

function statsUrl(workspaceId: string, section: string, range?: StatsRange): string {
  return `/api/v1/workspaces/${workspaceId}/stats/${section}${rangeQuery(range)}`;
}

export function fetchGlobalStats(
  workspaceId: string,
  options?: ApiRequestOptions,
  range?: StatsRange,
) {
  return fetchJson<GlobalStatsDTO>(statsUrl(workspaceId, "global", range), options);
}

export function fetchTaskStats(
  workspaceId: string,
  options?: ApiRequestOptions,
  range?: StatsRange,
) {
  return fetchJson<TaskStatsResponse>(statsUrl(workspaceId, "tasks", range), options);
}

export function fetchDailyActivity(
  workspaceId: string,
  options?: ApiRequestOptions,
  range?: StatsRange,
) {
  return fetchJson<DailyActivityDTO[]>(statsUrl(workspaceId, "daily-activity", range), options);
}

export function fetchCompletedActivity(
  workspaceId: string,
  options?: ApiRequestOptions,
  range?: StatsRange,
) {
  return fetchJson<CompletedTaskActivityDTO[]>(
    statsUrl(workspaceId, "completed-activity", range),
    options,
  );
}

export function fetchModelUsage(
  workspaceId: string,
  options?: ApiRequestOptions,
  range?: StatsRange,
) {
  return fetchJson<ModelUsageDTO[]>(statsUrl(workspaceId, "model-usage", range), options);
}

export function fetchRepositoryStats(
  workspaceId: string,
  options?: ApiRequestOptions,
  range?: StatsRange,
) {
  return fetchJson<RepositoryStatsDTO[]>(statsUrl(workspaceId, "repositories", range), options);
}

export function fetchGitStats(
  workspaceId: string,
  options?: ApiRequestOptions,
  range?: StatsRange,
) {
  return fetchJson<GitStatsDTO>(statsUrl(workspaceId, "git", range), options);
}

export type UsageStatsQuery = {
  source?: UsageSource;
  limit?: number;
  taskCursor?: string;
  providerCursor?: string;
};

export function fetchUsageStats(
  workspaceId: string,
  options?: ApiRequestOptions,
  range?: StatsRange,
  query: UsageStatsQuery = {},
) {
  const params = new URLSearchParams({
    range: range ?? "month",
    source: query.source ?? "all",
    limit: String(query.limit ?? 50),
  });
  if (query.taskCursor) params.set("task_cursor", query.taskCursor);
  if (query.providerCursor) params.set("provider_cursor", query.providerCursor);
  return fetchJson<UsageStatsDTO>(
    `${statsUrl(encodeURIComponent(workspaceId), "usage")}?${params.toString()}`,
    options,
  );
}
