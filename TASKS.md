# AMT 8000 Pro Task List

Status legend: `[ ]` not started, `[~]` in progress, `[x]` done.

## Phase 0 - Project Bootstrap

- [x] Choose project tracking format: Markdown tasks plus ADRs.
- [x] Choose implementation stack: Go single-binary LAN web app.
- [x] Initialize canonical repo at `~/Repositories/amt8000-pro`.
- [x] Make first bootstrap commit.

## Phase 1 - Read-Only Status MVP

- [x] Scaffold Go module and package layout.
- [x] Add browser login form for panel IP, port, and remote password.
- [x] Implement ISECNet frame encoding and checksum.
- [x] Implement remote password encoding.
- [x] Implement TCP connect/auth/status/disconnect client.
- [x] Implement status parser for firmware, partitions, zones, siren, tamper, and battery.
- [x] Add server-rendered dashboard at `/`.
- [x] Add JSON status endpoint at `/api/status`.
- [x] Add unit tests for protocol primitives and parser.
- [x] Run unit tests from the canonical repo.
- [x] Run real-panel status smoke test.
- [x] Record production evidence under `docs/test-runs/`.

## Phase 2 - Safe Control

- [x] Add explicit safety checklist for arm/disarm.
- [ ] Implement arm/disarm protocol methods.
- [ ] Add dashboard controls with confirmation.
- [ ] Test arm/disarm against local panel.
- [ ] Record production evidence under `docs/test-runs/`.

## Phase 3 - Programming Discovery

- [ ] Run AMT Remoto Desktop in a Windows PC/VM on the LAN.
- [ ] Capture local-IP programming sessions to TCP `9009`.
- [ ] Map read commands for high-priority programming categories.
- [ ] Map write commands only after read behavior is understood.
- [ ] Add fixtures from sanitized captures.

## Phase 4 - Programming UI

- [ ] Add read-only configuration views.
- [ ] Add guarded write flows with diff preview.
- [ ] Add read-after-write verification for every setting.
- [ ] Keep firmware update, reset, and memory unlock disabled.

## Phase 5 - AMT Remoto / Programador AMT 8000 Feature Parity

Detailed backlog: `docs/backlog/amt-remoto-parity.md`.

- [x] Finish current read-only LAN console evidence and sanitized fixtures.
- [~] Implement Online tab parity: richer read-only status is implemented; arm/disarm, PGM, clock sync, bypass, and clear alarm memory remain evidence-gated.
- [ ] Implement LAN device discovery and saved panel profiles.
- [ ] Decide and implement local users/roles/audit if needed for multi-device LAN access.
- [ ] Implement read-only event download and export.
- [ ] Implement full configuration download and immutable snapshots.
- [ ] Map Programador AMT 8000 configuration sections: Geral, Usuarios, Setores, Comunicacao, Monitoramento IP, Ethernet/WiFi, GPRS, Auto-ativacao, Dispositivos, Eventos monitoramento, and Eventos Push.
- [ ] Implement configuration editing only after protocol evidence and backup/read-after-write safety gates.
- [ ] Implement backup/restore workflows after configuration write support is proven.
- [ ] Document deferred AMT Remoto transports: serial/USB, modem, receptor-IP account, and Intelbras Cloud.
