import { useTranslation } from "react-i18next";
import { Card, CardContent, CardHeader, CardTitle } from "@kandev/ui/card";
import type { UsageStatsDTO } from "@/lib/types/http";
import { usageDailySeries } from "./usage-daily-series";
import { UsageTokens } from "./usage-totals";

export function UsageDailyChart({ data }: { data: UsageStatsDTO }) {
  const { t } = useTranslation();
  const series = usageDailySeries(data.daily, data.daily_from, data.daily_until);
  const maximum = Math.max(1, ...series.map((day) => day.knownTokens));
  return (
    <Card className="min-w-0 rounded-sm">
      <CardHeader>
        <CardTitle className="text-sm">{t("stats:usageDaily")}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        <div
          role="img"
          aria-label={t("stats:usageDailyChart")}
          className="flex h-32 min-w-0 items-end gap-px border-b"
        >
          {series.map((day) => (
            <div
              key={day.date}
              aria-hidden
              className={`min-w-0 flex-1 rounded-t-sm ${day.incomplete ? "bg-amber-500/70" : "bg-emerald-500/70"}`}
              style={{ height: `${(100 * day.knownTokens) / maximum}%` }}
            />
          ))}
        </div>
        {series.length > 0 && (
          <p className="text-xs text-muted-foreground">
            {t("stats:usageHorizon", {
              from: series[0].date,
              until: series[series.length - 1].date,
            })}
          </p>
        )}
        <details>
          <summary className="flex min-h-11 cursor-pointer items-center text-sm underline underline-offset-2">
            {t("stats:usageDailyValues")}
          </summary>
          <table data-testid="usage-daily-values" className="w-full table-fixed text-left text-sm">
            <thead>
              <tr>
                <th className="w-2/5 py-2">{t("stats:usageDate")}</th>
                <th className="py-2">{t("stats:usageTokens")}</th>
              </tr>
            </thead>
            <tbody>
              {series.map((day) => {
                const observation = data.daily.find((item) => item.date === day.date);
                return (
                  <tr key={day.date} className="border-t">
                    <td className="py-2 align-top">{day.date}</td>
                    <td className="break-words py-2">
                      {observation ? <UsageTokens totals={observation} /> : day.knownTokens}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </details>
      </CardContent>
    </Card>
  );
}
