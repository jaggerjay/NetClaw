# NetClaw V1 Plan

## Done in this scaffold
- repository structure
- Go proxy core skeleton
- HTTPS CONNECT passthrough scaffold
- local CA / certificate management scaffold
- first-pass HTTPS MITM request path scaffold
- buffered MITM response serialization scaffold
- static MITM bypass rules
- in-memory session store
- local JSON API skeleton
- SwiftUI app source skeleton
- architecture notes

## Remaining for usable V1

### Proxy core
1. Validate and harden MITM intercepted HTTPS flow
2. Improve HTTPS CONNECT tunnel session metadata
3. Add body truncation markers and binary handling
4. Replace memory store with SQLite
5. HAR export
6. curl export
7. filtering/search API
8. add passthrough fallback policy for pinned / broken hosts

### macOS app
1. Create Xcode project and wire these sources
2. Start/stop bundled proxy process
3. Poll local API
4. Render list/details
5. Certificate installation guidance
6. System proxy setup guidance / automation

### Packaging
1. app bundle structure
2. embed Go binary
3. codesign / notarization planning

## Suggested implementation order
1. plain HTTP capture verification
2. CONNECT blind tunnel
3. MITM CA subsystem
4. session persistence
5. UI wiring
6. export features
