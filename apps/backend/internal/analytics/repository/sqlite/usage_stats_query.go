package sqlite

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	analyticserrors "github.com/kandev/kandev/internal/analytics"
	"github.com/kandev/kandev/internal/analytics/models"
	"github.com/kandev/kandev/internal/db/dialect"
)

// JSON escaping can expand each bounded metadata byte sixfold before base64.
const maxUsageProviderCursorBytes = 8192

var errInvalidUsageQuery = analyticserrors.ErrInvalidUsageQuery

type usageProviderCursor struct {
	Source   string `json:"source"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

func validateUsageQuery(query models.UsageStatsQuery) error {
	if query.Source != "all" && query.Source != "native" && query.Source != "external" {
		return errInvalidUsageQuery
	}
	span := query.Until.Sub(query.DailyStart)
	if query.Limit < 1 || query.Limit > 100 || query.DailyStart.IsZero() || span <= 0 || span > 365*24*time.Hour || len(query.TaskCursor) > 128 {
		return errInvalidUsageQuery
	}
	_, err := decodeUsageProviderCursor(query.ProviderCursor)
	return err
}

func decodeUsageProviderCursor(value string) (usageProviderCursor, error) {
	var cursor usageProviderCursor
	if value == "" {
		return cursor, nil
	}
	if len(value) > maxUsageProviderCursorBytes {
		return cursor, errInvalidUsageQuery
	}
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || json.Unmarshal(data, &cursor) != nil {
		return cursor, errInvalidUsageQuery
	}
	if (cursor.Source != "native" && cursor.Source != "external") || len(cursor.Provider) > 256 || len(cursor.Model) > 256 {
		return cursor, errInvalidUsageQuery
	}
	return cursor, nil
}

func encodeUsageProviderCursor(provider models.UsageProvider) string {
	body, _ := json.Marshal(usageProviderCursor{Source: provider.Source, Provider: provider.Provider, Model: provider.Model})
	return base64.RawURLEncoding.EncodeToString(body)
}

func usageStatsCTE(driver string) string {
	occurred := dialect.NaiveUTCTimestampOf(driver, "occurred_at")
	parameter := dialect.DateTimeOf(driver, "?")
	return `WITH eligible_tasks AS (
		SELECT t.id,t.title,t.state FROM tasks t WHERE t.workspace_id=? AND t.is_ephemeral=0` + andNotAutomationOriginT + `
	), observations AS (
		SELECT e.task_id,'external' AS source,e.provider,e.model,e.occurred_at,
			COALESCE(e.input_tokens,0)+COALESCE(e.output_tokens,0) AS known_tokens_total,
			CASE WHEN e.kind='worker_run' THEN 1 ELSE 0 END AS run_count,0 AS native_event_count,
			CASE WHEN e.kind='worker_run' AND (e.input_tokens IS NULL OR e.output_tokens IS NULL) THEN 1 ELSE 0 END AS incomplete_run_count,
			0 AS incomplete_native_event_count,
			CASE WHEN e.kind='review_completed' THEN 1 ELSE 0 END AS review_iterations,
			CASE WHEN e.kind='qa_completed' THEN 1 ELSE 0 END AS qa_runs,
			COALESCE(e.duration_ms,0) AS worker_duration_ms
		FROM task_external_activity e JOIN eligible_tasks t ON t.id=e.task_id
		UNION ALL
		SELECT e.task_id,'native',e.provider,e.model,e.occurred_at,e.tokens_total,
			0,1,0,CASE WHEN e.tokens_out IS NULL OR e.usage_completeness IN ('unknown','partial','pending') THEN 1 ELSE 0 END,
			0,0,0
		FROM task_usage_events e JOIN eligible_tasks t ON t.id=e.task_id
	), selected_events AS (
		SELECT * FROM observations WHERE (?='all' OR source=?) AND (` +
		dialect.NullableTimestamp(driver, "?") + ` IS NULL OR ` + occurred + `>=` + parameter + `) AND ` + occurred + `<` + parameter + `
	) `
}

func usageStatsArguments(workspace string, query models.UsageStatsQuery) []any {
	start := rangeStartArg(query.Start)
	return []any{workspace, query.Source, query.Source, start, start, query.Until.UTC()}
}

func usageAggregate(alias string) string {
	columns := []string{"known_tokens_total", "run_count", "native_event_count", "incomplete_run_count",
		"incomplete_native_event_count", "review_iterations", "qa_runs", "worker_duration_ms"}
	result := ""
	for i, column := range columns {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("COALESCE(SUM(%s.%s),0) AS %s", alias, column, column)
	}
	return result
}

func completeUsageTotal(totals *models.UsageTotals) {
	if totals.IncompleteRunCount+totals.IncompleteNativeEventCount == 0 {
		value := totals.KnownTokensTotal
		totals.TokensTotal = &value
	}
}
