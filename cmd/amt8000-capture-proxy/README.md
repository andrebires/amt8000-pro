# AMT 8000 Capture Proxy

`amt8000-capture-proxy` is a local TCP pass-through proxy for observing
Intelbras AMT 8000 Pro LAN protocol sessions. It is meant for protocol
discovery with the official Android/Desktop applications while keeping the
panel traffic path otherwise unchanged.

The proxy listens on a local address, forwards bytes to the real panel, and
writes a JSONL capture file with frame metadata and hex payloads. Authentication
frames and client chunks containing the auth command are redacted, but captures
can still contain sensitive panel state and should not be committed.

## Usage

Point the official app at the proxy address and port, then run the action you
want to study.

```sh
AMT_HOST=192.168.0.102 \
AMT_PORT=9009 \
AMT_PROXY_ADDR=0.0.0.0:9009 \
AMT_CAPTURE_OUT=/tmp/amt8000-captures/events.jsonl \
go run ./cmd/amt8000-capture-proxy
```

Environment variables:

| Name | Required | Default | Purpose |
| --- | --- | --- | --- |
| `AMT_HOST` | yes | none | Real panel IP or hostname. |
| `AMT_PORT` | no | `9009` | Real panel TCP port. |
| `AMT_PROXY_ADDR` | no | `127.0.0.1:19009` | Local address the official app connects to. |
| `AMT_CAPTURE_OUT` | no | `/tmp/amt8000-captures/<timestamp>-isecnet.jsonl` | JSONL capture destination. |

## Safety Notes

- Use this only on a trusted LAN with your own panel.
- Do not publish raw capture files without reviewing them.
- Keep the proxy read-only: it forwards the official app's traffic and should
  not inject, replay, or modify frames.
- Prefer writing captures under `/tmp/amt8000-captures` so they stay out of the
  repository.
