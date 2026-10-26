import { randomUUID } from "node:crypto";
import { expect, type Page, type Route } from "@playwright/test";
import type { BackendContext } from "../../fixtures/backend";
import type { SeedData } from "../../fixtures/test-base";
import type { ApiClient } from "../../helpers/api-client";
import type { UsageStatsDTO } from "../../../lib/types/http";

type SeedOptions = {
  testPage: Page;
  backend: BackendContext;
  apiClient: ApiClient;
  seedData: SeedData;
  tasks?: number;
};

async function dropCommittedResponse(page: Page, path: string, body: object) {
  let committedId = "";
  const handler = async (route: Route) => {
    const response = await route.fetch();
    expect(response.status()).toBe(201);
    committedId = (await response.json()).id;
    await route.abort("failed");
  };
  await page.route(`**${path}`, handler);
  try {
    const dropped = await page.evaluate(
      async ({ path, body }) => {
        try {
          await fetch(path, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(body),
          });
          return false;
        } catch {
          return true;
        }
      },
      { path, body },
    );
    expect(dropped).toBe(true);
    expect(committedId).not.toBe("");
  } finally {
    await page.unroute(`**${path}`, handler);
  }
  return committedId;
}

export async function seedExternalUsage({
  testPage,
  backend,
  apiClient,
  seedData,
  tasks = 1,
}: SeedOptions) {
  for (let index = 0; index < tasks; index++) {
    await apiClient.createTask(seedData.workspaceId, `External usage ${index}`, {
      workflow_id: seedData.workflowId,
      workflow_step_id: seedData.startStepId,
      repository_ids: [seedData.repositoryId],
    });
  }
  const usageURL = `${backend.baseUrl}/api/v1/workspaces/${seedData.workspaceId}/stats/usage?range=all`;
  const response = await testPage.request.get(usageURL);
  expect(response.ok()).toBe(true);
  const initial: UsageStatsDTO = await response.json();
  const task = initial.tasks[0];
  expect(task).toBeDefined();
  await apiClient.updateTaskTitle(task.task_id, "External usage task");
  await testPage.goto(`/stats?workspaceId=${seedData.workspaceId}`);
  const occurrence = new Date(Date.now() - 2 * 86_400_000);
  occurrence.setUTCHours(1, 30, 0, 0);
  const offset =
    new Date(occurrence.getTime() - 2 * 3_600_000).toISOString().slice(0, 19) + "-02:00";
  const identity = {
    schema_version: 1,
    source: "kodi",
    source_instance: randomUUID(),
    occurred_at: offset,
  };
  const event = {
    ...identity,
    event_id: "worker:1",
    kind: "worker_run",
    phase: "coding",
    outcome: "succeeded",
    provider: "fixture-provider",
    model: "fixture-model",
    input_tokens: 100,
    output_tokens: 20,
    cached_input_tokens: 40,
  };
  const path = `/api/v1/tasks/${task.task_id}/external-activity`;
  const committedId = await dropCommittedResponse(testPage, path, event);
  const retry = await testPage.request.post(backend.baseUrl + path, { data: event });
  expect(retry.status()).toBe(200);
  expect((await retry.json()).id).toBe(committedId);
  const partial = await testPage.request.post(backend.baseUrl + path, {
    data: {
      ...identity,
      event_id: "worker:2",
      kind: "worker_run",
      phase: "review",
      outcome: "failed",
      provider: "fixture-provider",
      model: "fixture-model",
      input_tokens: 0,
    },
  });
  expect(partial.status()).toBe(201);
  for (const [event_id, kind, outcome] of [
    ["review:1", "review_completed", "rejected"],
    ["review:2", "review_completed", "approved"],
    ["qa:1", "qa_completed", "failed"],
  ]) {
    const result = await testPage.request.post(backend.baseUrl + path, {
      data: {
        ...identity,
        event_id,
        kind,
        outcome,
        candidate_sha: "a".repeat(40),
      },
    });
    expect(result.status()).toBe(201);
  }
  const sessions = await apiClient.listTaskSessions(task.task_id);
  expect(sessions.sessions).toHaveLength(0);
  await testPage.reload();
  return { taskId: task.task_id, utcDay: occurrence.toISOString().slice(0, 10) };
}

export async function expectUsageSummary(page: Page) {
  const section = page.getByTestId("usage-stats");
  await expect(section).toHaveCount(1);
  await expect(section.getByTestId("usage-token-total")).toHaveText("120 known");
  await expect(section.getByTestId("usage-worker-runs")).toHaveText("2");
  await expect(section.getByTestId("usage-review-iterations")).toHaveText("2");
  await expect(section.getByTestId("usage-qa-runs")).toHaveText("1");
  await expect(section.getByText("Incomplete observations: 1", { exact: true })).toBeVisible();
  return section;
}
