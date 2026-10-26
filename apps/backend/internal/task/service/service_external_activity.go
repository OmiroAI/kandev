package service

import (
	"context"
	"errors"

	"github.com/kandev/kandev/internal/auth/authn"
	"github.com/kandev/kandev/internal/authz"
	"github.com/kandev/kandev/internal/task/models"
	"github.com/kandev/kandev/internal/task/repository"
	"github.com/kandev/kandev/internal/task/repository/repoerrors"
)

var errExternalActivityUnavailable = errors.New("external activity storage is unavailable")

// RecordExternalActivity authorizes the same workspace snapshot the writer locks.
func (s *Service) RecordExternalActivity(ctx context.Context, taskID string, activity *models.ExternalActivity) (models.ExternalActivityReceipt, error) {
	var empty models.ExternalActivityReceipt
	task, err := s.tasks.GetTask(ctx, taskID)
	if err != nil {
		return empty, err
	}
	if task == nil || task.WorkspaceID == "" {
		return empty, repoerrors.ErrTaskNotFound
	}
	_, err = s.requireWorkspaceScope(ctx, task.WorkspaceID, authz.ScopeTaskWrite, repoerrors.ErrTaskNotFound)
	if errors.Is(err, repoerrors.ErrWorkspaceNotFound) {
		return empty, repoerrors.ErrTaskNotFound
	}
	if err != nil {
		return empty, err
	}
	writer, ok := s.tasks.(repository.ExternalActivityWriter)
	if !ok {
		return empty, errExternalActivityUnavailable
	}
	identity, _ := authn.IdentityFromContext(ctx)
	return writer.RecordExternalActivity(ctx, task.ID, task.WorkspaceID, identity.UserID, activity)
}
