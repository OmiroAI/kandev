package sqlite

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/kandev/kandev/internal/db/dialect"
	"github.com/kandev/kandev/internal/task/models"
	"github.com/kandev/kandev/internal/task/repository/repoerrors"
)

// RecordExternalActivity binds immutable observations to a task without native sessions.
func (r *Repository) RecordExternalActivity(ctx context.Context, taskID, workspaceID, writerID string, activity *models.ExternalActivity) (models.ExternalActivityReceipt, error) {
	var receipt models.ExternalActivityReceipt
	if activity == nil {
		return receipt, models.ErrInvalidExternalActivity
	}
	if err := activity.Validate(time.Now()); err != nil {
		return receipt, err
	}
	observation := *activity
	observation.OccurredAt = observation.OccurredAt.UTC()
	payload, err := json.Marshal(struct {
		TaskID   string                  `json:"task_id"`
		Activity models.ExternalActivity `json:"activity"`
	}{taskID, observation})
	if err != nil {
		return receipt, err
	}
	digest := sha256.Sum256(payload)
	hash := hex.EncodeToString(digest[:])
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return receipt, err
	}
	defer func() { _ = tx.Rollback() }()
	if err := r.lockExternalActivityTask(ctx, tx, taskID, workspaceID); err != nil {
		return receipt, err
	}
	receipt.ID = uuid.NewString()
	result, err := tx.ExecContext(ctx, r.db.Rebind(`INSERT INTO task_external_activity (
		id,task_id,workspace_id,source_instance,event_id,source,kind,phase,outcome,
		provider,model,duration_ms,input_tokens,output_tokens,cached_input_tokens,
		reasoning_output_tokens,candidate_sha,payload_hash,payload,writer_id,occurred_at,created_at
	) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING`),
		receipt.ID, taskID, workspaceID, observation.SourceInstance, observation.EventID,
		observation.Source, observation.Kind, observation.Phase, observation.Outcome,
		observation.Provider, observation.Model, observation.DurationMs,
		observation.InputTokens, observation.OutputTokens, observation.CachedInputTokens,
		observation.ReasoningOutputTokens, observation.CandidateSHA, hash, string(payload), writerID,
		observation.OccurredAt, time.Now().UTC())
	if err != nil {
		return models.ExternalActivityReceipt{}, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return models.ExternalActivityReceipt{}, err
	}
	receipt.Created = count == 1
	if !receipt.Created {
		receipt.ID, err = r.externalActivityReplay(ctx, tx, taskID, workspaceID, &observation, hash)
		if err != nil {
			return models.ExternalActivityReceipt{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return models.ExternalActivityReceipt{}, err
	}
	return receipt, nil
}

func (r *Repository) lockExternalActivityTask(ctx context.Context, tx *sqlx.Tx, taskID, workspaceID string) error {
	query := `SELECT id FROM tasks WHERE id=? AND workspace_id=?`
	if dialect.IsPostgres(r.db.DriverName()) {
		query += forUpdateClause
	}
	var found string
	err := tx.QueryRowContext(ctx, r.db.Rebind(query), taskID, workspaceID).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return repoerrors.ErrTaskNotFound
	}
	return err
}

func (r *Repository) externalActivityReplay(ctx context.Context, tx *sqlx.Tx, taskID, workspaceID string, activity *models.ExternalActivity, hash string) (string, error) {
	var id, owner, recordedHash string
	err := tx.QueryRowContext(ctx, r.db.Rebind(`SELECT id,task_id,payload_hash FROM task_external_activity
  WHERE (workspace_id=? AND source_instance=? AND event_id=?)
  OR (task_id=? AND source_instance=? AND event_id=?)
  ORDER BY CASE WHEN task_id=? THEN 0 ELSE 1 END LIMIT 1`),
		workspaceID, activity.SourceInstance, activity.EventID, taskID, activity.SourceInstance, activity.EventID, taskID).
		Scan(&id, &owner, &recordedHash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", errors.New("external activity insertion failed")
	}
	if err != nil {
		return "", err
	}
	if owner != taskID || recordedHash != hash {
		return "", repoerrors.ErrExternalActivityConflict
	}
	return id, nil
}
