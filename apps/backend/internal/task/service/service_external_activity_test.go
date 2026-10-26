package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kandev/kandev/internal/task/models"
	"github.com/kandev/kandev/internal/task/repository/repoerrors"
)

func TestExternalActivityServiceRequiresTaskWrite(t *testing.T) {
	svc, _, repo := createTestService(t)
	seedScopedWorkspaces(t, repo)
	if err := repo.UpsertWorkspaceMember(context.Background(), &models.WorkspaceMember{
		WorkspaceID: "ws-b", UserID: "viewer", Role: "viewer",
	}); err != nil {
		t.Fatal(err)
	}
	activity := &models.ExternalActivity{SchemaVersion: 1, Source: "kodi",
		SourceInstance: "d2d93284-5aab-4384-b50c-ac5bebc8e133", EventID: "usage:1",
		Kind: "worker_run", Phase: "coding", Outcome: "succeeded",
		OccurredAt: time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC)}
	for _, test := range []struct {
		name, user, task string
		want             error
	}{
		{"foreign", "user-a", "task-b", repoerrors.ErrTaskNotFound},
		{"missing", "user-b", "absent", repoerrors.ErrTaskNotFound},
		{"read only", "viewer", "task-b", ErrForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := svc.RecordExternalActivity(ctxAs(test.user), test.task, activity)
			if !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
		})
	}
	invalid := *activity
	negative := int64(-1)
	invalid.InputTokens = &negative
	if _, err := svc.RecordExternalActivity(ctxAs("user-b"), "task-b", &invalid); !errors.Is(err, models.ErrInvalidExternalActivity) {
		t.Fatalf("invalid measurement: %v", err)
	}
	first, err := svc.RecordExternalActivity(ctxAs("user-b"), "task-b", activity)
	if err != nil || !first.Created {
		t.Fatalf("owner admission: %+v %v", first, err)
	}
	replay, err := svc.RecordExternalActivity(ctxAs("user-b"), "task-b", activity)
	if err != nil || replay.Created || replay.ID != first.ID {
		t.Fatalf("owner replay: %+v %v", replay, err)
	}
}
