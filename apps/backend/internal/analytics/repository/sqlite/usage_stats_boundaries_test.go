package sqlite

import (
	"context"
	"errors"
	taskmodels "github.com/kandev/kandev/internal/task/models"
	"strings"
	"testing"
	"time"

	"github.com/kandev/kandev/internal/analytics"
	"github.com/kandev/kandev/internal/analytics/models"
)

// Covers AC-COSTS-EXTERNAL-003.4.

func TestUsageStatsRejectsInvalidBoundsAndCursors(t *testing.T) {
	repo, _ := usageStatsFixture(t)
	for name, change := range map[string]func(*models.UsageStatsQuery){
		"source":    func(q *models.UsageStatsQuery) { q.Source = "other" },
		"limit":     func(q *models.UsageStatsQuery) { q.Limit = 101 },
		"reversed":  func(q *models.UsageStatsQuery) { q.DailyStart = q.Until },
		"unbounded": func(q *models.UsageStatsQuery) { q.DailyStart = q.Until.AddDate(-2, 0, 0) },
		"cursor":    func(q *models.UsageStatsQuery) { q.ProviderCursor = "invalid!" },
	} {
		t.Run(name, func(t *testing.T) {
			query := statsQuery()
			change(&query)
			if _, err := repo.GetUsageStats(context.Background(), "ws", query); !errors.Is(err, analytics.ErrInvalidUsageQuery) {
				t.Fatalf("error=%v", err)
			}
		})
	}
}

func TestUsageStatsAllTimeBackfillAndCurrentWorkspaceOwnership(t *testing.T) {
	repo, tasks := usageStatsFixture(t)
	seedStatsObservations(t, tasks)
	query := statsQuery()
	query.DailyStart = time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)
	stats, err := repo.GetUsageStats(context.Background(), "ws", query)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Totals.KnownTokensTotal != 76 || len(stats.Daily) != 0 {
		t.Fatalf("all-time backfill=%+v", stats)
	}
	query.Start = &query.DailyStart
	ranged, err := repo.GetUsageStats(context.Background(), "ws", query)
	if err != nil {
		t.Fatal(err)
	}
	if ranged.Totals.KnownTokensTotal != 0 || ranged.Totals.RunCount != 0 {
		t.Fatalf("range=%+v", ranged.Totals)
	}
	if _, err := repo.db.Exec(`UPDATE tasks SET workspace_id='destination' WHERE id='a-external'`); err != nil {
		t.Fatal(err)
	}
	query.Start = nil
	moved, err := repo.GetUsageStats(context.Background(), "destination", query)
	if err != nil {
		t.Fatal(err)
	}
	old, err := repo.GetUsageStats(context.Background(), "ws", query)
	if err != nil {
		t.Fatal(err)
	}
	if moved.Totals.KnownTokensTotal != 12 || moved.Totals.ReviewIterations != 2 || old.Totals.KnownTokensTotal != 64 || old.Totals.RunCount != 0 {
		t.Fatalf("destination=%+v old=%+v", moved.Totals, old.Totals)
	}
}

func TestUsageStatsPagesEscapedProviderMetadata(t *testing.T) {
	repo, tasks := usageStatsFixture(t)
	activity := taskmodels.ExternalActivity{SchemaVersion: 1, Source: "kodi", SourceInstance: "d2d93284-5aab-4384-b50c-ac5bebc8e133",
		Kind: "worker_run", Phase: "unknown", Outcome: "unknown", OccurredAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		Provider: strings.Repeat("<", 256), Model: strings.Repeat("&", 256), EventID: "cursor:1"}
	if _, err := tasks.RecordExternalActivity(context.Background(), "a-external", "ws", "owner", &activity); err != nil {
		t.Fatal(err)
	}
	activity.Provider, activity.Model, activity.EventID = "z-provider", "model", "cursor:2"
	if _, err := tasks.RecordExternalActivity(context.Background(), "a-external", "ws", "owner", &activity); err != nil {
		t.Fatal(err)
	}
	query := statsQuery()
	first, err := repo.GetUsageStats(context.Background(), "ws", query)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Providers) != 1 || first.NextProviderCursor == "" {
		t.Fatalf("first page=%+v", first)
	}
	query.ProviderCursor = first.NextProviderCursor
	second, err := repo.GetUsageStats(context.Background(), "ws", query)
	if err != nil {
		t.Fatalf("server-generated cursor rejected: %v", err)
	}
	if len(second.Providers) != 1 || second.Providers[0].Provider != "z-provider" || second.NextProviderCursor != "" {
		t.Fatalf("second page=%+v", second)
	}
}
