import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type Dispatch,
  type SetStateAction,
} from "react";
import { useForegroundRefresh } from "@/hooks/use-foreground-refresh";
import {
  fetchUsageStats,
  type StatsRange,
  type UsageStatsQuery,
} from "@/lib/api/domains/stats-api";
import type { UsageSource, UsageStatsDTO } from "@/lib/types/http";

export type UsageStatsState = {
  data?: UsageStatsDTO;
  loading: boolean;
  error: boolean;
  refresh: () => Promise<void>;
  nextTasks: () => Promise<void>;
  nextProviders: () => Promise<void>;
};

type Snapshot = { key: string; data?: UsageStatsDTO; loading: boolean; error: boolean };
type Owner = {
  key: string;
  controller: AbortController;
  query: UsageStatsQuery;
  failedQuery?: UsageStatsQuery;
  pending?: Promise<void>;
};
type Request = {
  owner: Owner;
  workspaceId: string;
  range: StatsRange;
  isCurrent: () => boolean;
  publish: Dispatch<SetStateAction<Snapshot>>;
};

function requestUsage(request: Request, query: UsageStatsQuery): Promise<void> {
  const { owner, workspaceId, range, isCurrent, publish } = request;
  if (!isCurrent()) return Promise.resolve();
  if (owner.pending) return owner.pending;
  publish((previous) => ({
    key: owner.key,
    data: previous.key === owner.key ? previous.data : undefined,
    loading: true,
    error: false,
  }));
  const pending = fetchUsageStats(
    workspaceId,
    { cache: "no-store", init: { signal: owner.controller.signal } },
    range,
    query,
  )
    .then((data) => {
      if (!isCurrent()) return;
      owner.query = query;
      owner.failedQuery = undefined;
      publish({ key: owner.key, data, loading: false, error: false });
    })
    .catch(() => {
      if (!isCurrent()) return;
      owner.failedQuery = query;
      publish((previous) => ({ ...previous, loading: false, error: true }));
    })
    .finally(() => {
      if (owner.pending === pending) owner.pending = undefined;
    });
  owner.pending = pending;
  return pending;
}

export function useUsageStats(
  workspaceId: string | undefined,
  range: StatsRange,
  source: UsageSource,
): UsageStatsState {
  const key = JSON.stringify([workspaceId, range, source]);
  const activeKey = useRef(key);
  activeKey.current = key;
  const ownerRef = useRef<Owner | null>(null);
  const [snapshot, publish] = useState<Snapshot>({
    key,
    loading: Boolean(workspaceId),
    error: false,
  });
  const load = useCallback(
    (patch?: UsageStatsQuery): Promise<void> => {
      const owner = ownerRef.current;
      if (!owner || owner.key !== key || !workspaceId) return Promise.resolve();
      const query = patch ? { ...owner.query, ...patch } : (owner.failedQuery ?? owner.query);
      return requestUsage(
        {
          owner,
          workspaceId,
          range,
          publish,
          isCurrent: () =>
            ownerRef.current === owner &&
            activeKey.current === owner.key &&
            !owner.controller.signal.aborted,
        },
        query,
      );
    },
    [key, workspaceId, range],
  );

  useEffect(() => {
    const owner: Owner = { key, controller: new AbortController(), query: { source } };
    ownerRef.current = owner;
    if (workspaceId) void load();
    else publish({ key, loading: false, error: false });
    return () => {
      owner.controller.abort();
      if (ownerRef.current === owner) ownerRef.current = null;
    };
  }, [key, source, workspaceId, load]);
  useForegroundRefresh(load, Boolean(workspaceId), key);

  const data = snapshot.key === key ? snapshot.data : undefined;
  const nextTasks = useCallback(
    () =>
      data?.next_task_cursor ? load({ taskCursor: data.next_task_cursor }) : Promise.resolve(),
    [data?.next_task_cursor, load],
  );
  const nextProviders = useCallback(
    () =>
      data?.next_provider_cursor
        ? load({ providerCursor: data.next_provider_cursor })
        : Promise.resolve(),
    [data?.next_provider_cursor, load],
  );
  return {
    data,
    loading: snapshot.key === key ? snapshot.loading : Boolean(workspaceId),
    error: snapshot.key === key && snapshot.error,
    refresh: load,
    nextTasks,
    nextProviders,
  };
}
