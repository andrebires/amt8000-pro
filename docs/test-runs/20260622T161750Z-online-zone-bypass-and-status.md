# Online Zone Bypass And Status Capture - 2026-06-22

## Scope

Captured AMT 8000 Pro mobile-app traffic through `amt8000-capture-proxy` while
running Online actions for zone status, temporary bypass, PGM access, connection
status, and network-card status.

## Environment

- Panel: AMT 8000 Pro on local TCP `9009`; IP omitted
- Proxy listen: `0.0.0.0:9009`
- Mobile client observed by proxy; IP omitted
- Capture session: `20260622T161750.703591000Z`

## Observed Commands

| UTC time | Action | Command | Request payload | Response | Result |
| --- | --- | --- | --- | --- | --- |
| `2026-06-22T16:18:37Z` and later | zone/sectors status | `0x0b73`, `0x0b74` | `01` | status payloads | read-only status displayed |
| `2026-06-22T16:19:51.415235Z` | bypass zone 1 | `0x401f` | `00010100020003000400` | `0xf0fe` | zone 1 bypassed |
| `2026-06-22T16:20:06.469198Z` | clear zone 1 bypass | `0x401f` | `00000100020003000400` | `0xf0fe` | zone 1 un-bypassed |
| `2026-06-22T16:21:17.446746Z` | PGM 1 attempt | `0x45af` | `0001` | app error | PGM not accessible |
| `2026-06-22T16:21:32.466911Z` | PGM 1 retry | `0x45af` | `0001` | app error | PGM not accessible |
| `2026-06-22T16:21:56.633741Z` | connection status | `0x0b71` | empty | `0301` | read-only status displayed |
| `2026-06-22T16:22:24Z` | network-card status | `0x3812`, `0x3813`, `0x3814`, `0x3815`, `0x3faa` | see protocol doc | read payloads | read-only status displayed |
| `2026-06-22T16:23:10.360978Z` | bypass zone 2 | `0x401f` | `00000101020003000400` | `0xf0fe` | zone 2 bypassed |
| `2026-06-22T16:23:22.239772Z` | clear zone 2 bypass | `0x401f` | `00000100020003000400` | `0xf0fe` | zone 2 un-bypassed |

## Verification

- The app issued status reads after every bypass change.
- The LAN console implementation verifies bypass changes with status command
  `0x0b4a` and checks the target zone's `bypassed` bit.
- PGM remains blocked because the official app reported the output as
  inaccessible in this installation.

## Sanitization

- Remote password and authentication frames are not recorded here.
- Raw capture file remains outside the repository.
- This report stores only command IDs, command payloads, response summaries,
  timestamps, and high-level observed results.
