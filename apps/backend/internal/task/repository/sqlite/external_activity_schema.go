package sqlite

import "fmt"

func (r *Repository) initTaskExternalActivitySchema() error {
	_, err := r.db.ExecContext(r.migrationContext(), `
	CREATE TABLE IF NOT EXISTS task_external_activity (
		id TEXT PRIMARY KEY,
		task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
		workspace_id TEXT NOT NULL,
		source_instance TEXT NOT NULL,
		event_id TEXT NOT NULL,
		source TEXT NOT NULL,
		kind TEXT NOT NULL,
		phase TEXT NOT NULL,
		outcome TEXT NOT NULL,
		provider TEXT NOT NULL,
		model TEXT NOT NULL,
		duration_ms BIGINT CHECK (duration_ms >= 0),
		input_tokens BIGINT CHECK (input_tokens >= 0),
		output_tokens BIGINT CHECK (output_tokens >= 0),
		cached_input_tokens BIGINT CHECK (cached_input_tokens >= 0),
		reasoning_output_tokens BIGINT CHECK (reasoning_output_tokens >= 0),
		candidate_sha TEXT NOT NULL,
		payload_hash TEXT NOT NULL,
		payload TEXT NOT NULL,
		writer_id TEXT NOT NULL,
		occurred_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL
	);
	CREATE UNIQUE INDEX IF NOT EXISTS uniq_external_activity_workspace_event
		ON task_external_activity(workspace_id, source_instance, event_id);
	CREATE UNIQUE INDEX IF NOT EXISTS uniq_external_activity_task_event
		ON task_external_activity(task_id, source_instance, event_id);
	CREATE INDEX IF NOT EXISTS idx_external_activity_task_time
		ON task_external_activity(task_id, occurred_at);
	CREATE INDEX IF NOT EXISTS idx_external_activity_workspace_time
		ON task_external_activity(workspace_id, occurred_at);
	`)
	if err != nil {
		return fmt.Errorf("init external activity schema: %w", err)
	}
	return nil
}
