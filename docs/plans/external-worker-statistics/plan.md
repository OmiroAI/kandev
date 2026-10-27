---
created: 2026-10-25
status: draft
requirements:
  - REQ-COSTS-EXTERNAL-001
  - REQ-COSTS-EXTERNAL-002
  - REQ-COSTS-EXTERNAL-003
  - REQ-COSTS-EXTERNAL-004
system_design:
  - ../../specs/costs/system-design/external-worker-statistics.md
legacy_specs: []
---

# Implementation Plan: External worker statistics

## Overview

Add external activity ingestion to the fork, expose it in existing Stats views,
then connect Kodi's durable ledger. Complete real Go/web onboarding before any
repository dispatch. No agent/session startup, deployment, or merge-policy change
is required to collect the statistics.

## Intent and assumptions

Confirmed: fork Kandev under the organization, add it to Kodi, and support daily
usage, task usage and review iteration tracking. Verified: upstream has a native
usage ledger and read-only Stats endpoints; external tasks lack native sessions.
Routine design choices: UTC completion-date grouping, stable event identities,
explicit external provenance, and no guessed historical measurements.

Ownership: costs owns the accounting/projection contract, with task persistence
and access used as dependencies. This is distinct from conversation usage, which
owns native response/turn/session observations and chat display.

## Scope

### In scope

- Audited repository onboarding and a suitable offline Go/web QA profile.
- Bounded authorized immutable ingestion and idempotent replay.
- Daily, task, provider/model, review and QA statistics inside existing Stats.
- A recoverable Kodi adapter, historical backfill and old-dashboard compatibility.

### Out of scope

- Automatic native agent startup, Office or another dashboard.
- Model billing, invented counts, merge authority or production rollout.
- Changes to upstream license or release publishing.

## Technical approach

1. Replace the generic generated inventory QA scaffold with a real Go/web entry
   point, independent pinned toolchain/dependency image, CODEOWNERS and scoped CI.
   Inventory QA is not suitable for this source repository. Keep catalog disabled
   and auto-merge off until onboarding acceptance and human approval.
2. Add the external activity table, task-scoped write contract and workspace Stats
   read projection using existing task-write and workspace-read authorization.
   Preserve native ledger APIs; aggregate facts before metadata joins.
3. Extend the existing Stats client/section with source filters, honest partial
   totals, responsive task accounting and localized state handling.
4. Connect Kodi's immutable usage/evidence rows through a durable bounded outbox.
   Bind IDs to one persistent ledger instance; record new run metadata explicitly.
   Preserve the prior local CLI reports and task descriptions as fallback.

### Producer compatibility

| Producer | Identity/measurement | Intended behavior | Evidence | Unsupported fallback |
| --- | --- | --- | --- | --- |
| Native Kandev | Existing ledger event IDs and disjoint counts | Existing recording; read-only combined Stats projection | Native ledger regressions | Preserve original API meaning |
| Kodi Codex | Ledger instance + usage row; inclusive input/output | Normalize reported components once | Adapter fixtures with cached subsets | Unknown component |
| Kodi Gemini/Antigravity | Same row identity; provider-specific envelope | Normalize only verified reported shapes | Adapter fixtures with missing fields | Unknown component |
| Kodi local reviewer | Same row identity; prompt/completion counts | Known zeros retained | Local reviewer fixture | Unknown component |
| Historical Kodi evidence | Evidence row ID, kind, timestamp | Review/QA count once, including same-SHA revalidation | Replay/restart tests | Do not invent worker role |
| Older Kandev | Missing write route | Leave events pending, retain descriptions | 404/cooldown compatibility test | No delivery-state change |

## ASCII UI preview

### UI-01: Stats usage, desktop, populated/partial

Entry: existing Stats navigation, selected workspace.

```text
Stats                          [Week | Month | All]
Usage source [All v]            Time zone: UTC
+---------------------+-----------------------------+
| Tokens: 1.2M known   | Worker runs: 18              |
| 2 runs incomplete   | Reviews: 7    QA runs: 9     |
+---------------------+-----------------------------+
Daily tokens   [chart + accessible daily summary]
Provider / model          Known tokens   Incomplete
Codex / reported model       1.2M              2
Task              Tokens    Runs   Reviews   QA
Example task      45K          3        2      2
[Next task page]
```

