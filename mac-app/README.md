# NetClaw macOS App

SwiftUI-based macOS test shell for the NetClaw proxy core.

## Current test-shell features

- configurable local API base URL
- connection health indicator
- auto-refreshing session list
- session search and filters
- request / response detail view
- certificate authority info panel
- empty-state guidance for manual testing

## Expected local API

By default the app expects the proxy-core API at:

- `http://127.0.0.1:9091`

You can change the API base URL inside the app to point at another local or remote development instance.

## To run on macOS

1. Create an Xcode macOS app project named `NetClaw`
2. Add the files from `mac-app/NetClaw/` to the target
3. Start the proxy-core separately
4. Launch the app and confirm the health indicator turns green

## Next app steps

- start / stop proxy-core from the app
- proxy setup and CA trust guidance
- export selected sessions
- richer body rendering for JSON / binary payloads
