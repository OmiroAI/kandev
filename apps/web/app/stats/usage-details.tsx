import { useTranslation } from "react-i18next";
import { Button } from "@kandev/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@kandev/ui/card";
import TaskLink from "@/components/routing/task-link";
import type { UsageStatsDTO } from "@/lib/types/http";
import { UsageTokens } from "./usage-totals";

const TASK_METRICS = [
  ["run_count", "stats:usageWorkerRuns"],
  ["review_iterations", "stats:usageReviews"],
  ["qa_runs", "stats:usageQA"],
  ["native_event_count", "stats:usageNativeRecords"],
] as const;

type DetailProps = { data: UsageStatsDTO; disabled: boolean; next: () => Promise<void> };

export function UsageTasks({ data, disabled, next }: DetailProps) {
  const { t } = useTranslation();
  return (
    <Card className="min-w-0 rounded-sm">
      <CardHeader>
        <CardTitle className="text-sm">{t("stats:usageTasks")}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <table className="hidden w-full table-fixed text-left text-sm md:table">
          <thead>
            <tr className="border-b">
              <th className="w-1/4 p-2">{t("stats:usageTask")}</th>
              <th className="p-2">{t("stats:usageTokens")}</th>
              {TASK_METRICS.map(([key, label]) => (
                <th key={key} className="break-words p-2">
                  {t(label)}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {data.tasks.map((task) => (
              <tr key={task.task_id} data-testid="usage-task-row" className="border-b align-top">
                <td className="break-words p-2">
                  <TaskLink
                    taskId={task.task_id}
                    className="cursor-pointer underline underline-offset-2"
                  >
                    {task.title}
                  </TaskLink>
                </td>
                <td className="break-words p-2">
                  <UsageTokens totals={task} testId="usage-task-tokens" />
                </td>
                {TASK_METRICS.map(([key]) => (
                  <td key={key} className="p-2 tabular-nums">
                    {task[key]}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
        <div className="space-y-3 md:hidden">
          {data.tasks.map((task) => (
            <article
              key={task.task_id}
              data-testid="usage-task-card"
              className="min-w-0 rounded-sm border p-3"
            >
              <TaskLink
                taskId={task.task_id}
                className="block min-h-11 cursor-pointer break-words font-medium underline underline-offset-2"
              >
                {task.title}
              </TaskLink>
              <dl className="grid grid-cols-2 gap-3 text-sm">
                <div className="col-span-2">
                  <dt className="text-muted-foreground">{t("stats:usageTokens")}</dt>
                  <dd>
                    <UsageTokens totals={task} />
                  </dd>
                </div>
                {TASK_METRICS.map(([key, label]) => (
                  <div key={key}>
                    <dt className="text-muted-foreground">{t(label)}</dt>
                    <dd className="tabular-nums">{task[key]}</dd>
                  </div>
                ))}
              </dl>
            </article>
          ))}
        </div>
        <Button
          className="h-auto min-h-11 max-w-full cursor-pointer whitespace-normal py-2"
          variant="outline"
          disabled={disabled || !data.next_task_cursor}
          onClick={() => void next()}
        >
          {t("stats:usageNextTasks")}
        </Button>
      </CardContent>
    </Card>
  );
}

export function UsageProviders({ data, disabled, next }: DetailProps) {
  const { t } = useTranslation();
  return (
    <Card className="min-w-0 rounded-sm">
      <CardHeader>
        <CardTitle className="text-sm">{t("stats:usageProviders")}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <dl className="space-y-3">
          {data.providers.map((provider) => (
            <div
              data-testid="usage-provider-row"
              key={JSON.stringify([provider.source, provider.provider, provider.model])}
              className="flex min-w-0 flex-wrap justify-between gap-3 border-b pb-3"
            >
              <dt className="min-w-0 break-words">
                <span className="block text-xs text-muted-foreground">
                  {t(
                    provider.source === "native"
                      ? "stats:usageSourceNative"
                      : "stats:usageSourceExternal",
                  )}
                </span>
                <span className="block break-all font-medium">
                  {provider.provider || t("stats:usageNotReported")}
                </span>
                <span className="block break-all text-sm text-muted-foreground">
                  {provider.model || t("stats:usageNotReported")}
                </span>
              </dt>
              <dd>
                <UsageTokens totals={provider} />
              </dd>
            </div>
          ))}
        </dl>
        <Button
          className="h-auto min-h-11 max-w-full cursor-pointer whitespace-normal py-2"
          variant="outline"
          disabled={disabled || !data.next_provider_cursor}
          onClick={() => void next()}
        >
          {t("stats:usageNextProviders")}
        </Button>
      </CardContent>
    </Card>
  );
}
