package sqlite

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/kandev/kandev/internal/task/models"
	"github.com/kandev/kandev/internal/task/repository/repoerrors"
	"github.com/kandev/kandev/internal/testutil"
)

func externalRunFixture() *models.ExternalActivity {
	input := int64(0)
	return &models.ExternalActivity{SchemaVersion: 1, Source: "kodi",
		SourceInstance: "d2d93284-5aab-4384-b50c-ac5bebc8e133", EventID: "usage:42",
		Kind: "worker_run", Phase: "coding", Outcome: "failed", InputTokens: &input,
		OccurredAt: time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC)}
}

// Covers AC-COSTS-EXTERNAL-001.2 and AC-COSTS-EXTERNAL-001.4.

func TestExternalActivityStorageReplayAndConflicts(t *testing.T) {
	verifyExternalActivityStorage(t, newUsageEventsTestRepo(t))
}

func TestPostgresExternalActivityStorageReplayAndConflicts(t *testing.T) {
	db := testutil.OpenIsolatedPostgres(t, testutil.PostgresDSNFromEnv(t))
	repo, err := NewWithDB(db, db, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateWorkspace(context.Background(), &models.Workspace{ID: "ws-1", Name: "External activity"}); err != nil {
		t.Fatal(err)
	}
	verifyExternalActivityStorage(t, repo)
}

func verifyExternalActivityStorage(t *testing.T, repo *Repository) {
	t.Helper()
	createUsageEventsTestTask(t, repo, "external-task")
	createUsageEventsTestTask(t, repo, "other-task")
	ctx := context.Background()
	activity := externalRunFixture()
	first, err := repo.RecordExternalActivity(ctx, "external-task", "ws-1", "user", activity)
	if err != nil || !first.Created || first.ID == "" {
		t.Fatalf("first insert = %+v, %v", first, err)
	}
	activity.OccurredAt = activity.OccurredAt.In(time.FixedZone("offset", 2*60*60))
	replay, err := repo.RecordExternalActivity(ctx, "external-task", "ws-1", "user", activity)
	if err != nil || replay.Created || replay.ID != first.ID {
		t.Fatalf("replay = %+v, %v", replay, err)
	}
	_, err = repo.RecordExternalActivity(ctx, "other-task", "ws-1", "user", activity)
	if !errors.Is(err, repoerrors.ErrExternalActivityConflict) {
		t.Fatalf("changed task: %v", err)
	}
	activity.Model = "changed"
	_, err = repo.RecordExternalActivity(ctx, "external-task", "ws-1", "user", activity)
	if !errors.Is(err, repoerrors.ErrExternalActivityConflict) {
		t.Fatalf("changed measurement: %v", err)
	}
	var count, input, unknown, native int
	if err := repo.db.QueryRow(`SELECT COUNT(*), SUM(input_tokens), SUM(CASE WHEN output_tokens IS NULL THEN 1 ELSE 0 END) FROM task_external_activity`).Scan(&count, &input, &unknown); err != nil {
		t.Fatal(err)
	}
	if count != 1 || input != 0 || unknown != 1 {
		t.Fatalf("recorded counts = %d/%d/%d", count, input, unknown)
	}
	if err := repo.db.Get(&native, `SELECT COUNT(*) FROM task_usage_events`); err != nil {
		t.Fatal(err)
	}
	if native != 0 {
		t.Fatal("external work manufactured a native usage event")
	}
	if err := repo.initSchema(); err != nil {
		t.Fatalf("schema replay: %v", err)
	}
}

func TestExternalActivityConcurrentReplayCountsOnce(t *testing.T) {
	verifyExternalActivityConcurrentReplay(t, newUsageEventsTestRepo(t))
}

func TestPostgresExternalActivityConcurrentReplayCountsOnce(t *testing.T) {
	db := testutil.OpenIsolatedPostgres(t, testutil.PostgresDSNFromEnv(t))
	repo, err := NewWithDB(db, db, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateWorkspace(context.Background(), &models.Workspace{ID: "ws-1", Name: "External activity"}); err != nil {
		t.Fatal(err)
	}
	verifyExternalActivityConcurrentReplay(t, repo)
}

func verifyExternalActivityConcurrentReplay(t *testing.T, repo *Repository) {
	t.Helper()
	createUsageEventsTestTask(t, repo, "external-race")
	activity := externalRunFixture()
	results := make(chan models.ExternalActivityReceipt, 8)
	errorsFound := make(chan error, 8)
	var wait sync.WaitGroup
	for range 8 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			receipt, err := repo.RecordExternalActivity(context.Background(), "external-race", "ws-1", "user", activity)
			results <- receipt
			errorsFound <- err
		}()
	}
	wait.Wait()
	close(results)
	close(errorsFound)
	for err := range errorsFound {
		if err != nil {
			t.Fatal(err)
		}
	}
	created := 0
	identities := map[string]bool{}
	for result := range results {
		if result.Created {
			created++
		}
		identities[result.ID] = true
	}
	if created != 1 || len(identities) != 1 || identities[""] {
		t.Fatalf("created=%d identities=%v", created, identities)
	}
}

