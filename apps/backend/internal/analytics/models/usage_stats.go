package models

import "time"

// UsageTotals keeps known contributions visible without claiming complete totals.
type UsageTotals struct {
	KnownTokensTotal           int64  `json:"known_tokens_total" db:"known_tokens_total"`
	TokensTotal                *int64 `json:"tokens_total" db:"tokens_total"`
	RunCount                   int64  `json:"run_count" db:"run_count"`
	NativeEventCount           int64  `json:"native_event_count" db:"native_event_count"`
	IncompleteRunCount         int64  `json:"incomplete_run_count" db:"incomplete_run_count"`
	IncompleteNativeEventCount int64  `json:"incomplete_native_event_count" db:"incomplete_native_event_count"`
	ReviewIterations           int64  `json:"review_iterations" db:"review_iterations"`
	QARuns                     int64  `json:"qa_runs" db:"qa_runs"`
	WorkerDurationMs           int64  `json:"worker_duration_ms" db:"worker_duration_ms"`
}

type UsageDay struct {
	Date string `json:"date" db:"date"`
	UsageTotals
}

type UsageTask struct {
	TaskID string `json:"task_id" db:"task_id"`
	Title  string `json:"title" db:"title"`
	State  string `json:"state" db:"state"`
	UsageTotals
}

type UsageProvider struct {
	Source   string `json:"source" db:"source"`
	Provider string `json:"provider" db:"provider"`
	Model    string `json:"model" db:"model"`
	UsageTotals
}

type UsageStats struct {
	Totals             UsageTotals     `json:"totals"`
	Daily              []UsageDay      `json:"daily"`
	Tasks              []UsageTask     `json:"tasks"`
	Providers          []UsageProvider `json:"providers"`
	NextTaskCursor     string          `json:"next_task_cursor"`
	NextProviderCursor string          `json:"next_provider_cursor"`
	DailyFrom          string          `json:"daily_from"`
	DailyUntil         string          `json:"daily_until"`
}

// UsageStatsQuery selects a bounded daily series and independent detail pages.
// A nil Start keeps all historical contributions in the summary and detail pages.
type UsageStatsQuery struct {
	Start          *time.Time
	DailyStart     time.Time
	Until          time.Time
	Source         string
	Limit          int
	TaskCursor     string
	ProviderCursor string
}
