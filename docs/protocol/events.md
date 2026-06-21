# Events Command

This document records the current implementation boundary for the AMT 8000 Pro
event buffer.

## Current State

- Web endpoint: `GET /api/events`
- Export endpoint: `GET /api/events/export?format=csv|json`
- Limit: latest `256` events
- UI filters: free text query, partition, and delivery status.
- API/export filters also accept receptor-IP blocked state for diagnostics, but
  the web UI hides it until blocked/disabled behavior is proven.
- Real panel client behavior: reads event records with ISECNet command `0x3900`
  in 16-index batches, parses index/timestamp/raw code bytes, maps observed
  event descriptions, sorts by timestamp, and returns the latest 256 records.

## Command

- Command: `0x3900`
- Request payload: 16 big-endian event indexes, 2 bytes each.
- Response payload: 16 event records, 15 bytes each.
- Observed buffer size: at least indexes `0..511`.
- First implementation reads indexes `511..0`, filters invalid timestamps, sorts
  by timestamp descending, and returns the latest `256`.

Example captured request payload for indexes `21..6`:

```text
001500140013001200110010000f000e000d000c000b000a0009000800070006
```

Example captured response record:

```text
01d52605170719231133fffaa2aa00
```

Parsed as:

- index: `469`
- timestamp: `2026-05-17T07:19:23`
- code bytes: `0x1133`
- raw tail: `ff fa a2 aa 00`

## Normalized Event Model

The API and UI expose these fields:

| Field | Meaning |
| --- | --- |
| `index` | Event position in the latest downloaded buffer |
| `timestamp` | Panel-local timestamp, when present |
| `code` | Raw two-byte event/reporting code, formatted as hex |
| `description` | Human-readable event description, when the code has been mapped |
| `partition` | Partition number, when present |
| `zone` | Zone number, when present |
| `user` | User number, when present |
| `deliveryStatus` | Receptor delivery state such as `sent`, `pending`, `failed`, `blocked`, or `disabled` |
| `receptorIpBlocked` | Event is blocked for receptor-IP sending |
| `receptorIpDisabled` | Event type is disabled for receptor-IP sending |
| `raw` | Raw event bytes/text retained for parser validation |

## Remaining Evidence Needed

- Expand the description table as new event families are observed.
- Confirm whether bytes `10..14` include delivery status and receptor-IP
  blocked/disabled flags for reporting-specific events.
- Investigate observed "Reset date and time" events during some panel
  connections before adding any write-like behavior.

## Observed Description Mapping

These mappings were correlated from Android AMT Remoto Mobile screenshots and
the raw event records captured through the local proxy.

| Code | Parameter bytes | Description |
| --- | --- | --- |
| `0x1133` | `ff fa aN aa 00` | `24h zone alarm - Zone NN` |
| `0x3133` | `ff fa aN aa 00` | `Restoration zone alarm 24 hours - Zone NN` |
| `0x1145` | `ff fa aa aa 00` | `Tamper violation - Panel` |
| `0x3145` | `ff fa aa aa 00` | `Tamper restoration - Panel` |
| `0x1354` | `ff fa aa aa 00` | `Failure to report events` |
| `0x13a1` | `ff fa aa aa 00` | `Power grid failure` |
| `0x33a1` | `ff fa aa aa 00` | `Restoration power grid failure` |
| `0x13a6` | `ff f1 99 aa 00` | `Programming changed - User 99` |
| `0x1461` | `ff fa aa aa 00` | `Incorrect password event` |
| `0x14a1` | `ff f2 a1 aa 00` | `User deactivation - Partition 0 - User 01` |
| `0x34a1` | `ff f2 a1 aa 00` | `User activation - Partition 0 - User 01` |
| `0x14a7` | `ff f1 aa aa 00` | `Deactivation via keyboard or phone - Partition 0 - User Master` |
| `0x34a7` | `ff f1 aa aa 00` | `Activation via keyboard or phone - Partition 0 - User Master` |
| `0x1625` | `ff f2 a2 aa 00` | `Reset date and time` |
| `0x16a2` | `ff fa aa aa 00` | `Periodic test` |

For observed zone/user parameters, `a1`, `a2`, and `a4` map to values `1`,
`2`, and `4`, while BCD-like parameters such as `99` map directly.
