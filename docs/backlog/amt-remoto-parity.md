# AMT Remoto Parity Backlog

Source: Intelbras `Manual_AMT_Remoto_01-21_site - Arquivo Final.pdf`, published at
`https://backend.intelbras.com/sites/default/files/2024-09/Manual_AMT_Remoto_01-21_site%20-%20Arquivo%20Final.pdf`.

Secondary source: Intelbras `Manual_programador_AMT_8000_01-21_site.pdf`,
published at
`https://backend.intelbras.com/sites/default/files/2021-12/Manual_programador_AMT_8000_01-21_site.pdf`.

This is the working backlog for building AMT Remoto / Programador AMT 8000-style
capabilities into this LAN web console. The AMT Remoto manual is treated as a
suite capability inventory; the Programador AMT 8000 manual is treated as the
AMT 8000-specific UI/configuration taxonomy. Neither manual is protocol proof.
Every item that talks to the panel needs AMT 8000 Pro packet/protocol evidence
and a real-panel test report before it can be marked done.

## Status Legend

- `[ ]` Not started
- `[~]` In progress
- `[x]` Done
- `[blocked]` Blocked by missing protocol evidence, hardware, or product decision
- `[defer]` Explicitly outside the current LAN-first scope

## Definition Of Done For Panel Features

- [ ] Protocol command, response, and payload format documented under `docs/protocol/`.
- [ ] Unit tests or golden fixtures cover frame encoding/decoding and payload parsing.
- [ ] UI/API uses confirmation for control or configuration writes.
- [ ] Mutating features perform read-after-write verification where the panel supports it.
- [ ] Real-panel evidence is recorded under `docs/test-runs/`.
- [ ] Failure modes are visible in the UI without leaking passwords or raw secrets.

## Parity Milestones

### M0 - Current Read-Only LAN Console

- [x] Browser login with panel IP, port, and remote password.
- [x] Local Ethernet connection to panel over TCP `9009`.
- [x] Online dashboard shell.
- [x] Basic status parsing for model, firmware version, partitions, zones, siren, tamper, and battery.
- [x] Record first successful real-panel status test.
- [x] Save sanitized status frame fixture from the real panel.

### M1 - Online Tab Parity

Manual basis: the Online tab supports arm/disarm, PGM control, computer-time sync,
zone and panel status, source/battery voltage, pending problems, model/version,
clearing triggers, and temporary zone bypass.

- [x] `ONLINE-001` Add explicit safety checklist for live control testing.
- [x] `ONLINE-002` Implement full panel status refresh loop with connection timer.
- [x] `ONLINE-003` Display all zone states: open, closed, fired-open, fired-closed.
- [x] `ONLINE-004` Display partition state and siren/firing indicators.
- [blocked] `ONLINE-005` Display source voltage and battery voltage if exposed by AMT 8000 Pro status payloads. Current API/UI exposes nullable unsupported fields; payload offsets are not proven.
- [x] `ONLINE-006` Display pending problems/troubles from known status evidence: panel tamper, panel battery level, zone tamper, and zone low battery.
- [ ] `ONLINE-007` Implement arm/disarm commands with confirmation and production evidence.
- [blocked] `ONLINE-008` Implement PGM list/status read. Blocked by missing AMT 8000 Pro protocol evidence.
- [ ] `ONLINE-009` Implement PGM activation/deactivation with confirmation.
- [x] `ONLINE-010` Implement panel date/time read.
- [ ] `ONLINE-011` Implement sync panel date/time from server time.
- [ ] `ONLINE-012` Implement temporary zone bypass/un-bypass.
- [ ] `ONLINE-013` Implement clear fired-zone/alarm memory command if supported.
- [ ] `ONLINE-014` Add API endpoints for each Online command with audit log entries.
- [ ] `ONLINE-015` Add real-panel test reports for every Online command.

### M2 - Device Discovery And Connection Management

Manual basis: AMT Remoto includes a local network buscador that shows device IPs
and MACs, client connection profiles, connection status, connection elapsed time,
and automatic disconnection.

- [ ] `CONN-001` Research AMT 8000 Pro local discovery mechanism.
- [ ] `CONN-002` Implement LAN device discovery with IP and MAC display.
- [ ] `CONN-003` Add saved local panel profiles for trusted LAN devices.
- [ ] `CONN-004` Store connection profiles without storing remote passwords by default.
- [ ] `CONN-005` Add active connection state: connected user, panel name, elapsed time.
- [ ] `CONN-006` Add configurable automatic disconnect timer, default 10 minutes.
- [ ] `CONN-007` Support 10-60 minute disconnect policy if we decide to match AMT Remoto exactly.
- [ ] `CONN-008` Add manual disconnect action.
- [ ] `CONN-009` Add reconnect flow that preserves the selected panel profile.
- [ ] `CONN-010` Add real-panel test report for reconnect and automatic disconnect behavior.

