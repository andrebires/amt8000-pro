# Status Command

This document records the current AMT 8000 Pro read-only status evidence used by
the LAN console. It is implementation evidence, not official Intelbras
documentation.

## Command

- Command: `0x0b4a`
- Transport: ISECNet v2 over TCP `9009`
- Session flow: connect, authenticate with remote password, send status command,
  parse one status response frame, send disconnect command.

## Response Payload

The current parser requires at least `143` payload bytes.

Known offsets:

| Offset | Length | Meaning |
| --- | ---: | --- |
| `0` | 1 | model byte |
| `1..3` | 3 | firmware version as major, minor, patch |
| `12..19` | 8 | enabled zone bitmap |
| `20` | 1 | global status bits |
| `21..36` | 16 | partition records |
| `38..45` | 8 | open zone bitmap |
| `46..53` | 8 | fired/violated zone bitmap |
| `54..61` | 8 | bypassed zone bitmap |
| `64..69` | 6 | panel local date/time as BCD: day, month, year, hour, minute, second |
| `71` | 1 | panel tamper bit `0x02` |
| `89..96` | 8 | zone tamper bitmap |
| `105..112` | 8 | zone low-battery bitmap |
| `134` | 1 | panel battery level enum |

Global status byte `20`:

| Bits | Meaning |
| --- | --- |
| `0x02` | siren live |
| `0x04` | zones closed |
| `0x08` | zones firing |
| `(byte >> 5) & 0x03` | global state enum: `0` disarmed, `1` partial, `3` armed |

Partition byte:

| Bit | Meaning |
| --- | --- |
| `0x80` | partition enabled |
| `0x40` | stay mode |
| `0x08` | fired |
| `0x04` | firing |
| `0x01` | armed |

Panel battery enum:

| Value | Meaning |
| ---: | --- |
| `1` | dead |
| `2` | low |
| `3` | middle |
| `4` | full |

## Derived Read Model

- Zone state is derived from open and fired bits:
  - `CLOSED`: not open, not fired
  - `OPEN`: open, not fired
  - `FIRED_CLOSED`: fired, not open
  - `FIRED_OPEN`: fired and open
- Partition state is derived from partition bits, in priority order:
  `FIRING`, `FIRED`, `STAY`, `ARMED`, `DISARMED`.
- Panel date/time is parsed from payload bytes `64..69` as a timezone-free
  local timestamp in the format `YYYY-MM-DDTHH:MM:SS`. The first real-panel
  fixture contained `20 06 26 19 15 21`, matching `2026-06-20T19:15:21`.
- Pending problems are derived only from known bits:
  panel tamper, low/dead panel battery, zone tamper, and zone low battery.

## Unsupported Until More Evidence

These Online-tab fields are intentionally not parsed yet because the offsets or
commands are not proven for AMT 8000 Pro:

- source voltage
- battery voltage
- richer pending problem codes
- PGM list/status
- alarm-memory clear state

## Zone Names

Zone names are read separately from the full status command.

- Command: `0x33e0`
- Request payload: 16 zero-based zone indexes.
- Response payload: 16 records, 15 bytes each.
- Record format: 1 byte zero-based zone index followed by a 14-byte
  space-padded text label.
- Current implementation reads four batches: `0..15`, `16..31`, `32..47`,
  and `48..63`.

Sanitized real-panel fixtures are stored under `docs/fixtures/status/`.
