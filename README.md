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

This workspace currently contains a first-pass project skeleton and design docs, including:
- HTTP proxy capture scaffold
- HTTPS CONNECT passthrough scaffold
- local CA / leaf certificate generation scaffold
- MITM request/response and fallback scaffolds
- upstream TLS transport/error-handling scaffold
- SQLite-backed session persistence
- local sessions API with basic filtering (`q`, `host`, `method`, `has_error`, `tls_intercepted`, `limit`)
- SwiftUI viewer shell with basic filter controls

## Body capture note

- body capture now defaults to preserving the original request/response body data (`MaxBodyBytes = 0`)
- `-max-body-bytes 0` means unlimited body capture
- positive values cap stored request/response body bytes per item
- the macOS app previews large bodies with a truncated view first, with an option to show the full content

The Go and Swift toolchains were not available in the current environment, so the code is provided as a scaffold and should be built on a macOS development machine with Go and Xcode installed.
