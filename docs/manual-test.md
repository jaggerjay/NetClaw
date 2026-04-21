# NetClaw Manual Test Guide

This is the quickest way to verify the current V1 proxy-core on a development machine.

## 1. Start the proxy core

```bash
cd proxy-core
go run ./cmd/netclaw-proxy \
  -proxy-listen 127.0.0.1:9090 \
  -api-listen 127.0.0.1:9091 \
  -data-dir .netclaw-data/dev-test
```

Expected startup logs include:
- proxy listening address
- api listening address
- root CA certificate path
- session database path

## 2. Send HTTP traffic through the proxy

```bash
curl --proxy http://127.0.0.1:9090 http://example.com/
```

## 3. Inspect captured sessions via the local API

```bash
curl http://127.0.0.1:9091/api/sessions
curl 'http://127.0.0.1:9091/api/sessions?host=example.com'
curl 'http://127.0.0.1:9091/api/sessions?q=example&limit=10'
```

## 4. Inspect one session in detail

Copy a session ID from `/api/sessions`, then:

```bash
curl http://127.0.0.1:9091/api/sessions/<session-id>
```

## 5. Verify persistence

1. Stop the proxy core
2. Start it again with the same `-data-dir`
3. Query `/api/sessions` again
4. Confirm previously captured sessions still exist

## Optional: HTTPS MITM testing

The HTTPS path still needs more hardening. For now:
- use the generated root CA at `.netclaw-data/dev-test/certs/netclaw-root-ca.pem`
- trust it on your macOS test machine
- configure your system/app proxy to `127.0.0.1:9090`
- test with a simple HTTP/1.1 HTTPS endpoint first

If a host breaks under MITM, NetClaw should fall back to passthrough after failure backoff.
