# NetClaw macOS App

SwiftUI-based macOS test shell for the NetClaw proxy core.

## Current test-shell features

- checked-in `NetClaw.xcodeproj`
- start / stop local proxy-core from the app
- folder picker for local `proxy-core` path
- startup validation for working directory / command / Go availability
- more resilient Go discovery for Xcode-launched app processes
- editable working directory and launch command
- saved recent launch settings via `UserDefaults`
- inline proxy logs
- quick API health check and clearer error display
- setup guide for macOS proxy configuration and CA trust
- HAR export from the macOS test shell using current filters
- asynchronous save-panel based export flow to avoid blocking the UI during HAR export
- richer body rendering with JSON, XML, form-urlencoded formatting, image preview, and truncation hints
- preview-first body display with Show All / Show Less for large content
- one-click copy actions for URL, headers, bodies, and curl reproduction commands
- configurable local API base URL
- connection health indicator
- auto-refreshing session list
- session search and filters
- request / response detail view
- CONNECT / MITM detail metadata including capture mode and tunnel byte counts
- certificate authority info panel
- empty-state guidance for manual testing

## Expected local API

By default the app expects the proxy-core API at:

- `http://127.0.0.1:9091`

You can change the API base URL inside the app to point at another local or remote development instance.

## To run on macOS

1. Open `mac-app/NetClaw.xcodeproj`
2. Select the `NetClaw` scheme
3. Start the proxy-core separately
4. Run the app from Xcode
5. Confirm the health indicator turns green

## Quick proxy-core command

```bash
cd proxy-core
go run ./cmd/netclaw-proxy -proxy-listen 127.0.0.1:9090 -api-listen 127.0.0.1:9091 -data-dir .netclaw-data/dev
```

## Next app steps

- start / stop proxy-core from the app
- proxy setup and CA trust guidance
- export selected sessions
- richer body rendering for JSON / binary payloads
