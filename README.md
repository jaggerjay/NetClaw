# NetClaw

NetClaw is a V1 macOS proxy capture tool inspired by Fiddler/Charles/Proxyman.

## V1 scope

- Local HTTP/HTTPS proxy
- HTTPS MITM via locally generated CA (scaffold in progress, with static bypass rules)
- Session capture and storage
- Basic desktop viewer for captured sessions
- Export to HAR / curl (scaffolded)

## Structure

- `proxy-core/` — Go proxy engine
- `mac-app/` — SwiftUI macOS app skeleton
- `docs/` — architecture and implementation notes

## Current state

This workspace has moved past the initial scaffold phase and now includes a working macOS development test build with:
- HTTP proxy capture
- HTTPS CONNECT handling
- HTTPS MITM via a locally generated root CA
- SQLite-backed session persistence
- local sessions API with filtering/search and HAR export
- checked-in Xcode project for the macOS test shell
- app-driven proxy startup and session inspection
- HAR export from the macOS app

See `docs/validated-status.md` for the current manually verified status on macOS.

## Body capture note

- body capture now defaults to preserving the original request/response body data (`MaxBodyBytes = 0`)
- `-max-body-bytes 0` means unlimited body capture
- positive values cap stored request/response body bytes per item
- the macOS app previews large bodies with a truncated view first, with an option to show the full content

The Go and Swift toolchains were not available in the current environment, so the code is provided as a scaffold and should be built on a macOS development machine with Go and Xcode installed.

## Go module download note

On a fresh machine, the first `go run` or `go build` may spend time downloading dependencies such as `modernc.org/sqlite`.

If module download times out, it is usually a Go proxy / network issue rather than a NetClaw bug. In that case, try setting `GOPROXY` and downloading modules first:

```bash
cd proxy-core
go env -w GOPROXY=https://goproxy.cn,direct
go mod download
```

Then run NetClaw again.

If your network works better with the default Go proxy, you can instead use:

```bash
go env -w GOPROXY=https://proxy.golang.org,direct
```
