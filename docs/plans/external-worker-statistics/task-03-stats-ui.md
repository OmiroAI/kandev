---
id: "03-stats-ui"
title: "Existing Stats UI"
status: in_progress
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
(cd apps/web && pnpm test app/stats lib/api/domains/stats-api.test.ts hooks/domains/stats/use-usage-stats.test.ts)
(cd apps/web && pnpm run typecheck)
(cd apps/web && pnpm run i18n:check)
(cd apps/web && pnpm e2e:run --project chromium e2e/tests/stats/external-worker-statistics.spec.ts --retries=0)
(cd apps/web && pnpm e2e:run --project mobile-chrome e2e/tests/stats/mobile-external-worker-statistics.spec.ts --retries=0)
git diff --check
```

## Files likely touched

- `apps/web/app/stats`
- `apps/web/lib/api/domains/stats-api.ts`
- `apps/web/lib/types/http`
- `apps/web/src/locales`
- `apps/web/e2e/tests/stats/external-worker-statistics.spec.ts`
- `apps/web/e2e/tests/stats/mobile-external-worker-statistics.spec.ts`

## Dependencies

Prior work order: 02-activity-api. Its local implementation, full task/analytics
regressions, SQLite/PostgreSQL race/upgrade checks, lint and hooks are complete
in PR #2. Independent acceptance and fresh different-model review remain pending. Reversible local UI work proceeds on this stacked
branch without enabling repository dispatch or changing merge gates.

## Risks

The existing E2E runner supports `--project=chromium` and file arguments. Dependency installation is setup only; prehydrate the independent offline QA image instead of downloading during QA.

## Parallelism

`sequential`

## Inputs

- [Requirements](../../specs/costs/requirements/external-worker-statistics.md)
- [System design](../../specs/costs/system-design/external-worker-statistics.md)
- [Full plan](plan.md)

## Results

Implemented the usage section in existing Stats, using the shared range control
and a dedicated all/native/external source selector. It includes known/complete
token totals, incomplete observations, worker/review/QA/native record counts,
reported worker time, a bounded UTC chart with a daily values table, independent
provider/task paging, desktop rows and phone cards. Copy resolves at render in
all required locales; Traditional Chinese and pseudo were generated with the
repository tools.

API-client, request-ownership and daily-series regressions passed with the
owning Stats suite (47 tests). Full TypeScript, locale parity and scoped lint
passed. The browser RED gate observed zero displayed versus 120 known tokens
after a real committed POST with a deliberately dropped response and an
identical replay. Desktop and phone runs use the managed runner, production
assets, one worker, no retries, and an offline pinned toolchain image.

The desktop scenario covers UTC dates, partial counts, review/QA iterations,
keyboard source selection, 50-row task pagination, refresh failure and retry.
The phone scenario verifies cards, a 44px touch target, an actual hit test and
no horizontal overflow at 320px. Legacy Stats recovery scenarios are included;
their injected outages now remain active until the explicit retry step so a
foreground refresh cannot erase the assertion window.

Independent host acceptance, combined-PR model review, software CI and human
approval remain required. No dashboard/runtime deployment or dispatch activation
is part of this change.


### CI follow-up

Owner-requested fork scope makes Windows validation opt-in. Linux/static,
PostgreSQL and browser gates remain required. Docs previews are opt-in so docs
validation does not depend on the upstream publisher repository. CI remediation
also names the external-source constant for full-branch Go lint, selects exact
repository option names, confirms Monaco keyboard focus, and measures settled
mobile targets at browser layout precision. Affected workflow contracts and exact
failed browser scenarios are the focused verification; remote acceptance of the
remediation commit remains pending until its checks finish.


Local follow-up verification passed: 26 backend workflow contract tests, three
docs workflow tests, analytics `UsageStats` repository tests under `-race`, scoped
Go lint with zero findings, and ESLint/spec checks. The two failed desktop
scenarios each passed three consecutive attempts, and the failed mobile scenario
passed three consecutive attempts, all with retries disabled on the fresh build.

A subsequent full CI run also exposed the native-source lint constant and browser
setup races for pseudo-locale boot, mobile outside-tap dismissal, and a subtask
check that assumed workflow automation leaves the composer idle. Their targeted
fixes passed 12 pseudo-locale attempts, three menu-gesture attempts and three
subtask attempts without retries; the subtask run used the fresh backend build.
Final remote CI and independent acceptance remain pending on the resulting head.
