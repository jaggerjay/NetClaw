# NetClaw V1 Plan

## Verified working now
- repository structure
- Go proxy core
- HTTP proxy capture
- HTTPS CONNECT handling
- HTTPS MITM with locally generated root CA
- upstream TLS transport and error classification
- static MITM bypass rules and temporary MITM failure backoff cache
- SQLite-backed session persistence
- local JSON API
- filtering/search API for session lists
- runtime-info API for the macOS app
- HAR export API
- checked-in Xcode project for the macOS test shell
- macOS app start/stop for local proxy-core
- macOS app session list/detail UI
- setup guidance for proxy configuration and CA trust
- HAR export from the macOS app

## Verified manually on macOS
- Xcode project opens and builds
- app launches successfully
- proxy-core can be started from the app
- HTTP capture works end-to-end
- HTTPS MITM works end-to-end
- HAR export works end-to-end

## Remaining for a more polished V1

### Proxy core
1. Improve fallback policy for pinned / broken hosts beyond current static + temporary rules
2. Add curl export / replay-oriented helpers
3. Improve capture fidelity for more real-world edge cases
4. Add more protocol / site compatibility testing coverage

### macOS app
1. Improve response body preview UX for very large bodies
2. Add richer list-row status indicators and polish
3. Improve export / save UX beyond the current Downloads-first workaround
4. Add system proxy setup guidance / automation
5. Add certificate installation / trust automation guidance where feasible

### Packaging
1. bundle the proxy-core binary into the app
2. reduce dependence on a local Go toolchain for normal testing
3. codesign / notarization planning
4. app bundle structure cleanup for distribution

## Suggested implementation order from here
1. polish body preview and remaining UI issues
2. improve export / reproduction workflows (curl export, better save UX)
3. bundle proxy-core for a smoother macOS testing experience
4. expand MITM compatibility testing and fallback behavior
