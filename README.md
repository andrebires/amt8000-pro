# AMT 8000 Pro LAN Console

Local web console for Intelbras AMT 8000 Pro panels.

The first implementation target is a read-only status dashboard over the local
ISECNet v2 protocol on TCP port `9009`. Programming writes are intentionally
later-phase work and must be validated with read-after-write checks against the
local panel.

## Requirements

- Go 1.26+
- AMT 8000 Pro reachable from the host running this service
- Remote access password from the panel QR label or configured installer docs

## Configuration

Copy `.env.example` to `.env` locally, then set:

```sh
AMT_HOST=192.168.1.50
AMT_PORT=9009
AMT_PASSWORD=878787
AMT_HTTP_ADDR=:8080
```

Do not commit `.env` or packet captures.

## Run

```sh
go run ./cmd/amt8000-pro
```

Open `http://localhost:8080`.

## Test

Unit tests:

```sh
go test ./...
```

Real-panel status smoke test:

```sh
AMT_HOST=192.168.1.50 AMT_PASSWORD=878787 scripts/production-status-test.sh
```

The script writes a Markdown report under `docs/test-runs/`.

## Safety Policy

- Read-only features can be tested directly against the panel.
- Arm/disarm requires an explicit manual safety checklist before implementation.
- Programming/configuration writes require packet-capture evidence and
  read-after-write verification.
- Firmware update, memory unlock, reset, and destructive commands are out of
  scope until explicitly requested.

