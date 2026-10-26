---
title: External coordinator accounting
description: Record external worker usage, review iterations, and QA results against Kandev tasks.
---

# External coordinator accounting

External coordinators can append completed worker, review, and QA observations
to existing tasks. These records feed usage statistics without creating native
agent sessions or changing task state. The API uses the deployment's existing
authentication; writers need task-write access and readers need workspace access.

## Append an observation

Send `POST /api/v1/tasks/{task_id}/external-activity` with
`Content-Type: application/json`. The body must be one JSON object, at most
16 KiB, with no unknown fields. Do not send prompts, transcripts, credentials,
or worker output.

```json
{
  "schema_version": 1,
  "source": "kodi",
  "source_instance": "d2d93284-5aab-4384-b50c-ac5bebc8e133",
  "event_id": "worker:123",
  "kind": "worker_run",
  "phase": "coding",
  "outcome": "succeeded",
  "occurred_at": "2026-01-02T01:30:00Z",
  "provider": "example-provider",
  "model": "example-model",
  "input_tokens": 100,
  "output_tokens": 20,
  "cached_input_tokens": 40,
  "duration_ms": 1200
}
```

Keep a persistent, nonzero, canonical UUID for `source_instance` and a stable
`event_id` for each observation. Event IDs accept 1–128 ASCII letters, digits,
periods, underscores, colons, and hyphens. Reuse the same identity and payload
when retrying. A committed insertion returns `201`; an identical replay returns
`200` with the same receipt `id`. Both return `created` to distinguish insertion
from replay. A changed payload or task under an existing identity returns `409`.
A task move preserves its receipts and original ingestion namespace.

Required fields are `schema_version` (currently `1`), `source` (1–128 bytes),
`source_instance`, `event_id`, `kind`, `outcome`, and `occurred_at`.
Timestamps accept RFC 3339 offsets, are normalized to UTC, and may be at most
five minutes in the future. Optional `candidate_sha` accepts a lowercase
40- or 64-character hexadecimal commit ID.

| Kind | Required phase | Outcomes | Additional fields |
| --- | --- | --- | --- |
| `worker_run` | `coding`, `review`, `recovery`, or `unknown` | `succeeded`, `failed`, `timeout`, `unknown` | Optional provider/model, duration, and token counts |
| `review_completed` | None | `approved`, `rejected` | Optional candidate SHA |
| `qa_completed` | None | `passed`, `failed` | Optional candidate SHA |

Provider and model strings accept at most 256 bytes. Identity and descriptive
strings must be valid UTF-8 without surrounding whitespace or control characters.
Duration and token fields are nonnegative signed 64-bit integers. Omit an unknown
measurement or send `null`; send `0` only for a measured zero. Input and output
counts are inclusive totals. `cached_input_tokens` and
`reasoning_output_tokens` are subsets and are never added again. A known subset
cannot exceed its known parent count.

Invalid input returns `400`, oversized bodies `413`, and unsupported content types
`415`. Missing or inaccessible tasks return `404`; insufficient write scope
returns `403`. Temporary storage unavailability can return `503`. Retry transport
failures with the original identity and body; investigate `409` before resending.

## Read usage statistics

Send `GET /api/v1/workspaces/{workspace_id}/stats/usage`.

| Query | Values and defaults |
| --- | --- |
| `source` | `all` (default), `native`, `external` |
| `range` | `month` (default, 30 UTC days), `week` (7 UTC days), `all` |
| `limit` | 1–100, default 50, for each detail page |
| `task_cursor` | Previous response's `next_task_cursor` |
| `provider_cursor` | Previous response's `next_provider_cursor` |

The response contains `totals`, `daily`, `tasks`, `providers`, and independent
continuation cursors for tasks and providers. Cursors are empty at the end of a
list. Retain the same source and range while paging. Summary and daily results
remain independent of detail cursors.

Day buckets use UTC calendar days, including today. `daily_from` is inclusive;
`daily_until` is exclusive. The `all` range includes older backfilled observations
in totals and detail pages while limiting the daily series to the latest 365 days.
Task results include eligible tasks with no observations or native sessions.
Task history follows the task's current workspace access.

`known_tokens_total` sums available contributions. `tokens_total` is `null` when
any included observation has incomplete usage. `incomplete_run_count` and
`incomplete_native_event_count` show why a complete total is unavailable.
`run_count` counts external worker runs; `native_event_count` counts native ledger
observations. `review_iterations` and `qa_runs` count completed results, including
rejections, failures, and distinct reruns against the same commit.
`worker_duration_ms` sums reported external worker duration.

Native ledger accounting retains its existing provider semantics. The API does
not estimate monetary charges or infer missing historical measurements. Invalid
queries return `400`; inaccessible workspaces return `404`. Busy reads return
`503` with `Retry-After: 2`.
