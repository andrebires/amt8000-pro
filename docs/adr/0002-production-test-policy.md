# ADR 0002: Real Panel Test Policy

## Decision

Every panel feature must be tested against the local AMT 8000 Pro before it is
marked complete.

## Context

The ISECNet v2 protocol is proprietary and community documentation is partial.
Unit tests prove our local code, but they do not prove compatibility with the
installed panel firmware and configuration.

## Consequences

- `TASKS.md` tracks real-panel validation separately from implementation.
- Read-only features can be tested directly.
- Control features require an explicit manual safety checklist.
- Configuration writes require read-after-write validation and recorded evidence.
- Test logs go under `docs/test-runs/` and must not include secrets.