func TestExternalActivityRetainsIdentityAfterTaskTransfer(t *testing.T) {
	repo := newUsageEventsTestRepo(t)
	createUsageEventsTestTask(t, repo, "external-transfer")
	activity := externalRunFixture()
	first, err := repo.RecordExternalActivity(context.Background(), "external-transfer", "ws-1", "user", activity)
	if err != nil || !first.Created {
		t.Fatalf("insert: %+v %v", first, err)
	}
	if _, err := repo.db.Exec(`UPDATE tasks SET workspace_id='ws-2' WHERE id='external-transfer'`); err != nil {
		t.Fatal(err)
	}
	second, err := repo.RecordExternalActivity(context.Background(), "external-transfer", "ws-2", "user", activity)
	if err != nil || second.Created || second.ID != first.ID {
		t.Fatalf("transferred replay: %+v %v", second, err)
	}
	if _, err := repo.RecordExternalActivity(context.Background(), "external-transfer", "ws-1", "user", activity); err == nil {
		t.Fatal("stale workspace admitted")
	}
}

func TestExternalActivityReplayAfterTransferRetainsOriginalNamespace(t *testing.T) {
	verifyExternalActivityTransferNamespace(t, newUsageEventsTestRepo(t))
}

func TestPostgresExternalActivityReplayAfterTransferRetainsOriginalNamespace(t *testing.T) {
	db := testutil.OpenIsolatedPostgres(t, testutil.PostgresDSNFromEnv(t))
	repo, err := NewWithDB(db, db, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateWorkspace(context.Background(), &models.Workspace{ID: "ws-1", Name: "Original"}); err != nil {
		t.Fatal(err)
	}
	verifyExternalActivityTransferNamespace(t, repo)
}

func verifyExternalActivityTransferNamespace(t *testing.T, repo *Repository) {
	t.Helper()
	ctx := context.Background()
	if err := repo.CreateWorkspace(ctx, &models.Workspace{ID: "ws-transfer", Name: "Destination"}); err != nil {
		t.Fatal(err)
	}
	createUsageEventsTestTask(t, repo, "external-namespace-original")
	if err := repo.CreateTask(ctx, &models.Task{ID: "external-namespace-other", WorkspaceID: "ws-transfer", Title: "Other"}); err != nil {
		t.Fatal(err)
	}
	activity := externalRunFixture()
	original, err := repo.RecordExternalActivity(ctx, "external-namespace-original", "ws-1", "owner", activity)
	if err != nil {
		t.Fatal(err)
	}
	other, err := repo.RecordExternalActivity(ctx, "external-namespace-other", "ws-transfer", "owner", activity)
	if err != nil || other.ID == original.ID {
		t.Fatalf("separate namespace receipt=%+v error=%v", other, err)
	}
	if _, err := repo.db.Exec(`UPDATE tasks SET workspace_id='ws-transfer' WHERE id='external-namespace-original'`); err != nil {
		t.Fatal(err)
	}
	replay, err := repo.RecordExternalActivity(ctx, "external-namespace-original", "ws-transfer", "owner", activity)
	if err != nil || replay.Created || replay.ID != original.ID {
		t.Fatalf("original namespace replay=%+v error=%v", replay, err)
	}
	activity.Model = "changed"
	if _, err := repo.RecordExternalActivity(ctx, "external-namespace-original", "ws-transfer", "owner", activity); !errors.Is(err, repoerrors.ErrExternalActivityConflict) {
		t.Fatalf("changed payload error=%v", err)
	}
	var count int
	if err := repo.db.Get(&count, `SELECT COUNT(*) FROM task_external_activity`); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("namespace records=%d want 2", count)
	}
}
