# NetClaw Validated Status

This document records what has been manually verified on a real macOS machine.

## Verified working

### Project / build
- `mac-app/NetClaw.xcodeproj` opens in Xcode
- the macOS app builds successfully
- the macOS app launches successfully

### Local proxy runtime
- proxy-core can be launched from the macOS app
- Go dependency download issues were solved by configuring `GOPROXY`
- the local API responds on `http://127.0.0.1:9091/health`

### Capture path
- plain HTTP capture works
- HTTPS CONNECT handling works
- HTTPS MITM works with the generated NetClaw root CA

### App inspection path
- captured sessions appear in the session list
- session detail view works for inspected requests
- HAR export works from the macOS app

## Known rough edges still worth polishing
- response body preview UX still needs cleanup for some large responses
- export UX uses a Downloads-first workaround instead of a polished save flow
- the app still depends on a local Go toolchain to start proxy-core during testing
- the overall setup flow is still developer-oriented rather than installer-grade

## Practical conclusion
NetClaw is no longer just a scaffold. It currently behaves like a working macOS development test build for HTTP/HTTPS proxy capture and HAR export.
