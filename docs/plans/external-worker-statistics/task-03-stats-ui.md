---
id: "03-stats-ui"
title: "Existing Stats UI"
status: pending
wave: 3
depends_on: ['02-activity-api']
plan: "plan.md"
requirements:
  - REQ-COSTS-EXTERNAL-002
  - REQ-COSTS-EXTERNAL-003
acceptance_criteria:
  - AC-COSTS-EXTERNAL-002.1
  - AC-COSTS-EXTERNAL-002.4
  - AC-COSTS-EXTERNAL-003.1
  - AC-COSTS-EXTERNAL-003.2
  - AC-COSTS-EXTERNAL-003.3
  - AC-COSTS-EXTERNAL-003.4
system_design:
  - ../../specs/costs/system-design/external-worker-statistics.md
---

# Task 03: Existing Stats UI

## Summary

Add external usage to the existing Stats UI with daily and task views and native/external source filtering. Keep all states responsive, accessible and localized.

## In scope

- Stats API/types and rendered usage section, source/range state, provider/model detail and task pagination.
- Desktop table and phone cards; known subtotal/incomplete count, empty/loading/error/retry states.
- Locale parity and actual HTTP/browser replay scenario.

## Out of scope

- Another dashboard, agent/session startup, pricing UI or Kodi sync implementation.

## Acceptance

- UI-01 through UI-03 show the required task/daily/provider data with honest partial totals.
- Desktop and phone expose identical information and reachable controls without overflow.
- Browser reload after duplicate ingestion preserves totals; error/retry retains prior data.

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

See the [full plan](plan.md#ascii-ui-preview).

## Verification

Run from the Kandev repository root unless stated otherwise. Setup downloads
belong only in the image-build phase; independent QA has networking disabled.

```bash
(cd apps && pnpm install --frozen-lockfile)
(cd apps/web && pnpm test app/stats lib/api/domains/stats-api.test.ts)
(cd apps/web && pnpm run typecheck)
(cd apps/web && pnpm run i18n:check)
(cd apps/web && pnpm e2e:run e2e/tests/external-worker-statistics.spec.ts --project=chromium)
git diff --check
```

## Files likely touched

- `apps/web/app/stats`
- `apps/web/lib/api/domains/stats-api.ts`
- `apps/web/lib/types/http`
- `apps/web/src/locales`
- `apps/web/e2e/tests/external-worker-statistics.spec.ts`

## Dependencies

Prior work orders: ['02-activity-api'].

## Risks

The existing E2E runner supports `--project=chromium` and file arguments. Dependency installation is setup only; prehydrate the independent offline QA image instead of downloading during QA.

## Parallelism

`sequential`

## Inputs

- [Requirements](../../specs/costs/requirements/external-worker-statistics.md)
- [System design](../../specs/costs/system-design/external-worker-statistics.md)
- [Full plan](plan.md)

## Results

Pending.
