# nextendo-telemetry-nx

*(still in alpha testing)*

**A new service implementation by nx-mod** for the Nextendo Network.

Telemetry **sink** for [Nextendo Network](https://nextendo.network). Source only. Not affiliated with Nintendo.

## Why

Switch titles and the system upload telemetry to several Nintendo services — **prepo** / play-report (`receive-*.dg.srv.nintendo.net`), **erpt** / error-report (`receive-*.er.srv.nintendo.net`), and assorted analytics. On the Nextendo stack these fall through to Nintendo; a game that insists on a successful upload can stall or retry until it gets one. This accepts every telemetry request, counts it, discards it (or samples it), and answers success — so nothing waits on Nintendo.

## What it does

- **Permissive:** any method, any path, always `200 {}`.
- **Counts** by host + path (the dashboard doubles as a map of what each title reports).
- **Optional capture:** set `TELEMETRY_DUMP=<dir>` to keep bodies for research; empty discards them.

sni-router routes the telemetry hosts here (TLS passthrough), or set `CERT_FILE`/`KEY_FILE`.

## Run

```sh
go build -o server .   # Go 1.23+, stdlib only
go test ./...
./server
```

| setting | default | meaning |
|---|---|---|
| `TELEMETRY_PORT` | 8472 | HTTP(S) port |
| `TELEMETRY_DUMP` | (off) | keep bodies in this dir |
| `TELEMETRY_MAX_BODY` | 1048576 | max bytes read per request |
| `DASH_PORT`/`DASH_TOKEN` | 8101 | `/api/stats`, `/healthz` |

## Credits

- **[Nextendo Network](https://nextendo.network)** — the stack this plugs into.
- **[switchbrew](https://switchbrew.org/wiki/Services_API)** / **[kinnay/NintendoClients](https://github.com/kinnay/NintendoClients/wiki)** — the prepo/erpt telemetry endpoints.

Protocol facts were read and reimplemented; no code was copied.

## Credits

Built by nx-mod for the **Nextendo Network**, on the work of the Nextendo Network team — https://nextendo.network. Nextendo is awesome.
