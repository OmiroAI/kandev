package sqlite

import (
	"context"

	"github.com/kandev/kandev/internal/analytics/models"
	"github.com/kandev/kandev/internal/db/dialect"
)

// GetUsageStats reads native and external facts without changing either source.
func (r *Repository) GetUsageStats(parent context.Context, workspace string, query models.UsageStatsQuery) (stats *models.UsageStats, err error) {
	if err := validateUsageQuery(query); err != nil {
		return nil, err
	}
	ctx, release, err := r.beginAnalyticsOperation(parent)
	if err != nil {
		return nil, err
	}
	defer func() { err = normalizeAnalyticsOperationError(parent, ctx, err); release() }()
	stats = &models.UsageStats{Daily: []models.UsageDay{}, Tasks: []models.UsageTask{}, Providers: []models.UsageProvider{},
		DailyFrom: query.DailyStart.UTC().Format("2006-01-02"), DailyUntil: query.Until.UTC().Format("2006-01-02")}
	prefix := usageStatsCTE(r.ro.DriverName())
	arguments := usageStatsArguments(workspace, query)
	if err := r.ro.GetContext(ctx, &stats.Totals, r.ro.Rebind(prefix+`SELECT `+usageAggregate("e")+` FROM selected_events e`), arguments...); err != nil {
		return nil, err
	}
	completeUsageTotal(&stats.Totals)
	if err := r.readUsageDays(ctx, stats, prefix, arguments, query); err != nil {
		return nil, err
	}
	if err := r.readUsageTasks(ctx, stats, prefix, arguments, query); err != nil {
		return nil, err
	}
	if err := r.readUsageProviders(ctx, stats, prefix, arguments, query); err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *Repository) readUsageDays(ctx context.Context, stats *models.UsageStats, prefix string, arguments []any, query models.UsageStatsQuery) error {
	driver := r.ro.DriverName()
	day := dialect.DateText(driver, "e.occurred_at")
	statement := prefix + `SELECT ` + day + ` AS date,` + usageAggregate("e") + ` FROM selected_events e WHERE ` +
		dialect.NaiveUTCTimestampOf(driver, "e.occurred_at") + `>=` + dialect.DateTimeOf(driver, "?") + ` GROUP BY ` + day + ` ORDER BY date`
	args := append(append([]any{}, arguments...), query.DailyStart.UTC())
	if err := r.ro.SelectContext(ctx, &stats.Daily, r.ro.Rebind(statement), args...); err != nil {
		return err
	}
	for i := range stats.Daily {
		completeUsageTotal(&stats.Daily[i].UsageTotals)
	}
	return nil
}

func (r *Repository) readUsageTasks(ctx context.Context, stats *models.UsageStats, prefix string, arguments []any, query models.UsageStatsQuery) error {
	statement := prefix + `SELECT t.id AS task_id,t.title,t.state,` + usageAggregate("e") + `
		FROM eligible_tasks t LEFT JOIN selected_events e ON e.task_id=t.id
		WHERE t.id>? GROUP BY t.id,t.title,t.state ORDER BY t.id LIMIT ?`
	args := append(append([]any{}, arguments...), query.TaskCursor, query.Limit+1)
	if err := r.ro.SelectContext(ctx, &stats.Tasks, r.ro.Rebind(statement), args...); err != nil {
		return err
	}
	if len(stats.Tasks) > query.Limit {
		stats.Tasks = stats.Tasks[:query.Limit]
		stats.NextTaskCursor = stats.Tasks[len(stats.Tasks)-1].TaskID
	}
	for i := range stats.Tasks {
		completeUsageTotal(&stats.Tasks[i].UsageTotals)
	}
	return nil
}

func (r *Repository) readUsageProviders(ctx context.Context, stats *models.UsageStats, prefix string, arguments []any, query models.UsageStatsQuery) error {
	cursor, err := decodeUsageProviderCursor(query.ProviderCursor)
	if err != nil {
		return err
	}
	statement := prefix + `SELECT e.source,e.provider,e.model,` + usageAggregate("e") + ` FROM selected_events e
		WHERE (e.run_count=1 OR e.native_event_count=1) AND (e.source,e.provider,e.model)>(?,?,?)
		GROUP BY e.source,e.provider,e.model ORDER BY e.source,e.provider,e.model LIMIT ?`
	args := append(append([]any{}, arguments...), cursor.Source, cursor.Provider, cursor.Model, query.Limit+1)
	if err := r.ro.SelectContext(ctx, &stats.Providers, r.ro.Rebind(statement), args...); err != nil {
		return err
	}
	if len(stats.Providers) > query.Limit {
		stats.Providers = stats.Providers[:query.Limit]
		stats.NextProviderCursor = encodeUsageProviderCursor(stats.Providers[len(stats.Providers)-1])
	}
	for i := range stats.Providers {
		completeUsageTotal(&stats.Providers[i].UsageTotals)
	}
	return nil
}
