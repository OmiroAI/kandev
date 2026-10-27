---
id: "02-activity-api"
title: "External activity API and projections"
status: in_progress
wave: 2
depends_on: ['01-fork-onboarding']
plan: "plan.md"
requirements:
  - REQ-COSTS-EXTERNAL-001
  - REQ-COSTS-EXTERNAL-002
  - REQ-COSTS-EXTERNAL-003
acceptance_criteria:
  - AC-COSTS-EXTERNAL-001.1
  - AC-COSTS-EXTERNAL-001.2
  - AC-COSTS-EXTERNAL-001.3
  - AC-COSTS-EXTERNAL-001.4
  - AC-COSTS-EXTERNAL-002.1
  - AC-COSTS-EXTERNAL-002.2
  - AC-COSTS-EXTERNAL-002.3
  - AC-COSTS-EXTERNAL-002.4
  - AC-COSTS-EXTERNAL-003.1
  - AC-COSTS-EXTERNAL-003.2
  - AC-COSTS-EXTERNAL-003.4
system_design:
  - ../../specs/costs/system-design/external-worker-statistics.md
---

# Task 02: External activity API and projections

## Summary

Implement authorized append-only external activity and an aggregate Stats read surface without native sessions. Preserve existing native accounting contracts.

## In scope

- Versioned bounded ingestion, task-write checks, immutable canonical deduplication and conflicts.
- SQLite/PostgreSQL migration/repository parity, nullable inclusive token components and bounded grouped reads.
- UTC/source/range/task pagination; explicit incomplete data and no join multiplication.

## Out of scope

- React UI, Kodi transport, pricing, native session creation.

## Acceptance

- Identical concurrent retries count once; unauthorized, invalid and conflicting writes have no partial effects.
- Queries include external tasks without sessions and preserve unknown/zero/subset semantics.
- Existing native usage and session/turn statistics retain their meaning.

## Verification

Run from the Kandev repository root unless stated otherwise. Setup downloads
belong only in the image-build phase; independent QA has networking disabled.

```bash
(cd apps/backend && go test ./internal/task/models ./internal/task/dto ./internal/task/repository/sqlite ./internal/task/service ./internal/task/handlers ./internal/task/usage ./internal/analytics/...)
python3 scripts/list-docs.py validate
python3 scripts/lint-spec-files.py --all
git diff --check
```

## Files likely touched

- `apps/backend/internal/task/models/external_activity.go`
- `apps/backend/internal/task/repository/sqlite/external_activity.go`
- `apps/backend/internal/task/service/service_external_activity.go`
- `apps/backend/internal/task/handlers/task_external_activity_handlers.go`
- `apps/backend/internal/task/dto/external_activity.go`
- `apps/backend/internal/analytics/models/models.go`
- `apps/backend/internal/analytics/repository/interface.go`
- `apps/backend/internal/analytics/repository/sqlite/external_usage.go`
- `apps/backend/internal/analytics/handlers/stats_handlers.go`
- `apps/backend/internal/task/repository/sqlite/external_activity_migrations.go`

## Dependencies

Prior work order: 01-fork-onboarding. Its local toolchain, scoped offline
acceptance, Gemini review and software CI are complete. Operator/independent
acceptance remains a dispatch gate, not a prerequisite for reversible local
implementation on this stacked development branch.

## Risks

Follow `usage_events_migrations.go` and `base_migrations.go` for migration registration; retain both database backends. Run real PostgreSQL parity tests using the existing test database harness, not only SQLite mocks.

## Parallelism

`sequential`

## Inputs

- [Requirements](../../specs/costs/requirements/external-worker-statistics.md)
- [System design](../../specs/costs/system-design/external-worker-statistics.md)
- [Full plan](plan.md)

## Results

Implemented the append-only task activity ledger, authorized strict JSON write
route, and workspace usage projections. Replay identities survive retries and
task transfers. Current workspace ownership controls reads; ingestion namespace
snapshots remain immutable. Native session and usage contracts stay separate.

Focused red/green tests cover validation, concurrent replay, conflicts, task
scope, HTTP limits, known/unknown counts, independent cursors, UTC dates,
backfill, and current workspace ownership. Actual PostgreSQL storage and usage
projection tests passed, including a non-UTC database session. SQL guard,
specification validation, public documentation validation, and dependency lock
regressions passed locally.

All planned task and analytics regression packages passed locally. The initial
full run hit its default timeout while creating temporary databases on disk;
in-memory scratch completed the full task-store package within that timeout.
Service/handler directory tests additionally required scratch outside the
shared-writable checkout parent. Their full packages passed with private trusted
scratch; production checks and sandbox security settings were unchanged.

The required persistence store-conformance suite passed with `-race` on SQLite
and with the actual isolated PostgreSQL fixture. This includes the tagged
previous-stable upgrade path. Accounting-specific race checks also passed across models, repositories,
services, handlers, and analytics, including actual PostgreSQL concurrent
retries. A final boundary regression caught JSON escaping expanding a valid
provider cursor beyond its input guard; the guard now accommodates the bounded
escaped payload, and all analytics tests passed after the fix. The final scoped race rerun and Go lint passed with the fix included.
A transfer regression on SQLite and PostgreSQL also verifies that a valid
original task receipt wins over an unrelated destination-workspace event with
the same source identity. Changed payloads still conflict and no new row is
created by the replay. Local implementation and work-order checks are complete. Exact-commit
independent QA, different-model review, CI, and human approval remain pending. Public API reference:
`docs/public/external-coordinators.md`. This work order remains in progress;
operator acceptance of the fork still gates dispatch activation.
