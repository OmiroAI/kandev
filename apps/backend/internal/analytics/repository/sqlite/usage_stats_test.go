package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/kandev/kandev/internal/analytics/models"
	dbutil "github.com/kandev/kandev/internal/db"
	taskmodels "github.com/kandev/kandev/internal/task/models"
	tasksqlite "github.com/kandev/kandev/internal/task/repository/sqlite"
	"github.com/kandev/kandev/internal/testutil"
)

func usageStatsFixture(t *testing.T) (*Repository, *tasksqlite.Repository) {
	t.Helper()
	filename := filepath.Join(t.TempDir(), "usage-stats.db")
	raw, err := dbutil.OpenSQLite(filename)
	if err != nil {
		t.Fatal(err)
	}
	writer := sqlx.NewDb(raw, "sqlite3")
	t.Cleanup(func() { _ = writer.Close() })
	tasks, err := tasksqlite.NewWithDB(writer, writer, nil)
	if err != nil {
		t.Fatal(err)
	}
	rawReader, err := dbutil.OpenSQLiteReader(filename)
	if err != nil {
		t.Fatal(err)
	}
	reader := sqlx.NewDb(rawReader, "sqlite3")
	t.Cleanup(func() { _ = reader.Close() })
	repo, err := NewWithDB(writer, reader)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"a-external", "b-native", "c-empty"} {
		if err := tasks.CreateTask(context.Background(), &taskmodels.Task{ID: id, WorkspaceID: "ws", Title: id}); err != nil {
			t.Fatal(err)
		}
	}
	return repo, tasks
}

func statsCount(value int64) *int64 { return &value }

func statsQuery() models.UsageStatsQuery {
	return models.UsageStatsQuery{DailyStart: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Until: time.Date(2026, 1, 4, 0, 0, 0, 0, time.UTC), Source: "all", Limit: 1}
}

func seedStatsObservations(t *testing.T, tasks *tasksqlite.Repository) {
	t.Helper()
	ctx := context.Background()
	date := time.Date(2026, 1, 2, 1, 30, 0, 0, time.UTC)
	activity := taskmodels.ExternalActivity{SchemaVersion: 1, Source: "kodi",
		SourceInstance: "d2d93284-5aab-4384-b50c-ac5bebc8e133", EventID: "usage:1",
		Kind: "worker_run", Phase: "coding", Outcome: "succeeded", Provider: "codex", Model: "codex-model",
		OccurredAt: date.In(time.FixedZone("offset", -2*60*60)), InputTokens: statsCount(10), OutputTokens: statsCount(2), CachedInputTokens: statsCount(6)}
	for _, eventID := range []string{"usage:1", "usage:2"} {
		activity.EventID = eventID
		if _, err := tasks.RecordExternalActivity(ctx, "a-external", "ws", "owner", &activity); err != nil {
			t.Fatal(err)
		}
		activity.InputTokens, activity.OutputTokens, activity.CachedInputTokens = nil, nil, nil
		activity.Outcome = "failed"
	}
	activity = taskmodels.ExternalActivity{SchemaVersion: 1, Source: "kodi", SourceInstance: activity.SourceInstance,
		Kind: "review_completed", Outcome: "rejected", CandidateSHA: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", OccurredAt: date}
	for _, eventID := range []string{"review:1", "review:2"} {
		activity.EventID = eventID
		if _, err := tasks.RecordExternalActivity(ctx, "a-external", "ws", "owner", &activity); err != nil {
			t.Fatal(err)
		}
	}
	activity.Kind, activity.Outcome, activity.EventID = "qa_completed", "failed", "qa:1"
	if _, err := tasks.RecordExternalActivity(ctx, "a-external", "ws", "owner", &activity); err != nil {
		t.Fatal(err)
	}
	native := &taskmodels.TaskUsageEvent{UsageEventID: "native:1", TaskID: "b-native", AgentType: "gemini", Provider: "google", Model: "native-model",
		TokensIn: 50, TokensCachedRead: statsCount(10), TokensOut: statsCount(4), TokensTotal: 64,
		CostSource: "unpriced", ContractVersion: 1, OccurredAt: date, CreatedAt: date}
	if err := tasks.CreateTaskUsageEvent(ctx, native); err != nil {
		t.Fatal(err)
	}
}

// Covers AC-COSTS-EXTERNAL-002.4 and AC-COSTS-EXTERNAL-003.1 through AC-COSTS-EXTERNAL-003.2.

func TestUsageStatsCombinesKnownCountsAndIndependentPages(t *testing.T) {
	repo, tasks := usageStatsFixture(t)
	seedStatsObservations(t, tasks)
	query := statsQuery()
	stats, err := repo.GetUsageStats(context.Background(), "ws", query)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Totals.KnownTokensTotal != 76 || stats.Totals.TokensTotal != nil || stats.Totals.IncompleteRunCount != 1 {
		t.Fatalf("totals=%+v", stats.Totals)
	}
	if stats.Totals.RunCount != 2 || stats.Totals.NativeEventCount != 1 || stats.Totals.ReviewIterations != 2 || stats.Totals.QARuns != 1 {
		t.Fatalf("activity=%+v", stats.Totals)
	}
	if len(stats.Daily) != 1 || stats.Daily[0].Date != "2026-01-02" || stats.Daily[0].KnownTokensTotal != 76 {
		t.Fatalf("daily=%+v", stats.Daily)
	}
	if len(stats.Tasks) != 1 || stats.Tasks[0].TaskID != "a-external" || stats.NextTaskCursor != "a-external" {
		t.Fatalf("task page=%+v cursor=%q", stats.Tasks, stats.NextTaskCursor)
	}
	if len(stats.Providers) != 1 || stats.Providers[0].Source != usageSourceExternal || stats.NextProviderCursor == "" {
		t.Fatalf("provider page=%+v", stats.Providers)
	}
	query.TaskCursor, query.ProviderCursor = stats.NextTaskCursor, stats.NextProviderCursor
	second, err := repo.GetUsageStats(context.Background(), "ws", query)
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Tasks) != 1 || second.Tasks[0].TaskID != "b-native" || second.Tasks[0].TokensTotal == nil || *second.Tasks[0].TokensTotal != 64 {
		t.Fatalf("native task=%+v", second.Tasks)
	}
	if len(second.Providers) != 1 || second.Providers[0].Source != "native" || second.NextProviderCursor != "" {
		t.Fatalf("native provider=%+v", second.Providers)
	}
	query.TaskCursor = second.NextTaskCursor
	third, err := repo.GetUsageStats(context.Background(), "ws", query)
	if err != nil {
		t.Fatal(err)
	}
	if len(third.Tasks) != 1 || third.Tasks[0].TaskID != "c-empty" || third.Tasks[0].TokensTotal == nil || *third.Tasks[0].TokensTotal != 0 {
		t.Fatalf("zero-activity task=%+v", third.Tasks)
	}
}

