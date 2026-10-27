import { expect, test } from "../../fixtures/test-base";
import { seedExternalUsage, expectUsageSummary } from "./external-worker-statistics-helpers";

// Covers AC-COSTS-EXTERNAL-003.3 phone parity and touch controls.
test("phone Stats keeps usage cards and source controls reachable without overflow", async ({
  testPage,
  backend,
  apiClient,
  seedData,
}) => {
  await testPage.setViewportSize({ width: 320, height: 740 });
  const { taskId } = await seedExternalUsage({ testPage, backend, apiClient, seedData });
  const section = await expectUsageSummary(testPage);
  const card = section
    .getByTestId("usage-task-card")
    .filter({ has: testPage.locator(`a[href="/t/${taskId}"]`) });
  await expect(card).toHaveCount(1);
  await expect(card).toContainText("120 known");
  await expect(card).toContainText("Review iterations");
  await expect(card).toContainText("QA runs");
  const source = section.getByRole("combobox", { name: "Usage source", exact: true });
  await source.scrollIntoViewIfNeeded();
  const box = await source.boundingBox();
  expect(box).not.toBeNull();
  expect(box!.height).toBeGreaterThanOrEqual(44);
  expect(
    await source.evaluate((element) => {
      const rect = element.getBoundingClientRect();
      return element.contains(
        document.elementFromPoint(rect.x + rect.width / 2, rect.y + rect.height / 2),
      );
    }),
  ).toBe(true);
  await source.tap();
  await testPage.getByRole("option", { name: "External workers", exact: true }).tap();
  await expectUsageSummary(testPage);
  expect(await testPage.evaluate(() => document.documentElement.scrollWidth)).toBe(
    await testPage.evaluate(() => document.documentElement.clientWidth),
  );
});
