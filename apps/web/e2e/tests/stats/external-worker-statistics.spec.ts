import { expect, test } from "../../fixtures/test-base";
import { seedExternalUsage, expectUsageSummary } from "./external-worker-statistics-helpers";

// Covers AC-COSTS-EXTERNAL-001.2, 002.1 through 002.4, and 003.1 through 003.4.
test("Stats shows external usage once after a lost acknowledgment and recovers a refresh", async ({
  testPage,
  backend,
  apiClient,
  seedData,
}) => {
  const { taskId, utcDay } = await seedExternalUsage({
    testPage,
    backend,
    apiClient,
    seedData,
    tasks: 51,
  });
  const section = await expectUsageSummary(testPage);
  const taskRow = section
    .getByTestId("usage-task-row")
    .filter({ has: testPage.locator(`a[href="/t/${taskId}"]`) });
  await expect(taskRow).toHaveCount(1);
  await expect(taskRow.getByTestId("usage-task-tokens")).toHaveText("120 known");
  await expect(section.getByText("fixture-provider", { exact: true })).toBeVisible();
  await section.getByText("Daily values", { exact: true }).click();
  const day = section
    .getByTestId("usage-daily-values")
    .getByRole("row")
    .filter({ hasText: utcDay });
  await expect(day).toContainText("120 known");
  const source = section.getByRole("combobox", { name: "Usage source", exact: true });
  await source.scrollIntoViewIfNeeded();
  await source.press("Enter");
  await expect(testPage.getByRole("option", { name: "All sources", exact: true })).toBeFocused();
  await testPage.keyboard.press("ArrowDown");
  await expect(testPage.getByRole("option", { name: "Native usage", exact: true })).toBeFocused();
  await testPage.keyboard.press("Enter");
  await expect(section.getByTestId("usage-token-total")).toHaveText("0");
  await expect(section.getByText("No recorded usage in this range", { exact: true })).toBeVisible();
  await source.click();
  await testPage.getByRole("option", { name: "External workers", exact: true }).click();
  await expectUsageSummary(testPage);
  await expect(section.getByTestId("usage-task-row")).toHaveCount(50);
  await section.getByRole("button", { name: "Next task page", exact: true }).click();
  await expect(section.getByTestId("usage-task-row")).toHaveCount(1);
  const route = `**/api/v1/workspaces/${seedData.workspaceId}/stats/usage?*`;
  await testPage.route(route, (request) =>
    request.fulfill({ status: 503, json: { error_code: "analytics_busy" } }),
  );
  await section.getByRole("button", { name: "Refresh usage", exact: true }).click();
  await expect(section.getByTestId("usage-error")).toBeVisible();
  await expect(section.getByTestId("usage-token-total")).toHaveText("120 known");
  await testPage.unroute(route);
  await section.getByRole("button", { name: "Retry", exact: true }).click();
  await expect(section.getByTestId("usage-error")).toHaveCount(0);
  expect((await apiClient.listTaskSessions(taskId)).sessions).toHaveLength(0);
});
