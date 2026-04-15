# internal/cert

This package owns NetClaw's local certificate authority lifecycle.

## Current contents
- root CA creation / persistence
- root CA loading
- per-host leaf certificate generation and caching
- API-friendly metadata via `Authority.Info()`

## Future additions
- trust status detection on macOS
- leaf certificate eviction policy
- certificate export helpers
- tls.Config integration for MITM CONNECT handling
