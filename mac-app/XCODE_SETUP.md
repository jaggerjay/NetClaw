# NetClaw macOS Test Shell - Xcode Setup

## Create the project

1. Open Xcode
2. Create a new project
3. Choose **App** under macOS
4. Product Name: `NetClaw`
5. Interface: `SwiftUI`
6. Language: `Swift`

## Add the source files

Add all files under:

- `mac-app/NetClaw/Models/`
- `mac-app/NetClaw/Services/`
- `mac-app/NetClaw/ViewModels/`
- `mac-app/NetClaw/Views/`
- `mac-app/NetClaw/NetClawApp.swift`

Make sure they are added to the main app target.

## Local permissions

No special sandbox entitlements are required for the current test shell because it only talks to the local HTTP API.

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