func TestUsageStatsSourceAndWorkspaceIsolation(t *testing.T) {
	repo, tasks := usageStatsFixture(t)
	seedStatsObservations(t, tasks)
	query := statsQuery()
	query.Source = "native"
	stats, err := repo.GetUsageStats(context.Background(), "ws", query)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Totals.TokensTotal == nil || *stats.Totals.TokensTotal != 64 || stats.Totals.RunCount != 0 {
		t.Fatalf("native totals=%+v", stats.Totals)
	}
	empty, err := repo.GetUsageStats(context.Background(), "another-workspace", query)
	if err != nil {
		t.Fatal(err)
	}
	if empty.Totals.KnownTokensTotal != 0 || len(empty.Tasks) != 0 || len(empty.Providers) != 0 {
		t.Fatalf("foreign rows: %+v", empty)
	}
}

func TestPostgresUsageStatsUsesUTCAndPreservesUnknownCounts(t *testing.T) {
	db := testutil.OpenIsolatedPostgres(t, testutil.PostgresDSNFromEnv(t))
	tasks, err := tasksqlite.NewWithDB(db, db, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := tasks.CreateWorkspace(context.Background(), &taskmodels.Workspace{ID: "ws", Name: "Usage"}); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"a-external", "b-native", "c-empty"} {
		if err := tasks.CreateTask(context.Background(), &taskmodels.Task{ID: id, WorkspaceID: "ws", Title: id}); err != nil {
			t.Fatal(err)
		}
	}
	repo, err := NewWithDB(db, db)
	if err != nil {
		t.Fatal(err)
	}
	seedStatsObservations(t, tasks)
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`SET TIME ZONE 'Pacific/Honolulu'`); err != nil {
		t.Fatal(err)
	}
	query := statsQuery()
	start := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	query.Start = &start
	stats, err := repo.GetUsageStats(context.Background(), "ws", query)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Totals.KnownTokensTotal != 76 || stats.Totals.TokensTotal != nil || stats.Totals.ReviewIterations != 2 {
		t.Fatalf("totals=%+v", stats.Totals)
	}
	if len(stats.Daily) != 1 || stats.Daily[0].Date != "2026-01-02" {
		t.Fatalf("daily=%+v", stats.Daily)
	}
}
