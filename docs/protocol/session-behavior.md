# Session Behavior

This document records observed TCP session behavior for AMT 8000 Pro local
network access over ISECNet v2.

## Current App Behavior

The web app does not keep a persistent panel session. Each status refresh,
event download, or event export opens a TCP connection, authenticates, runs the
needed command sequence, sends disconnect, and closes the connection.

Commands for the same panel key (`host:port`) are serialized in the backend.
This prevents concurrent authenticated sessions or overlapping command streams
against one central.

## Verified Panel Behavior

Two diagnostics were run against a real AMT 8000 Pro using local `.env`
connection settings. Secrets and payloads were not printed.

1. Two quick overlapping status-capture processes both succeeded.
2. A stronger held-session test authenticated session A, read status, kept A
   open, then attempted to authenticate session B while A was still open.
   Session B connected but was closed by the panel during authentication.
3. After session A disconnected, a normal status read succeeded again.

Conclusion: the panel appears to tolerate fast overlapping TCP attempts when the
first authenticated session exits quickly, but it behaves as a single active
authenticated remote session device. A second authenticated session should not
be assumed safe while the first is open.

## Official App Behavior

Captured Android AMT Remoto Mobile sessions show the official app keeping a
single TCP connection open:

- Event download used the same authenticated session.
- Event records were requested with `0x3900` batches.
- Batches were paced roughly every `300-500ms`.
- After event download, the app sent `0x0b4a` keepalives roughly every `10s`.

## Design Implications

The current per-panel backend lock is required even with short-lived
connections.

If persistent sessions are implemented later, the safe shape is:

- one backend TCP session per `host:port`
- one serialized command queue per session
- keepalive using observed `0x0b4a` behavior
- idle timeout and explicit close on logout
- reconnect only after confirmed session failure, with backoff
- no automatic event polling

A browser WebSocket can be useful for progress and live state updates, but the
core safety requirement is the backend singleton session and command queue, not
the browser transport.
