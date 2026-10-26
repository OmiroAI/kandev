---
status: draft
system: costs
requirements:
  - REQ-COSTS-EXTERNAL-001
  - REQ-COSTS-EXTERNAL-002
  - REQ-COSTS-EXTERNAL-003
  - REQ-COSTS-EXTERNAL-004
---

# External worker statistics system design

## Ownership and existing boundaries

The costs system owns this contract. Task service/storage own task access and
persistence. `internal/analytics` owns Stats reads; the existing React Stats
page owns display. The external producer is a separate coordinator such as Kodi.

Verified starting points:

- `internal/analytics/handlers/stats_handlers.go` registers read-only workspace
  statistics routes; `StatsHandlers.parseRequest` resolves the existing range.
- `internal/analytics/models/models.go` has task, daily, and model activity but
  no external-run or review iteration accounting.
- `internal/task/models/usage_event.go` and the task repository own native
  `task_usage_events`; `CreateTaskUsageEvent` also updates native session rollups.
- The native ledger requires `TokensIn int64` and uses its established disjoint
  components. A completely unknown external run cannot be represented there
  honestly without changing that established contract.
- `internal/task/service/service_usage.go` and task handlers are existing
  task-scoped accounting boundaries. New writes must use task-write authority,
  not just the read authorization used for usage queries.
- `apps/web/app/stats` and `lib/api/domains/stats-api.ts` supply the existing UI.

## Requirement mapping

| Requirement | Design sections |
| --- | --- |
| REQ-COSTS-EXTERNAL-001 | Ingestion, persistence, access |
| REQ-COSTS-EXTERNAL-002 | Measurements, aggregation |
| REQ-COSTS-EXTERNAL-003 | Read API, presentation |
| REQ-COSTS-EXTERNAL-004 | Coordinator adapter, recovery |

## Ingestion

Add `POST /api/v1/tasks/:id/external-activity` through task handlers/service.
A request is one versioned immutable event, at most 16 KiB; reject unknown
fields. The client cannot submit workspace, session, native turn, credential,
price, prompt, or arbitrary metadata fields.

Proposed v1 fields:

| Field | Meaning |
| --- | --- |
| `schema_version` | Integer `1` |
| `source` | Bounded display label such as `kodi`; not an authorization claim |
| `source_instance` | Stable UUID generated once per external ledger |
| `event_id` | Bounded stable string, independent of transport attempt |
| `kind` | `worker_run`, `review_completed`, or `qa_completed` |
| `occurred_at` | RFC3339 completion timestamp, normalized to UTC |
| `phase` | `coding`, `review`, `recovery`, or `unknown` for worker runs |
| `outcome` | Worker: `succeeded`, `failed`, `timeout`, or historical `unknown`; review: `approved` or `rejected`; QA: `passed` or `failed` |
| `provider`, `model` | Optional bounded strings, only for worker runs |
| `duration_ms` | Optional nonnegative signed-64-bit integer, worker runs only |
| `input_tokens`, `output_tokens` | Nullable nonnegative signed-64-bit integers |
| `cached_input_tokens`, `reasoning_output_tokens` | Nullable informational subsets |
| `candidate_sha` | Optional verified commit identifier for review/QA attribution |

Token fields are allowed only for worker runs. No inference of the source from
model names or provider names. Limit identifiers to 128 characters and provider/
model values to 256; reject invalid kind/outcome combinations and timestamps
more than five minutes ahead of server time. Old valid timestamps are allowed
for backfill. Capture a canonical payload digest after normalization.

Return 201 for insertion and 200 for an identical existing event, with its
server identity. Return 409 for an identity with different canonical content,
400 for invalid input, 413 for body size, and the established inaccessible-task
response for missing/forbidden ownership. Database/unavailable errors return
503 and never acknowledge an uncommitted write.

## Persistence and access

Add an additive `task_external_activity` table in task persistence, implemented
for both supported database backends using the existing dialect/migration
conventions. Required columns hold the normalized fields, task/workspace
identity, authenticated writer identity, server creation time, and payload
digest. All token columns are nullable. Index task/time and workspace/time.
Enforce unique `(workspace_id, source_instance, event_id)` in the database so
concurrent identical retries are safe. Task identity participates in the digest;
a retry against another task in the same workspace is a conflict.

Authorize task-write scope before duplicate lookup or insert using the task
service's existing scope helpers. A visible but non-writable task remains
non-writable. Workspace membership, PAT scopes, and auth-disabled local mode
retain existing semantics; no new secret or authentication scheme is added.
The supplied source UUID is an accounting namespace, never a capability.
Recheck ownership through the normal service/repository transaction boundary.
Task deletion cascades; archive retains events. The ingestion workspace is
an immutable namespace snapshot, not an authorization source or a cascading
workspace foreign key: deletion of a former workspace must not erase activity
for a transferred task. Reads authorize and filter the task's current workspace.
A second unique task/source/event key keeps replay idempotent after transfer;
workspace-scoped collisions on another task still conflict. Include the table
in the existing task-transfer preservation catalog. Immutable insertion has no
public update/delete route. No event-bus replay through native lifecycle.

