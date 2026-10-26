package models

import (
	"errors"
	"math"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
)

// ExternalActivity is a completed observation from an external coordinator.
// Input/output totals are inclusive; cache and reasoning are informational subsets.
type ExternalActivity struct {
	SchemaVersion         int       `json:"schema_version"`
	Source                string    `json:"source"`
	SourceInstance        string    `json:"source_instance"`
	EventID               string    `json:"event_id"`
	Kind                  string    `json:"kind"`
	Phase                 string    `json:"phase,omitempty"`
	Outcome               string    `json:"outcome"`
	OccurredAt            time.Time `json:"occurred_at"`
	Provider              string    `json:"provider,omitempty"`
	Model                 string    `json:"model,omitempty"`
	DurationMs            *int64    `json:"duration_ms,omitempty"`
	InputTokens           *int64    `json:"input_tokens,omitempty"`
	OutputTokens          *int64    `json:"output_tokens,omitempty"`
	CachedInputTokens     *int64    `json:"cached_input_tokens,omitempty"`
	ReasoningOutputTokens *int64    `json:"reasoning_output_tokens,omitempty"`
	CandidateSHA          string    `json:"candidate_sha,omitempty"`
}

var ErrInvalidExternalActivity = errors.New("invalid external activity")

const externalReviewCompleted = "review_completed"

var activityIdentityPattern = regexp.MustCompile(`^[A-Za-z0-9_.:-]{1,128}$`)
var activityCommitPattern = regexp.MustCompile(`^[0-9a-f]{40}([0-9a-f]{24})?$`)

// Validate checks measured counts and kind-specific fields without inferring usage.
func (a ExternalActivity) Validate(now time.Time) error {
	if !a.validActivityIdentity() || !a.validActivityTime(now) {
		return ErrInvalidExternalActivity
	}
	if a.Kind == "worker_run" {
		return a.validateWorkerActivity()
	}
	return a.validateResultActivity()
}

func (a ExternalActivity) validActivityIdentity() bool {
	instance, err := uuid.Parse(a.SourceInstance)
	if err != nil || instance.String() != a.SourceInstance || instance == uuid.Nil {
		return false
	}
	return a.SchemaVersion == 1 && a.Source != "" && boundedActivityText(a.Source, 128) &&
		activityIdentityPattern.MatchString(a.EventID) && boundedActivityText(a.Provider, 256) && boundedActivityText(a.Model, 256)
}

func (a ExternalActivity) validActivityTime(now time.Time) bool {
	if a.OccurredAt.IsZero() || a.OccurredAt.Year() < 1 || a.OccurredAt.After(now.Add(5*time.Minute)) {
		return false
	}
	return a.CandidateSHA == "" || activityCommitPattern.MatchString(a.CandidateSHA)
}

func (a ExternalActivity) validateResultActivity() error {
	if a.Phase != "" || a.Provider != "" || a.Model != "" || a.DurationMs != nil ||
		a.InputTokens != nil || a.OutputTokens != nil || a.CachedInputTokens != nil || a.ReasoningOutputTokens != nil {
		return ErrInvalidExternalActivity
	}
	switch a.Kind {
	case externalReviewCompleted:
		if a.Outcome == string(InteractionStatusApproved) || a.Outcome == "rejected" {
			return nil
		}
	case "qa_completed":
		if a.Outcome == "passed" || a.Outcome == "failed" {
			return nil
		}
	}
	return ErrInvalidExternalActivity
}

func (a ExternalActivity) validateWorkerActivity() error {
	switch a.Phase {
	case "coding", "review", "recovery", "unknown":
	default:
		return ErrInvalidExternalActivity
	}
	switch a.Outcome {
	case "succeeded", "failed", "timeout", "unknown":
	default:
		return ErrInvalidExternalActivity
	}
	return a.validateWorkerCounts()
}

func (a ExternalActivity) validateWorkerCounts() error {
	for _, count := range []*int64{a.InputTokens, a.OutputTokens, a.CachedInputTokens, a.ReasoningOutputTokens, a.DurationMs} {
		if count != nil && *count < 0 {
			return ErrInvalidExternalActivity
		}
	}
	if a.InputTokens != nil && a.CachedInputTokens != nil && *a.CachedInputTokens > *a.InputTokens {
		return ErrInvalidExternalActivity
	}
	if a.OutputTokens != nil && a.ReasoningOutputTokens != nil && *a.ReasoningOutputTokens > *a.OutputTokens {
		return ErrInvalidExternalActivity
	}
	if a.InputTokens != nil && a.OutputTokens != nil && *a.InputTokens > math.MaxInt64-*a.OutputTokens {
		return ErrInvalidExternalActivity
	}
	return nil
}

func boundedActivityText(value string, limit int) bool {
	if len(value) > limit || !utf8.ValidString(value) || strings.TrimSpace(value) != value {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}

// ExternalActivityReceipt acknowledges an insertion or an identical replay.
type ExternalActivityReceipt struct {
	ID      string `json:"id"`
	Created bool   `json:"created"`
}
