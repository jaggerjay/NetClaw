# NetClaw macOS Test Shell - Xcode Setup

## Open the checked-in project

1. Open Xcode
2. Open `mac-app/NetClaw.xcodeproj`
3. Select the `NetClaw` scheme
4. Build and run

## Local permissions

The current test shell has app sandbox disabled in project settings so it can talk to the local HTTP API without extra entitlement work during development.

## Start proxy-core

You now have two options.

### Option A: start it from inside the app

When the app opens:

1. Click **Choose…** and select your local `proxy-core` folder
2. Click **Suggested** to fill a launch command
3. Or click **Build+Run** if you want the app to build a local debug binary first
4. Click **Validate** to confirm the directory / command / Go setup look sane
5. Click **Start Proxy**
6. Confirm proxy logs appear and the API status turns green
7. Use **Quick Check** if you want to ping the local API on demand
8. The app will remember your recent launch settings
9. Use the built-in Setup Guide panel for proxy host/port and CA trust reminders
10. Inspect CONNECT sessions in the detail view to verify capture mode and tunnel byte counts
11. Use **Export HAR** to save the currently filtered request set as a HAR file
12. Check body panes for JSON, XML, and form-urlencoded formatting, plus image preview and truncation hints during testing
13. Large bodies now open in preview mode first; use Show All when you need the full captured content
14. Use the detail view copy actions to grab URL, headers, bodies, or a curl reproduction command

### Option B: start it manually from Terminal

From the repository root:

```bash
cd proxy-core
go run ./cmd/netclaw-proxy -proxy-listen 127.0.0.1:9090 -api-listen 127.0.0.1:9091 -data-dir .netclaw-data/dev
```

### If Go dependency download times out

If `go run` stalls or times out while downloading dependencies such as `modernc.org/sqlite`, configure a working Go proxy first:

```bash
cd proxy-core
go env -w GOPROXY=https://goproxy.cn,direct
go mod download
```

After that, retry `go run` or the in-app Start Proxy flow.

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
