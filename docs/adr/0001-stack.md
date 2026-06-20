# ADR 0001: Go Single-Binary LAN Service

## Decision

Use Go for the first implementation.

## Context

The service needs raw TCP access to the AMT 8000 Pro local protocol, a small LAN
web UI, and easy deployment on Mac, Linux, or Windows without a heavy runtime.

## Consequences

- The first version uses `net/http` and standard-library templates.
- Protocol code stays in `internal/isecnet`.
- Local persistence can be added with SQLite after the status MVP.
- UI complexity stays modest until protocol coverage is proven.