### M3 - Local Users, Roles, And Audit Trail

Manual basis: AMT Remoto has administrator, supervisor, and operator profiles,
local software users, access history, and supervision to release a client locked
in editing.

- [ ] `AUTH-001` Decide whether local app users are required for home LAN use.
- [ ] `AUTH-002` Add local app user model: administrator, supervisor, operator.
- [ ] `AUTH-003` Add login/session model for app users, separate from panel remote password.
- [ ] `AUTH-004` Restrict configuration writes and sensitive password views to administrators.
- [ ] `AUTH-005` Restrict operators to Online/status/control-only surfaces.
- [ ] `AUTH-006` Add access history view filtered by date.
- [ ] `AUTH-007` Audit all panel connections, disconnects, reads, controls, and writes.
- [ ] `AUTH-008` Add edit lock/supervision model for configuration sessions.
- [ ] `AUTH-009` Add administrator release action for stale edit locks.
- [ ] `AUTH-010` Add tests for role permissions and audit history.

### M4 - Configuration Download

Manual basis: AMT Remoto can download central programming and automatically save
the downloaded data. Programador AMT 8000 exposes configuration sections for
Geral, Usuarios, Setores, Comunicacao, Monitoramento IP, Ethernet/WiFi, GPRS,
Auto-ativacao, Dispositivos, Eventos monitoramento, and Eventos Push.

- [ ] `CFG-DL-001` Capture AMT Remoto Ethernet download session for AMT 8000 Pro.
- [ ] `CFG-DL-002` Identify all read commands used during full programming download.
- [ ] `CFG-DL-003` Define versioned local configuration snapshot format.
- [ ] `CFG-DL-004` Implement raw configuration download with progress reporting.
- [ ] `CFG-DL-005` Store immutable downloaded snapshots.
- [ ] `CFG-DL-006` Parse high-confidence fields first; keep unknown blocks as raw bytes.
- [ ] `CFG-DL-007` Add snapshot diff view between two downloads.
- [ ] `CFG-DL-008` Add export of raw and parsed configuration snapshots.
- [ ] `CFG-DL-009` Add real-panel test report for full configuration download.
- [ ] `CFG-DL-010` Map configuration category: Geral.
- [ ] `CFG-DL-011` Map configuration category: Usuarios.
- [ ] `CFG-DL-012` Map configuration category: Setores.
- [ ] `CFG-DL-013` Map configuration category: Comunicacao.
- [ ] `CFG-DL-014` Map configuration category: Monitoramento IP.
- [ ] `CFG-DL-015` Map configuration category: Ethernet/WiFi.
- [ ] `CFG-DL-016` Map configuration category: GPRS.
- [ ] `CFG-DL-017` Map configuration category: Auto-ativacao.
- [ ] `CFG-DL-018` Map configuration category: Dispositivos.
- [ ] `CFG-DL-019` Map configuration category: Eventos monitoramento.
- [ ] `CFG-DL-020` Map configuration category: Eventos Push.

### M5 - Configuration Editing And Sending

Manual basis: AMT Remoto supports editing programming, saving the edit, and
sending configuration to the panel. Programador AMT 8000 provides the AMT 8000
configuration category taxonomy.

- [ ] `CFG-WR-001` Establish write safety policy for configuration programming.
- [ ] `CFG-WR-002` Require a fresh backup/download before any write session.
- [ ] `CFG-WR-003` Implement edit draft model separate from live panel state.
- [ ] `CFG-WR-004` Add diff preview before sending any configuration.
- [ ] `CFG-WR-005` Implement one low-risk write command first after protocol evidence.
- [ ] `CFG-WR-006` Add read-after-write verification for that first setting.
- [ ] `CFG-WR-007` Gradually add supported configuration categories from the AMT 8000 Pro manual.
- [ ] `CFG-WR-008` Add rollback guidance using backups where the panel/protocol allows it.
- [ ] `CFG-WR-009` Add real-panel test report for every write-capable setting.
- [blocked] `CFG-WR-010` Firmware update, factory reset, memory unlock, and destructive operations remain blocked until explicitly requested.
- [ ] `CFG-WR-011` Add guarded edit/send support for Geral after read mapping is stable.
- [ ] `CFG-WR-012` Add guarded edit/send support for Usuarios after read mapping is stable.
- [ ] `CFG-WR-013` Add guarded edit/send support for Setores after read mapping is stable.
- [ ] `CFG-WR-014` Add guarded edit/send support for Comunicacao after read mapping is stable.
- [ ] `CFG-WR-015` Add guarded edit/send support for Monitoramento IP after read mapping is stable.
- [ ] `CFG-WR-016` Add guarded edit/send support for Ethernet/WiFi after read mapping is stable.
- [ ] `CFG-WR-017` Add guarded edit/send support for GPRS after read mapping is stable.
- [ ] `CFG-WR-018` Add guarded edit/send support for Auto-ativacao after read mapping is stable.
- [ ] `CFG-WR-019` Add guarded edit/send support for Dispositivos after read mapping is stable.
- [ ] `CFG-WR-020` Add guarded edit/send support for Eventos monitoramento after read mapping is stable.
- [ ] `CFG-WR-021` Add guarded edit/send support for Eventos Push after read mapping is stable.

