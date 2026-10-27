---
id: "01-fork-onboarding"
title: "Audited fork onboarding"
status: blocked
wave: 1
depends_on: []
plan: "plan.md"
requirements:
  - REQ-COSTS-EXTERNAL-001
  - REQ-COSTS-EXTERNAL-003
acceptance_criteria:
  - AC-COSTS-EXTERNAL-001.1
  - AC-COSTS-EXTERNAL-003.1
system_design:
  - ../../specs/costs/system-design/external-worker-statistics.md
---

# Task 01: Audited fork onboarding

## Summary

Prepare a real offline Go/web QA entry point and an auditable fork CI baseline. Keep dispatch disabled until its reviewed onboarding is accepted.

## In scope

- Replace the unmerged inventory scaffold; it is not valid Go/web QA.
- Pin toolchains, Go/web locks and browser assets in an offline image; use private operator profile storage for deployment-specific image settings.
- Audit upstream Actions triggers and publishing permissions before enabling only the intended CI; retain AGPL notices and owner review.

## Out of scope

- Feature implementation, production rollout, or fabricated acceptance.
- Broad App grants or automatic merge enablement.

## Acceptance

- The same committed QA entry point passes independently offline with exact dependency locks.
- Required checks and human CODEOWNER review are configured and verified before dispatch activation.
- The catalog retains auto-merge off; runtime activation occurs only after actual acceptance evidence.

## Verification

Run from the Kandev repository root unless stated otherwise. Setup downloads
belong only in the image-build phase; independent QA has networking disabled.

```bash
python3 .kodi/qa.py
python3 scripts/list-docs.py validate
python3 scripts/lint-spec-files.py --all
git diff --check
```

## Files likely touched

- `.kodi/qa.py`
- `.kodi/config.json`
- `.github/workflows/kodi-qa.yml`
- `.github/CODEOWNERS`
- `docs/kodi-context.md`

## Dependencies

None.

## Risks

The generic onboarding helper has no Go profile. Do not use a syntax-only or inventory baseline as acceptance for Go/web behavior. Operator image setup and host protection settings are outside model-worker authority.

## Parallelism

`sequential`

## Inputs

- [Requirements](../../specs/costs/requirements/external-worker-statistics.md)
- [System design](../../specs/costs/system-design/external-worker-statistics.md)
- [Full plan](plan.md)

## Results

The onboarding source and local acceptance entry point are prepared. Local
`python3 .kodi/qa.py` passes with networking disabled, a read-only image, one CPU,
8192 MiB memory/swap ceiling, 128 PIDs, and the existing non-executable `/tmp`.
The run used the pinned toolchain/dependency image plus a read-only prepared
compiler cache for the identical Go source; `-count=1` forces actual test runs.
All scoped Go checks, 34 Stats frontend tests, full web typecheck, locale checks,
and the three dependency-identity regressions passed. Specification catalog,
structure lint and pre-commit preflight also pass.

The current source export is 290,816,000 bytes before onboarding changes and
contains internal aliases. The proposed materialized export validates all source
paths without omissions; the dangling generated dependency launcher is removed.
The QA image verifies locks/manifests, warms the `fts5` compiler cache, constrains
Go heap growth, and provides an explicit 6144 MiB Node heap.

Blocked on the reviewed transport prerequisite in
[Kodi PR #5](https://github.com/OmiroAI/kodi/pull/5), operator installation of the
helper and dedicated bounded profile, then exact-commit independent QA,
different-model review, GitHub checks and human CODEOWNER approval for this
onboarding. The private profile and host capacity must be reviewed before running
a larger job; no live QA helper/profile or deployment has been changed.
Repository dispatch and auto-merge remain disabled. Work orders 02 through 04
remain pending this prerequisite.
