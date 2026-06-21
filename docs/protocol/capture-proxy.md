# ISECNet Capture Proxy

Use the capture proxy when a command is only known through AMT Remoto or
Programador AMT 8000. The proxy accepts the official application's TCP
connection, forwards it to the real panel, and writes one redacted JSONL record
per parsed ISECNet frame.

## Run

```sh
AMT_HOST=192.168.1.50 \
AMT_PROXY_ADDR=0.0.0.0:19009 \
go run ./cmd/amt8000-capture-proxy
```

Optional environment variables:

| Variable | Default | Meaning |
| --- | --- | --- |
| `AMT_HOST` | required | Real panel IP |
| `AMT_PORT` | `9009` | Real panel ISECNet port |
| `AMT_PROXY_ADDR` | `127.0.0.1:19009` | Local proxy bind address |
| `AMT_CAPTURE_OUT` | `/tmp/amt8000-captures/<timestamp>-isecnet.jsonl` | Redacted JSONL capture path |

If AMT Remoto or Programador AMT 8000 runs on another machine, bind the proxy to
`0.0.0.0:19009`, find this computer's LAN IP, then point the official app to
that LAN IP and port `19009`.

## Event Capture

1. Start the proxy.
2. In AMT Remoto or Programador AMT 8000, connect to the proxy address instead
   of the panel address.
3. Open the event buffer/download events view.
4. Download or refresh the latest events.
5. Stop the proxy.
6. Inspect the capture for the command sent immediately before the event-buffer
   response.

Auth frames with command `0xf0f0` are redacted because they contain the encoded
remote password. Do not commit unsanitized captures unless every sensitive field
has been reviewed.