This separate external activity source is deliberate: it records worker runs
and delivery results that do not have native session/turn identity and preserves
fully unknown input counts. Native ledger readers and pricing remain unchanged.
Stats uses a combined read projection; neither source is copied into the other.

## Measurements and aggregation

External v1 uses inclusive `input_tokens` (including cached input) and inclusive
`output_tokens` (including reasoning). Subset fields are never added to totals.
Validate subsets against their known parent; allow a known subset with an
unknown parent but keep the total incomplete. Check addition for int64 overflow.

Each aggregate returns nullable complete `tokens_total`, `known_tokens_total`
(the sum of known input/output components), `incomplete_run_count`, run count,
duration, review iterations, and QA results. An unknown run contributes a run
and its known components, never an invented zero or complete total. Empty
scopes return measured activity count zero, distinct from failed reads.
Native values come from `task_usage_events` with its existing disjoint-component
semantics and completeness/provenance, not from summing message snapshots.
Native and external subtotals remain visible when selecting all sources.
Count external review/QA event identities, not commit SHAs, moves between board
columns, or HTTP attempts. This yields two iterations for two verdicts on the
same commit but one for repeated delivery of one verdict.

## Read API

Add authorized `GET /api/v1/workspaces/:id/stats/usage` with existing
`range=week|month|all`, `source=all|native|external`, and bounded task pagination
(`limit` 1..100, default 50; stable cursor ordered by task ID). Authorize through
the same workspace gate as existing Stats routes. Return totals, UTC daily
series, provider/model breakdowns, and a task page with IDs/titles and accounting
columns. Task/provider sections have independent continuation cursors so neither
is silently truncated. Daily all-range data is bounded to the existing 365-day
chart horizon and explicitly returns its start/end; all-time summary/task
aggregates remain all-time. Existing analytics admission/query budgets apply.

Do not join event facts directly to multi-valued sessions or repositories:
aggregate each fact source before joining task metadata to avoid multiplied
counts. A task with external events and no sessions must remain present.
Repository filters use the actual task repository relations. No integration
field may grant access across workspaces. Native chat totals APIs retain their
wire shapes and meaning.

## Presentation

Extend existing Stats with a usage section and source selector. Reuse the range
control, chart primitives, task links, and design tokens. Show complete totals
or a known subtotal with an explicit incomplete-run count, never a false zero.
Include provider/model rows and external worker, review, and QA columns in the
task view. Label native turns separately. Empty, loading, unavailable, partial,
and populated states have localized copy and keyboard-accessible controls.
Retain known data while a refresh fails; expose retry without disabling other
Stats sections. Phone composition uses stacked task cards rather than an
unreadably wide table; daily charts have text summaries for accessibility.
Update English, pseudo, and all required locale catalogs through existing tools.

## Coordinator adapter and recovery

Kodi keeps SQLite usage/evidence as its source of truth. Add Board ingestion
methods plus a bounded coordinator sync of immutable usage/evidence IDs. Generate
one persisted source-instance UUID for that ledger, not one per process/host
name; store it privately. Event IDs are `usage:<row-id>` and
`evidence:<row-id>`. Both evidence kinds reuse the same evidence ID namespace.
Never derive IDs from mutable attempt count, current SHA, title, or board step.

For historical usage, absent role/outcome metadata becomes `unknown`. Add
explicit run metadata at new recording call sites; do not infer old roles from
provider names. Keep the existing ledger provider payload and its timestamps.
Normalize provider fields conservatively at the adapter boundary; unsupported
shapes produce unknown components. Preserve counts for failed invocations.

Send to the job's persisted board task ID. Persist acknowledgment/cursor only
after 200/201. On timeout or 5xx, keep the event pending and retry with the same
identity. Bound each tick to 100 events and a transport timeout; use a durable
outbox to prevent an old failed task starving new tasks. Quarantine 400/409
with an operator-visible reason rather than silently advancing. No raw server
body enters task descriptions or logs. Missing/disabled tasks remain pending
for explicit reconciliation, not remapped by title.

Detect an older server's missing endpoint once per bounded cooldown, leave
entries pending, and preserve existing description updates. Accounting sync
failure must not enter process failure handling or modify job acceptance state.
Use the same adapter for explicit bounded historical backfill. Sending ledger
rows does not require native agents, provider credentials, or production access.

## Verification and rollout

Test authenticated write/read boundaries, owner read-only scope, duplicate and
conflicting retries under concurrency, transaction rollback, SQLite/PostgreSQL
parity, UTC offsets, inclusive subsets, overflows, and partial measurements.
Render desktop/phone states and exercise an actual HTTP replay after a dropped
response. Keep existing native usage regression suites green.

Prepare a dedicated offline QA image containing pinned Go 1.26, Node 24, pnpm
9.15.9, locked Go/modules and web dependencies, and browser assets for the scoped
checks. QA must run with networking disabled and verify lock identities. The
current generic Python QA image is insufficient; never enable repository
dispatch based on an inventory/static-only baseline. Both repositories still
require exact-commit independent QA, different-model review, CI, and human
approval. Publishing and switching the dashboard image remain a later operator
rollout; the public fork retains upstream AGPL notices.
