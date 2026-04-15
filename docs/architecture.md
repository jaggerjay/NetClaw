# NetClaw V1 Architecture

## Goal

Build a macOS proxy capture tool similar to a lightweight Fiddler:

- explicit HTTP/HTTPS proxy
- HTTPS MITM with local root CA
- session list and detail viewer
- local persistence
- future-friendly architecture for intercept/replay/rules

## Components

### 1. Proxy Core (Go)
Responsibilities:
- listen on local proxy port
- handle HTTP forwarding
- handle HTTPS CONNECT tunneling with MITM
- generate/install-ready CA materials
- dynamically mint leaf certificates per host
- capture request/response metadata and bodies
- expose local API for viewer/UI
- persist sessions to SQLite or JSONL-backed store

### 2. macOS App (SwiftUI)
Responsibilities:
- start/stop proxy core process
- show certificate/proxy setup guidance
- render session list
- render request/response details
- search/filter sessions
- export selected session(s)

## V1 Data Flow

1. App launches proxy core
2. Proxy core creates/loads root CA
3. User configures macOS proxy to 127.0.0.1:port and trusts CA
4. Client traffic reaches proxy core
5. Proxy core captures request/response and persists session
6. App polls local API for list/details
7. User inspects sessions in UI

## V1 Constraints

- Only explicit proxy traffic is captured
- SSL pinning may prevent HTTPS inspection
- HTTP/3 / QUIC not supported in V1
- WebSocket/gRPC only future placeholders

## Suggested next steps

1. Implement plain HTTP proxy path end-to-end
2. Add CONNECT handling and blind tunnel fallback
3. Add CA generation and MITM TLS interception
4. Add session persistence
5. Wire UI to local API
