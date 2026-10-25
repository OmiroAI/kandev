---
id: "02-activity-api"
title: "External activity API and projections"
status: pending
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

Prior work orders: ['01-fork-onboarding'].

## Risks

Follow `usage_events_migrations.go` and `base_migrations.go` for migration registration; retain both database backends. Run real PostgreSQL parity tests using the existing test database harness, not only SQLite mocks.

## Parallelism

`sequential`

## Inputs

- [Requirements](../../specs/costs/requirements/external-worker-statistics.md)
- [System design](../../specs/costs/system-design/external-worker-statistics.md)
- [Full plan](plan.md)

## Results

Pending.