### UI-02: Stats usage, phone, same data

```text
Stats          [Range v]
[Usage source: All v]
UTC | Tokens: 1.2M known
2 runs incomplete
Runs 18 | Reviews 7 | QA 9
[Daily chart + text summary]
[Provider/model disclosure]
+-------------------------+
| Example task (link)     |
| Tokens 45K | Runs 3     |
| Reviews 2  | QA 2       |
+-------------------------+
[Next task page]
```

### UI-03: Usage section states, shared composition

```text
Loading:     Usage [progress indicator]
Empty:       No recorded usage in this range
Unavailable: Usage unavailable [Retry]
Refresh error with data:
             Tokens 45K (previous data) [Retry]
Partial:     45K known + 2 runs with missing usage
```

Source/range controls, task links, incomplete measurements, separate native-turn
labels, retry, and reachable pagination are structural requirements. Exact
numbers, spacing, and chart styling are illustrative. The desktop table becomes
stacked cards on phones; controls remain touch accessible and keyboard labelled.
All copy is localized. These views implement AC-COSTS-EXTERNAL-002.1,
AC-COSTS-EXTERNAL-002.4 and AC-COSTS-EXTERNAL-003.1 through 003.4.

## Tests

| Criteria | Planned evidence |
| --- | --- |
| AC-COSTS-EXTERNAL-001.1..001.4 | Task handler/service/repository external activity tests: read-only/forbidden/other-workspace writes, duplicate race, changed-task conflicts, immutable persistence and rollback |
| AC-COSTS-EXTERNAL-002.1..002.4 | External aggregate tests: unknown/zero/subsets/overflow, failed runs, repeated same-SHA verdicts, native counters unchanged |
| AC-COSTS-EXTERNAL-003.1..003.4 | Stats aggregate/API tests: offsets, backfill, source/range isolation, pagination, no-session task; rendered Stats states |
| AC-COSTS-EXTERNAL-004.1..004.3 | Kodi outbox/adapter tests: dropped response, restart, conflict, bounded replay, missing endpoint |

## E2E tests

Add `apps/web/e2e/tests/external-worker-statistics.spec.ts` to the existing Chromium
project: ingest externally executed work without a session, replay a request,
reload Stats, change range/source, and open the task. Exercise UI-01 through
UI-03 and a phone viewport in that same file. Cover 003.1..003.4 and 001.1..001.2.
Do not require provider accounts, Docker executor launch, or external network.

## Work orders

- [ ] [Task 01: Audited fork onboarding](task-01-fork-onboarding.md)
- [ ] [Task 02: External activity API and projections](task-02-activity-api.md)
- [ ] [Task 03: Existing Stats UI](task-03-stats-ui.md)
- [ ] [Task 04: Kodi ledger delivery](task-04-kodi-delivery.md)

Order: 01 -> 02 -> 03 -> 04. All are sequential; this package authorizes no
subagents. Kodi changes live in its separate development clone/task branch and
require their own accepted PR. Fork onboarding and application rollout are
separate from feature acceptance.

## Verification results

Onboarding source and its complete scoped offline entry point pass locally.
Task 01 records the exact scope and pending independent/operator acceptance.
Tasks 02 through 04 are not implemented. No dispatch or deployment readiness is
claimed by this package.

## Risks

- The default Python QA environment cannot test Go 1.26 or Node 24. An operator
  must prepare and audit the pinned offline dependency/browser image.
- Forked upstream workflows include scheduled releases and external integrations.
  Audit them before enabling selected CI; do not blanket-enable publishing jobs.
- Fork main is newer than the deployed 0.97.0 image. Record the reviewed source
  SHA and separately plan an operator-tested upgrade; do not silently repin.
- Native aggregates and external nullable measurements have different token
  semantics. Keep their conversion explicit and test partial totals.
- Historical Kodi rows do not contain every run role/outcome. Mark them unknown.
- Dispatch activation needs merged QA, exact-commit independent QA,
  different-model review, CI and human approval. This package is not that evidence.
