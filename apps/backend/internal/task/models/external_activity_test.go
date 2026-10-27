package models

import (
	"math"
	"strings"
	"testing"
	"time"
)

func externalActivityFixture() ExternalActivity {
	return ExternalActivity{SchemaVersion: 1, Source: "kodi", SourceInstance: "d2d93284-5aab-4384-b50c-ac5bebc8e133",
		EventID: "usage:42", Kind: "worker_run", Phase: "coding", Outcome: "succeeded",
		OccurredAt: time.Date(2026, 10, 25, 23, 0, 0, 0, time.UTC)}
}

func activityCount(value int64) *int64 { return &value }

// Covers AC-COSTS-EXTERNAL-002.1 and AC-COSTS-EXTERNAL-002.2.

func TestExternalActivityRejectsInvalidObservations(t *testing.T) {
	tests := []struct {
		name string
		edit func(*ExternalActivity)
	}{
		{"version", func(a *ExternalActivity) { a.SchemaVersion = 2 }},
		{"namespace", func(a *ExternalActivity) { a.SourceInstance = "not-a-uuid" }},
		{"empty identity", func(a *ExternalActivity) { a.EventID = "" }},
		{"oversized identity", func(a *ExternalActivity) { a.EventID = strings.Repeat("x", 129) }},
		{"negative", func(a *ExternalActivity) { a.InputTokens = activityCount(-1) }},
		{"overflow", func(a *ExternalActivity) {
			a.InputTokens = activityCount(math.MaxInt64)
			a.OutputTokens = activityCount(1)
		}},
		{"cache subset", func(a *ExternalActivity) { a.InputTokens = activityCount(10); a.CachedInputTokens = activityCount(11) }},
		{"reasoning subset", func(a *ExternalActivity) {
			a.OutputTokens = activityCount(0)
			a.ReasoningOutputTokens = activityCount(1)
		}},
		{"unknown outcome", func(a *ExternalActivity) { a.Outcome = "approved" }},
		{"unknown phase", func(a *ExternalActivity) { a.Phase = "autopilot" }},
		{"missing time", func(a *ExternalActivity) { a.OccurredAt = time.Time{} }},
		{"future time", func(a *ExternalActivity) { a.OccurredAt = a.OccurredAt.Add(6 * time.Minute) }},
		{"invalid candidate", func(a *ExternalActivity) { a.CandidateSHA = "main" }},
		{"review with usage", func(a *ExternalActivity) {
			a.Kind = "review_completed"
			a.Phase = ""
			a.Outcome = "approved"
			a.InputTokens = activityCount(1)
		}},
		{"review with provider", func(a *ExternalActivity) {
			a.Kind = "review_completed"
			a.Phase = ""
			a.Outcome = "approved"
			a.Provider = "codex"
		}},
		{"control characters", func(a *ExternalActivity) { a.Provider = "codex\nspoof" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			activity := externalActivityFixture()
			now := activity.OccurredAt
			test.edit(&activity)
			if activity.Validate(now) == nil {
				t.Fatal("invalid observation accepted")
			}
		})
	}
}

func TestExternalActivityAcceptsUnknownZeroAndInclusiveSubsets(t *testing.T) {
	for _, output := range []*int64{nil, activityCount(0), activityCount(20)} {
		activity := externalActivityFixture()
		activity.InputTokens = activityCount(100)
		activity.CachedInputTokens = activityCount(60)
		activity.OutputTokens = output
		if err := activity.Validate(activity.OccurredAt); err != nil {
			t.Fatal(err)
		}
	}
	for kind, outcome := range map[string]string{"review_completed": "rejected", "qa_completed": "failed"} {
		activity := externalActivityFixture()
		activity.Kind, activity.Outcome, activity.Phase = kind, outcome, ""
		if err := activity.Validate(activity.OccurredAt); err != nil {
			t.Fatal(err)
		}
	}
}
