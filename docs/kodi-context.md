# Kodi repository context

This fork retains upstream Kandev's AGPL-3.0 license, architecture and repository
instructions. Kodi coordinates external workers; native agent/session startup
must remain disabled in its dashboard deployment.

Run `python3 .kodi/qa.py` inside the dependency image built from `.kodi/Dockerfile`.
The image pins the upstream CI toolchain and verifies Go/pnpm/workspace manifest
identities before using prepared dependencies. This is scoped Go/task/analytics
and Stats/frontend acceptance; real PostgreSQL parity and browser scenarios are
additional evidence for changes crossing those boundaries.

Independent QA, different-model review, GitHub checks and human CODEOWNER review
must refer to the candidate commit before dispatch activation or merge. Keep
operator QA profiles, deployment mounts, credentials and runtime state outside
Git. Do not publish releases or switch the running dashboard as part of coding.

## Source export compatibility

The committed source contains internal instruction/configuration aliases and is
larger than Kodi's default source export. It requires a reviewed, explicitly
bounded materialized export profile; do not bypass extraction checks or omit
source paths to obtain acceptance. The tracked `apps/cli/bin/claude` dependency
launcher was dangling in the fork baseline and is removed during onboarding;
the npm CLI entry points remain defined by `apps/cli/package.json`.

## Confirmed lessons

Record confirmed reusable review findings with source URLs and verification.

### Receipt replay after task moves

[PR #2](https://github.com/OmiroAI/kandev/pull/2) covers a destination namespace
that already contains the same external source-instance/event identity on a
different task. A lookup that rejects any foreign match incorrectly
rejects an identical replay after transfer. Prioritize the existing task binding
and its payload hash; classify the destination collision only when that binding
is absent. The SQLite and PostgreSQL transfer-namespace regressions verify the
original receipt, changed-payload conflicts, and exactly two stored records.


## Fork CI scope and browser assertions

[PR #2](https://github.com/OmiroAI/kandev/pull/2) retains Linux, PostgreSQL and
browser acceptance. Windows CI is disabled by default at the owner's request;
`KANDEV_WINDOWS_CI=true` is an explicit repository-variable opt-in, and the
backend aggregate requires the resulting skip when it is off. Docs preview
publication likewise requires `KANDEV_DOCS_DEPLOY=true`; a fork without the
upstream landing repository must still validate docs without attempting deploy.
Workflow contract tests cover both gates.

The same PR's CI exposed prefix collisions in repository option names. Select
an exact accessible option name rather than the first substring match. Monaco
keyboard tests must confirm the native EditContext holds focus before shortcuts.
For animated mobile controls, wait for settled geometry and compare dimensions
at Chromium's 1/64px layout precision; keep the 44px target and real tap checks.
Verification uses the exact failed desktop and phone scenarios without retries.

Pseudo-locale browser coverage must seed `kandev_locale` on the browser context
before the first navigation. Changing the cookie between an English boot and a
reload races startup locale persistence. PR #2 covers the failed office-routines
screen and all helper call sites with 12 retry-free attempts.

MCP subtask coverage should observe the completed turn and persisted child, not
require an idle composer: workflow automation may already be starting its next
agent. Outside-tap menu tests wait for the opening animation before dismissal,
matching their other menu interactions. Both failed scenarios passed three
retry-free attempts after these changes.

Storage UI tests restore policy through the existing API helper; a second save
can legitimately overlap its first success toast. MCP report tests keep their
HTTP route installed across report-state cases, then remove only their handler.
PR #2's trace showed a reload bypassing a removed/reinstalled per-case route.
Focused cleanup and MCP cases passed three attempts each without retries.


## Recovery and readiness regressions

[PR #2](https://github.com/OmiroAI/kandev/pull/2) also covers five failures from
the fork's Linux CI. A pending turn-history snapshot can outlive a missed final
notification. Refresh pending summaries only after the session settles, merge
without resetting history pagination, and stop when resolved or unmounted.
Fake-clock tests cover the missed event, cursor preservation, active-turn silence
and timer cleanup; the full mobile turn-history spec verifies rendering.

When an SSH execution loses registration to an already-adopted execution, its
rollback must not kill the winning remote agentctl. Match the task/session and
complete SSH process identity before detaching only the losing transport; a
missing owner or different process still gets normal forced rollback. The race
regression, full lifecycle race suite and five real SSH passthrough scenarios
cover interruption, reconnect, sibling input, Stop and Resume.

Guided-setup tests select seeded profile IDs and wait for picker focus return
before opening the next picker. Lifecycle tests assert the related session
states from one observed snapshot. Tablet geometry uses settled dimensions at
1/64px layout precision while preserving the size and touch-target bounds.

## Editing external-worker tickets

[The clarification integration](https://github.com/OmiroAI/kodi/pull/45) exposed
a metadata-only task edit blocked by missing native agent compatibility, even
with a repository selected. Reuse the existing update-only handler when an edit
has no agent profile, and apply the same distinction to form/keyboard guards.
Keep pending uploads, edit dependencies and all native-start restrictions.
Footer and shortcut regressions reproduce the disabled Update path; owning
submission tests verify the existing metadata-save behavior.

Sessionless previews must also show the persisted description and questions,
without requiring a chat session. The preview already receives the task's
description; render that data directly. Keep editing in the existing task edit
dialog: the Office description editor requires its own page context provider.
Keep the previous empty state for tasks without descriptions and retain recovery
feedback. The disposable preview pilot verifies reading and saving actual text.
