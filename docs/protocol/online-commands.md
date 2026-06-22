# Online Commands

This document records the current implementation boundary for AMT 8000 Pro
Online-tab control commands.

## Current State

The web app has guarded API endpoints and append-only audit logging for the M1
Online command surface. Partition `0` arm/disarm and temporary zone
bypass/un-bypass are implemented from captured AMT 8000 Pro traffic. Other
mutating ISECNet methods return an unsupported error until a sanitized capture
proves the command, payload, response, and read-after-write behavior.

This is intentional. No command bytes in this file are inferred from AMT Remoto
manual text or from other alarm models.

## Guarded API Surface

All Online command endpoints require an authenticated panel session cookie and a
JSON body containing `{"confirm":true}`.

| Endpoint | Body fields | Current real-client behavior |
| --- | --- | --- |
| `POST /api/online/partitions/{partition}/arm` | `confirm`, optional `mode:"away"` | implemented for partition `0` |
| `POST /api/online/partitions/{partition}/disarm` | `confirm` | implemented for partition `0` |
| `POST /api/online/clock/sync` | `confirm`, optional RFC3339 `time` | unsupported |
| `POST /api/online/zones/{zone}/bypass` | `confirm`, `bypassed:true|false` | implemented for enabled zones |
| `POST /api/online/alarm-memory/clear` | `confirm` | unsupported |
| `POST /api/online/pgms/{pgm}` | `confirm`, `active:true|false` | unsupported |

Command responses use one shape:

```json
{"ok":false,"auditId":"...","error":"..."}
```

Successful future implementations must return:

```json
{"ok":true,"status":{},"auditId":"..."}
```

## Audit Records

Online command attempts are written as JSONL records to
`AMT_AUDIT_PATH` when configured, otherwise to the application config directory
for the local user.

Audit records include:

- timestamp
- panel host and port
- action
- target
- requested state
- result
- error text

Audit records must not include remote passwords, cookies, raw authentication
frames, or installation-sensitive packet captures.

## Evidence Required Before Enabling Commands

For each command, capture and document:

- exact ISECNet command ID
- exact request payload
- exact response command and payload
- response success/error semantics
- status fields used for read-after-write verification
- sanitized example request and response frames
- real-panel test report under `docs/test-runs/`

## Command Evidence Table

| Capability | Command | Request payload | Response payload | Read-after-write status | State |
| --- | --- | --- | --- | --- | --- |
| arm partition 0 away | `0x401e` | `ff01` | `ff91` | status `0x0b4a`: global `ARMED`, partition `0` armed | implemented |
| disarm partition 0 | `0x401e` | `ff00` | `ff90` | status `0x0b4a`: global `DISARMED`, partition `0` not armed | implemented |
| temporary zone bypass/un-bypass | `0x401f` | pairs of `zoneIndex0,bypassFlag` for enabled zones | `0xf0fe` empty payload | status `0x0b4a`: target zone `bypassed` bit | implemented |
| PGM activate/deactivate | unknown | unknown | unknown | unknown until PGM status is proven | blocked |
| panel clock sync | unknown | unknown | unknown | `panelDateTime` from status command `0x0b4a` | blocked |
| zone bypass/un-bypass | unknown | unknown | unknown | zone `bypassed` field from status command `0x0b4a` | blocked |
| clear alarm memory | unknown | unknown | unknown | fired-zone/alarm memory semantics not proven | blocked |

## Capture Workflow

1. Run `amt8000-capture-proxy` on a trusted LAN.
2. Point the official Intelbras app at the proxy.
3. Perform one Online action at a time during a safe test window.
4. Store raw captures outside the repo.
5. Redact and reduce captures into unit-test fixtures before committing.
6. Update this file and implement the corresponding typed client method.
7. Run automated tests and record a real-panel test report.

## Captured Partition Control Evidence

Session `20260622T160538.023946000Z` from the mobile app through the local
proxy showed these command frames:

| UTC time | Operator action | Client command | Client payload | Panel response payload |
| --- | --- | --- | --- | --- |
| `2026-06-22T16:06:23.921856Z` | disarm | `0x401e` | `ff00` | `ff90` |
| `2026-06-22T16:08:34.946768Z` | arm | `0x401e` | `ff01` | `ff91` |
| `2026-06-22T16:08:38.367493Z` | disarm | `0x401e` | `ff00` | `ff90` |

The post-command status sequence used by the mobile app included command
`0x0b01` with payload prefixes `6691...` after arm and `0690...` after disarm.
The LAN console uses the already-proven full status command `0x0b4a` for
read-after-write verification.

## Captured Zone Bypass Evidence

Session `20260622T161750.703591000Z` showed temporary zone bypass commands for
zones `1` and `2`. The command sends the full enabled-zone bypass map as
zero-based zone/index pairs and the panel acknowledges with command `0xf0fe`
and an empty payload.

| UTC time | Operator action | Client command | Client payload | Panel response |
| --- | --- | --- | --- | --- |
| `2026-06-22T16:19:51.415235Z` | bypass zone 1 | `0x401f` | `00010100020003000400` | `0xf0fe` |
| `2026-06-22T16:20:06.469198Z` | clear zone 1 bypass | `0x401f` | `00000100020003000400` | `0xf0fe` |
| `2026-06-22T16:23:10.360978Z` | bypass zone 2 | `0x401f` | `00000101020003000400` | `0xf0fe` |
| `2026-06-22T16:23:22.239772Z` | clear zone 2 bypass | `0x401f` | `00000100020003000400` | `0xf0fe` |

The implemented client reads current status first, preserves every other
enabled zone's current bypass state, changes only the requested zone, sends the
full pair payload, then verifies the target zone's `bypassed` bit through
status command `0x0b4a`.

## Additional Read/Blocked Evidence From The Same Capture

- Zone/sectors status used `0x0b73` and `0x0b74` with request payload `01`.
- PGM 1 attempts used `0x45af` with payload `0001`, but the app reported
  "PGM not accessible"; this remains blocked.
- Connection status used `0x0b71`; observed response payload `0301`.
- Network-card status used `0x3812`, `0x3813`, `0x3814`, `0x3815`, and `0x3faa`.
