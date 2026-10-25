---
id: "04-kodi-delivery"
title: "Kodi ledger delivery"
status: pending
wave: 4
depends_on: ['03-stats-ui']
plan: "plan.md"
requirements:
  - REQ-COSTS-EXTERNAL-001
  - REQ-COSTS-EXTERNAL-002
  - REQ-COSTS-EXTERNAL-004
acceptance_criteria:
  - AC-COSTS-EXTERNAL-001.2
  - AC-COSTS-EXTERNAL-002.1
  - AC-COSTS-EXTERNAL-002.2
  - AC-COSTS-EXTERNAL-002.3
  - AC-COSTS-EXTERNAL-004.1
  - AC-COSTS-EXTERNAL-004.2
  - AC-COSTS-EXTERNAL-004.3
system_design:
  - ../../specs/costs/system-design/external-worker-statistics.md
---

# Task 04: Kodi ledger delivery

## Summary

In the separate Kodi repository, connect existing immutable usage and evidence rows to the external activity API. Use a bounded durable outbox so dashboard errors cannot affect delivery state.

## In scope

- A private persistent ledger-instance UUID, stable event IDs and acknowledged delivery state.
- Explicit new run phase/outcome metadata, conservative provider normalization, unknown historical metadata.
- Bounded historical replay, retries, quarantined conflicts and old-server cooldown/fallback.

## Out of scope

- Kandev production changes, release/image repinning, enabling native agents, altering acceptance gates.

## Acceptance

- Crash after server commit but before local acknowledgment replays once without doubled totals.
- Dashboard failure/conflict cannot block coding, QA, review or approval, and leaves accounting recoverable.
- Provider fixtures preserve cached/reasoning subsets and unknown counts; local reports/descriptions remain compatible.

## Verification

Run from the Kandev repository root unless stated otherwise. Setup downloads
belong only in the image-build phase; independent QA has networking disabled.

```bash
# Run from the separate Kodi development repository:
python3 -m unittest discover -s tests -v
python3 scripts/check_public_source.py
git diff --check
```

## Files likely touched

- `Kodi: kodi/board.py`
- `Kodi: kodi/store.py`
- `Kodi: kodi/usage.py`
- `Kodi: kodi/coordinator.py`
- `Kodi: kodi/runner.py`
- `Kodi: kodi/recovery.py`
- `Kodi: tests/test_external_usage.py`
- `Kodi: docs/operations.md`

## Dependencies

Prior work orders: ['03-stats-ui'].

## Risks

Kodi has its own instructions and acceptance workflow. Do not put organization IDs, private paths, credentials, raw provider output or operator deployment settings in public source.

## Parallelism

`sequential`

## Inputs

- [Requirements](../../specs/costs/requirements/external-worker-statistics.md)
- [System design](../../specs/costs/system-design/external-worker-statistics.md)
- [Full plan](plan.md)

## Results

Pending.
