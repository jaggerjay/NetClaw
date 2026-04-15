# NetClaw V1 Plan

## Done in this scaffold
- repository structure
- Go proxy core skeleton
- in-memory session store
- local JSON API skeleton
- SwiftUI app source skeleton
- architecture notes

## Remaining for usable V1

### Proxy core
1. Implement HTTPS CONNECT tunnel
2. Add root CA generation and persistence
3. Add dynamic leaf certificate minting
4. MITM intercepted HTTPS requests through the same capture pipeline
5. Add body truncation markers and binary handling
6. Replace memory store with SQLite
7. HAR export
8. curl export
9. filtering/search API

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
