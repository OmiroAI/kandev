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
