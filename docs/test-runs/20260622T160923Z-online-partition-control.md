# Online Partition Control Capture - 2026-06-22

## Scope

Captured AMT 8000 Pro mobile-app traffic through `amt8000-capture-proxy` while
operating partition `0` arm/disarm from the official app.

## Environment

- Panel: AMT 8000 Pro on local TCP `9009`; IP omitted
- Proxy listen: `0.0.0.0:9009`
- Mobile client observed by proxy; IP omitted
- Capture session: `20260622T160538.023946000Z`

## Observed Commands

| UTC time | Action | Command | Request payload | Response payload | Result |
| --- | --- | --- | --- | --- | --- |
| `2026-06-22T16:06:23.921856Z` | disarm | `0x401e` | `ff00` | `ff90` | panel disarmed |
| `2026-06-22T16:08:34.946768Z` | arm | `0x401e` | `ff01` | `ff91` | panel armed |
| `2026-06-22T16:08:38.367493Z` | disarm | `0x401e` | `ff00` | `ff90` | panel disarmed |

## Verification

- The mobile app issued status reads after each command.
- Status evidence after arm included payload prefix `6691...` in the mobile
  status command stream.
- Status evidence after disarm included payload prefix `0690...` in the mobile
  status command stream.
- The LAN console implementation verifies command effects with the existing
  full status command `0x0b4a`.

## Sanitization

- Remote password and authentication frames are not recorded here.
- Raw capture file remains outside the repository.
- This report stores only command IDs, command payloads, response payloads,
  timestamps, and high-level observed result.
