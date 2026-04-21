# NetClaw macOS Test Shell - Xcode Setup

## Open the checked-in project

1. Open Xcode
2. Open `mac-app/NetClaw.xcodeproj`
3. Select the `NetClaw` scheme
4. Build and run

## Local permissions

The current test shell has app sandbox disabled in project settings so it can talk to the local HTTP API without extra entitlement work during development.

## Start proxy-core first

From the repository root:

```bash
cd proxy-core
go run ./cmd/netclaw-proxy -proxy-listen 127.0.0.1:9090 -api-listen 127.0.0.1:9091 -data-dir .netclaw-data/dev
```

## Launch the app

When the app opens:

1. Leave API Base URL as `http://127.0.0.1:9091`
2. Click **Apply** if needed
3. Confirm status becomes **Connected**
4. Send traffic through the proxy using your browser or curl
5. Inspect sessions in the sidebar and detail panel

## Quick local test

In another terminal:

```bash
curl --proxy http://127.0.0.1:9090 http://example.com
```

The request should appear in the app automatically if auto-refresh is enabled.

## Notes

- App icons are placeholder entries right now
- This is a development test shell, not a notarized distribution build
- The next logical step is teaching the app to launch and stop `proxy-core` itself
