# NetClaw

NetClaw is a V1 macOS proxy capture tool inspired by Fiddler/Charles/Proxyman.

## V1 scope

- Local HTTP/HTTPS proxy
- HTTPS MITM via locally generated CA
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
- SwiftUI viewer shell

The Go and Swift toolchains were not available in the current environment, so the code is provided as a scaffold and should be built on a macOS development machine with Go and Xcode installed.
