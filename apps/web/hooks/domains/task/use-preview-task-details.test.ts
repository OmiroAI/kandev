import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, expect, it, vi } from "vitest";
import { useStore } from "zustand";
import { createAppStore, type AppState } from "@/lib/state/store";
import { applyTaskOverviewPatch } from "@/lib/ws/handlers/task-overview";
import type { Task } from "@/lib/types/http";
import { usePreviewTaskDetails } from "./use-preview-task-details";

const api = vi.hoisted(() => ({ fetchTask: vi.fn() }));
let store: ReturnType<typeof createAppStore>;
vi.mock("@/lib/api/domains/kanban-api", async (original) => ({
  ...(await original<typeof import("@/lib/api/domains/kanban-api")>()),
  fetchTask: api.fetchTask,
}));
vi.mock("@/components/state-provider", () => ({
  useAppStoreApi: () => store,
  useAppStore: (selector: (state: AppState) => unknown) => useStore(store, selector),
}));

function task(id: string, description = "Question: What should search do?") {
  return {
    id,
    title: id,
    description,
    workspace_id: "ws",
    workflow_id: "wf",
    state: "CREATED",
    workflow_step_id: "ready",
    position: 0,
    priority: "medium",
  } as Task;
}

beforeEach(() => {
  vi.clearAllMocks();
  store = createAppStore();
  store.setState((s) => ({ workspaces: { ...s.workspaces, activeId: "ws" } }));
});

it("hydrates description data omitted by a lightweight board record", async () => {
  api.fetchTask.mockResolvedValue(task("a"));
  const { result } = renderHook(() => usePreviewTaskDetails("a"));
  await waitFor(() => expect(result.current.description).toContain("Question:"));
  expect(result.current.loading).toBe(false);
  expect(store.getState().taskOverview.byId.a.description).toContain("Question:");
});

it("cancels switched reads and rejects late responses for the old task", async () => {
  let resolveA!: (task: Task) => void;
  api.fetchTask.mockImplementation((id: string) =>
    id === "a"
      ? new Promise<Task>((r) => {
          resolveA = r;
        })
      : Promise.resolve(task(id)),
  );
  const { result, rerender } = renderHook(({ id }) => usePreviewTaskDetails(id), {
    initialProps: { id: "a" },
  });
  const signal = api.fetchTask.mock.calls[0][1].init.signal as AbortSignal;
  rerender({ id: "b" });
  await waitFor(() => expect(result.current.loading).toBe(false));
  await act(async () => {
    resolveA(task("a", "stale"));
  });
  expect(signal.aborted).toBe(true);
  expect(store.getState().taskOverview.byId.a).toBeUndefined();
  expect(result.current.description).toContain("Question:");
});

it("keeps a newer live description when an older HTTP read completes", async () => {
  let resolve!: (task: Task) => void;
  api.fetchTask.mockImplementation(
    () =>
      new Promise<Task>((r) => {
        resolve = r;
      }),
  );
  const { result } = renderHook(() => usePreviewTaskDetails("a"));
  act(() =>
    store.setState((s) =>
      applyTaskOverviewPatch(s, "a", { id: "a", description: "New question" }, "ws"),
    ),
  );
  await act(async () => {
    resolve(task("a", "Old question"));
  });
  await waitFor(() => expect(result.current.loading).toBe(false));
  expect(result.current.description).toBe("New question");
});

it("reports an unavailable read and permits a scoped retry", async () => {
  api.fetchTask.mockRejectedValueOnce(new Error("offline")).mockResolvedValueOnce(task("a"));
  const { result } = renderHook(() => usePreviewTaskDetails("a"));
  await waitFor(() => expect(result.current.error).toBe(true));
  act(() => result.current.retry());
  await waitFor(() => expect(result.current.description).toContain("Question:"));
  expect(api.fetchTask).toHaveBeenCalledTimes(2);
});