### M6 - Backup And Restore

Manual basis: AMT Remoto creates backups when saving current configuration,
saving an edit, or downloading/editing/sending complete configuration, and can
recover a selected backup with confirmation.

- [ ] `BKP-001` Create local backup metadata model.
- [ ] `BKP-002` Create backup automatically after successful configuration download.
- [ ] `BKP-003` Create backup automatically before configuration send.
- [ ] `BKP-004` Add backup list and detail view.
- [ ] `BKP-005` Add backup export/import.
- [ ] `BKP-006` Add backup restore dry-run diff.
- [ ] `BKP-007` Implement restore only after full write protocol confidence.
- [ ] `BKP-008` Add confirmation and real-panel evidence for restore.

### M7 - Events

Manual basis: AMT Remoto can download and save the latest 256 events, and marks
events disabled/blocked for receptor-IP sending. Programador AMT 8000 exposes a
buffer de eventos view from the main menu.

- [ ] `EVT-001` Capture event download command and response format.
- [ ] `EVT-002` Implement download of latest panel events.
- [ ] `EVT-003` Parse event timestamp, code, partition, zone/user, and delivery status if present.
- [ ] `EVT-004` Display last 256 events with filters.
- [ ] `EVT-005` Mark events blocked/disabled for receptor-IP sending if protocol exposes it.
- [ ] `EVT-006` Export events as CSV.
- [ ] `EVT-007` Export events as JSON.
- [ ] `EVT-008` Add real-panel test report for event download/export.

### M8 - AMT Remoto Transport Parity

Manual basis: AMT Remoto supports Ethernet, serial/USB, modem/telephone line,
receptor-IP account, and Intelbras Cloud via MAC. Programador AMT 8000 lists
IP LOCAL, CONTA via RECEPTOR IP, and MAC via CLOUD. This project is LAN-first,
so non-LAN transports stay lower priority until the local Ethernet protocol is
solid.

- [x] `TRN-001` Ethernet direct IP/port connection.
- [defer] `TRN-002` Serial/USB connection support.
- [defer] `TRN-003` Modem/V-21 telephone-line connection support.
- [defer] `TRN-004` Receptor-IP account connection support on port `9010`.
- [defer] `TRN-005` Intelbras Cloud/MAC connection support.
- [ ] `TRN-006` Document why each deferred transport is not needed for LAN-first use.

### M9 - Documentation, Help, And About

Manual basis: AMT Remoto includes manual/about surfaces and exposes software
version information.

- [ ] `DOC-001` Add in-app About view with app version and protocol support level.
- [ ] `DOC-002` Add local help page that links to Intelbras official manuals.
- [ ] `DOC-003` Add protocol support matrix for AMT 8000 Pro.
- [ ] `DOC-004` Add production test index linked to `docs/test-runs/`.
- [ ] `DOC-005` Add troubleshooting page for connection/auth/status failures.

## Protocol Discovery Backlog

- [ ] `PROTO-001` Build a packet-capture checklist for AMT Remoto Desktop over Ethernet.
- [ ] `PROTO-002` Set up a Windows VM/PC on the same LAN as the panel.
- [ ] `PROTO-003` Capture connect/auth/disconnect session.
- [ ] `PROTO-004` Capture Online status refresh session.
- [ ] `PROTO-005` Capture arm/disarm session with safe panel state.
- [ ] `PROTO-006` Capture PGM toggle session if a safe PGM is available.
- [ ] `PROTO-007` Capture bypass/un-bypass session on a safe test zone.
- [ ] `PROTO-008` Capture date/time sync session.
- [ ] `PROTO-009` Capture event download session.
- [ ] `PROTO-010` Capture full configuration download session.
- [ ] `PROTO-011` Capture configuration section reads for Geral, Usuarios, Setores, Comunicacao, Monitoramento IP, Ethernet/WiFi, GPRS, Auto-ativacao, Dispositivos, Eventos monitoramento, and Eventos Push.
- [ ] `PROTO-012` Capture one low-risk configuration write session.
- [ ] `PROTO-013` Sanitize captures and commit only derived fixtures/notes, never secrets.

## Priority Recommendation

1. Finish M0 evidence: record successful real-panel status and sanitized fixture.
2. Implement M1 Online read completeness before any new write command.
3. Add M2 connection lifecycle and automatic disconnect.
4. Add M7 events, because it is likely read-only and useful.
5. Start M4 configuration download.
6. Only then start M5 configuration writes.
