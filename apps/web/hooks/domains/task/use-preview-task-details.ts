import { useCallback, useEffect, useId, useState } from "react";
import { useAppStore, useAppStoreApi } from "@/components/state-provider";
import { fetchTask } from "@/lib/api/domains/kanban-api";
import { toKanbanTask } from "@/lib/kanban/map-task";
import { stateReadScopeIdentity } from "@/lib/state/shared-resource-reads";

export function usePreviewTaskDetails(taskId: string | null, enabled = true) {
  const store = useAppStoreApi();
  const scope = useAppStore(stateReadScopeIdentity);
  const description = useAppStore((s) =>
    taskId ? s.taskOverview.byId[taskId]?.description : undefined,
  );
  const ownerId = useId();
  const key = JSON.stringify([scope, taskId]);
  const [attempt, setAttempt] = useState(0);
  const [result, setResult] = useState<{ key: string; loading: boolean; error: boolean }>();
  useEffect(() => {
    if (!taskId || !enabled) return;
    let active = true;
    let finished = false;
    const controller = new AbortController();
    // i18n-exempt: private overview-retention owner identity
    const owner = `preview-description:${ownerId}:${taskId}`;
    const read = store.getState().beginTaskOverviewRead();
    const finish = () => {
      if (!finished) store.getState().finishTaskOverviewRead(read);
      finished = true;
    };
    setResult({ key, loading: true, error: false });
    fetchTask(taskId, { cache: "no-store", init: { signal: controller.signal } })
      .then((task) => {
        if (!active || stateReadScopeIdentity(store.getState()) !== scope) return;
        const accepted =
          task.id === taskId &&
          store.getState().retainTaskOverviews(owner, [toKanbanTask(task)], read);
        setResult({ key, loading: false, error: !accepted });
      })
      .catch(() => {
        if (active && stateReadScopeIdentity(store.getState()) === scope)
          setResult({ key, loading: false, error: true });
      })
      .finally(finish);
    return () => {
      active = false;
      controller.abort();
      finish();
      store.getState().releaseTaskOverviews(owner);
    };
  }, [taskId, enabled, scope, key, attempt, store, ownerId]);
  const current = result?.key === key ? result : undefined;
  const retry = useCallback(() => setAttempt((value) => value + 1), []);
  return {
    description,
    loading: Boolean(taskId && enabled && (!current || current.loading)),
    error: Boolean(taskId && enabled && current?.error),
    retry,
  };
}
