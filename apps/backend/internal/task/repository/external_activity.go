package repository

import (
	"context"

	"github.com/kandev/kandev/internal/task/models"
	"github.com/kandev/kandev/internal/task/repository/repoerrors"
)

var ErrExternalActivityConflict = repoerrors.ErrExternalActivityConflict

// ExternalActivityWriter records immutable externally executed work.
type ExternalActivityWriter interface {
	RecordExternalActivity(context.Context, string, string, string, *models.ExternalActivity) (models.ExternalActivityReceipt, error)
}
