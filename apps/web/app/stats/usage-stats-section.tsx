import { useId, useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@kandev/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@kandev/ui/select";
import { useUsageStats } from "@/hooks/domains/stats/use-usage-stats";
import type { StatsRange } from "@/lib/api/domains/stats-api";
import type { UsageSource } from "@/lib/types/http";
import { UsageSummary } from "./usage-totals";
import { UsageTasks, UsageProviders } from "./usage-details";
import { UsageDailyChart } from "./usage-daily-chart";

function UsageReadError({
  hasPrevious,
  refresh,
}: {
  hasPrevious: boolean;
  refresh: () => Promise<void>;
}) {
  const { t } = useTranslation();
  return (
    <div data-testid="usage-error" className="space-y-2 rounded-sm border p-3">
      <p role="status">{t("stats:usageUnavailable")}</p>
      {hasPrevious && <p className="text-sm text-muted-foreground">{t("stats:usagePrevious")}</p>}
      <Button
        className="h-auto min-h-11 max-w-full cursor-pointer whitespace-normal py-2"
        variant="outline"
        onClick={() => void refresh()}
      >
        {t("stats:retry")}
      </Button>
    </div>
  );
}

export function UsageStatsSection({
  workspaceId,
  range,
}: {
  workspaceId: string;
  range: StatsRange;
}) {
  const { t } = useTranslation();
  const titleId = useId();
  const sourceId = useId();
  const [source, setSource] = useState<UsageSource>("all");
  const { data, loading, error, refresh, nextTasks, nextProviders } = useUsageStats(
    workspaceId,
    range,
    source,
  );
  const empty =
    data &&
    data.totals.run_count +
      data.totals.native_event_count +
      data.totals.review_iterations +
      data.totals.qa_runs ===
      0;
  return (
    <section
      data-testid="usage-stats"
      aria-labelledby={titleId}
      aria-busy={loading}
      className="min-w-0 space-y-4"
    >
      <div className="flex flex-wrap items-end justify-between gap-3">
        <h2 id={titleId} className="text-lg font-semibold">
          {t("stats:usageTitle")}
        </h2>
        <div className="flex min-w-0 flex-wrap items-end gap-3">
          <div className="space-y-1">
            <label htmlFor={sourceId} className="block text-sm text-muted-foreground">
              {t("stats:usageSource")}
            </label>
            <Select
              value={source}
              onValueChange={(value) => {
                if (value === "all" || value === "native" || value === "external") setSource(value);
              }}
            >
              <SelectTrigger id={sourceId} className="min-h-11 w-48 cursor-pointer">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem
                  className="h-auto min-h-11 max-w-full cursor-pointer whitespace-normal py-2"
                  value="all"
                >
                  {t("stats:usageSourceAll")}
                </SelectItem>
                <SelectItem
                  className="h-auto min-h-11 max-w-full cursor-pointer whitespace-normal py-2"
                  value="native"
                >
                  {t("stats:usageSourceNative")}
                </SelectItem>
                <SelectItem
                  className="h-auto min-h-11 max-w-full cursor-pointer whitespace-normal py-2"
                  value="external"
                >
                  {t("stats:usageSourceExternal")}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
          <Button
            className="h-auto min-h-11 max-w-full cursor-pointer whitespace-normal py-2"
            variant="outline"
            disabled={loading}
            onClick={() => void refresh()}
          >
            {t("stats:usageRefresh")}
          </Button>
        </div>
      </div>
      {loading && (
        <p role="status" className="text-sm text-muted-foreground">
          {t("stats:usageLoading")}
        </p>
      )}
      {error && <UsageReadError hasPrevious={Boolean(data)} refresh={refresh} />}
      {data && (
        <>
          <UsageSummary totals={data.totals} />
          {empty && <p className="text-sm text-muted-foreground">{t("stats:usageEmpty")}</p>}
          <UsageDailyChart data={data} />
          <UsageProviders data={data} disabled={loading || error} next={nextProviders} />
          <UsageTasks data={data} disabled={loading || error} next={nextTasks} />
        </>
      )}
    </section>
  );
}
