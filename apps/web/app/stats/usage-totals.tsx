import { useTranslation } from "react-i18next";
import type { UsageTotalsDTO } from "@/lib/types/http";
import { formatDuration } from "./stats-utils";

export function UsageTokens({
  totals,
  testId,
  showMissing = true,
}: {
  totals: UsageTotalsDTO;
  testId?: string;
  showMissing?: boolean;
}) {
  const { t, i18n } = useTranslation();
  const value = (totals.tokens_total ?? totals.known_tokens_total).toLocaleString(i18n.language);
  const missing = totals.incomplete_run_count + totals.incomplete_native_event_count;
  return (
    <div className="min-w-0 tabular-nums">
      <span data-testid={testId}>
        {totals.tokens_total === null ? t("stats:usageKnownValue", { value }) : value}
      </span>
      {showMissing && missing > 0 && (
        <p className="text-xs text-muted-foreground">
          {t("stats:usageMissing", { value: missing })}
        </p>
      )}
    </div>
  );
}

export function UsageSummary({ totals }: { totals: UsageTotalsDTO }) {
  const { t } = useTranslation();
  const metrics = [
    { label: t("stats:usageWorkerRuns"), value: totals.run_count, id: "usage-worker-runs" },
    {
      label: t("stats:usageReviews"),
      value: totals.review_iterations,
      id: "usage-review-iterations",
    },
    { label: t("stats:usageQA"), value: totals.qa_runs, id: "usage-qa-runs" },
    {
      label: t("stats:usageNativeRecords"),
      value: totals.native_event_count,
      id: "usage-native-records",
    },
  ];
  const missing = totals.incomplete_run_count + totals.incomplete_native_event_count;
  return (
    <dl className="grid min-w-0 grid-cols-2 gap-4 lg:grid-cols-3">
      <div className="col-span-2 rounded-sm border p-4 lg:col-span-1">
        <dt className="text-sm text-muted-foreground">{t("stats:usageTokens")}</dt>
        <dd className="break-words text-2xl font-semibold">
          <UsageTokens totals={totals} testId="usage-token-total" showMissing={false} />
        </dd>
        {missing > 0 && (
          <dd data-testid="usage-incomplete-count" className="text-sm text-muted-foreground">
            {t("stats:usageIncomplete", { value: missing })}
          </dd>
        )}
      </div>
      {metrics.map((metric) => (
        <div key={metric.id} className="min-w-0 rounded-sm border p-4">
          <dt className="text-sm text-muted-foreground">{metric.label}</dt>
          <dd data-testid={metric.id} className="text-2xl font-semibold tabular-nums">
            {metric.value}
          </dd>
        </div>
      ))}
      <div className="rounded-sm border p-4">
        <dt className="text-sm text-muted-foreground">{t("stats:usageDuration")}</dt>
        <dd className="text-2xl font-semibold tabular-nums">
          {formatDuration(totals.worker_duration_ms)}
        </dd>
      </div>
    </dl>
  );
}
