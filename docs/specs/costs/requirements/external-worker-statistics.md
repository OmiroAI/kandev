---
status: draft
system: costs
created: 2026-10-25
owners:
  - kandev
---

# External worker usage and delivery statistics

## Overview

An external coordinator can report work performed outside Kandev so operators
can inspect daily token usage, task usage, and delivery iterations in Stats.
The costs system owns measurement provenance, deduplication, aggregation, and
these views. Task identity and access remain owned by the task system.
This complements [conversation usage](conversation-usage.md); it does not
manufacture a native agent conversation for an externally executed task.

## Terms

- **Worker run:** One completed external provider invocation, including failure.
- **Review iteration:** One completed, validated verdict, approved or rejected.
  A transport retry or failed reviewer launch is not a review iteration.
- **QA run:** One persisted QA result, passed or failed.
- **Source instance:** A stable identity for one coordinator ledger.
- **Known tokens:** Recorded measurements; missing measurements remain unknown.

## Requirements

### REQ-COSTS-EXTERNAL-001: Authorized immutable activity ingestion

#### Acceptance criteria

- **AC-COSTS-EXTERNAL-001.1:** A caller with permission to update a task shall be
  able to record a worker run, review verdict, or QA result for that task without
  creating a session, starting an agent, or changing task state.
- **AC-COSTS-EXTERNAL-001.2:** An identical event retried after a timeout or restart
  shall be acknowledged without increasing any total. Reuse of its identity with
  changed content or a different task shall be rejected as a conflict.
- **AC-COSTS-EXTERNAL-001.3:** Missing tasks and inaccessible tasks shall not leak
  existence or statistics. Invalid events and forbidden writes shall produce no
  partial accounting effects.
- **AC-COSTS-EXTERNAL-001.4:** Accepted events shall survive restart and retain
  source, event identity, occurrence time, activity kind, outcome, and known
  provider/model attribution. Only task deletion may remove that task's events.

### REQ-COSTS-EXTERNAL-002: Honest token and iteration accounting

#### Acceptance criteria

- **AC-COSTS-EXTERNAL-002.1:** Missing input, output, cached-input, and reasoning
  measurements shall remain distinguishable from measured zero. A partial total
  shall show its known contribution and number of incomplete worker runs.
- **AC-COSTS-EXTERNAL-002.2:** Cached input and reasoning output reported as
  subsets shall not be added twice. Negative, fractional, overflowing, or
  impossible subset counts shall be rejected.
- **AC-COSTS-EXTERNAL-002.3:** A worker run that failed, timed out, retried, or
  repaired a candidate shall contribute its reported usage once. Completed
  review verdicts and QA results shall each count once independently of outcome;
  repeated reviews of the same commit shall remain distinct iterations.
- **AC-COSTS-EXTERNAL-002.4:** Native turns and sessions shall retain their existing
  meaning. External runs shall not increase native turn/session counts or create
  fabricated API charges for subscription usage.

### REQ-COSTS-EXTERNAL-003: Daily and task statistics in the existing UI

#### Acceptance criteria

- **AC-COSTS-EXTERNAL-003.1:** Stats shall show daily tokens by UTC occurrence date,
  task totals, provider/model breakdowns, worker run counts, review iterations,
  and QA runs for the selected workspace and time range. A run crossing midnight
  shall be attributed to its completion date.
- **AC-COSTS-EXTERNAL-003.2:** Operators shall be able to select all, native, or
  external usage. Task identity and links shall remain available for tasks with
  no native sessions, including archived tasks subject to existing access rules.
- **AC-COSTS-EXTERNAL-003.3:** Desktop and phone users shall reach the same
  information. Missing data, no activity, loading, and fetch failures shall be
  distinct; a failure shall preserve known results and offer retry.
- **AC-COSTS-EXTERNAL-003.4:** Backfilled events shall appear on their occurrence
  dates rather than import dates. Reload shall show the same totals. A bounded
  or paginated result shall disclose truncation and allow navigation.

### REQ-COSTS-EXTERNAL-004: Recoverable coordinator delivery

#### Acceptance criteria

- **AC-COSTS-EXTERNAL-004.1:** A coordinator shall be able to replay historical
  ledger entries and send new entries with stable identities. A dashboard outage
  shall not discard local accounting or change delivery/acceptance state.
- **AC-COSTS-EXTERNAL-004.2:** Repeated sync, process restart, and an uncertain HTTP
  response shall not double count events. A conflict shall remain visible to the
  operator and shall not be overwritten or silently acknowledged.
- **AC-COSTS-EXTERNAL-004.3:** Unsupported older dashboards shall continue to show
  existing task descriptions and shall not block coding, QA, review, or approval.

## Out of scope

- Launching native agents, Office, or another dashboard/framework.
- Changing merge permissions, QA gates, human approval, or production deployment.
- Reconstructing measurements or run roles absent from historical evidence.
- Invoices, subscription-to-dollar conversion, or automatic pricing of external runs.
- Arbitrary log, prompt, transcript, credential, or artifact uploads.

## Implementation plans

- [External worker statistics](../../../plans/external-worker-statistics/plan.md)
