# NetClaw macOS Local Dev Setup

## Requirements
- macOS with Xcode installed
- Go 1.22+
- admin access for certificate trust / system proxy changes

## Proxy core
```bash
cd proxy-core
go mod tidy
go run ./cmd/netclaw-proxy
```

Proxy API:
- proxy listener: `127.0.0.1:9090`
- local viewer API: `127.0.0.1:9091`

## macOS app
1. Open Xcode
2. Create a new **macOS App** project named `NetClaw`
3. Replace generated Swift files with sources from `mac-app/NetClaw/`
4. Run the app locally

## Configure macOS proxy for testing
In System Settings -> Network -> active network -> Details -> Proxies:
- Web Proxy (HTTP): `127.0.0.1`, port `9090`
- Secure Web Proxy (HTTPS): `127.0.0.1`, port `9090`

## Current limitation
HTTPS MITM is not implemented yet in this scaffold, so HTTPS requests will not yet be decrypted.
