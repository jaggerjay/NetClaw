# Immediate Next Implementation Step

NetClaw now has a first-pass MITM CONNECT scaffold. The next coding milestone should be:

## Step 1: validate and harden the MITM request/response path
- compile and fix any type/runtime issues
- ensure intercepted responses are serialized correctly
- verify keep-alive behavior and body framing
- test with plain HTTP/1.1 sites through the HTTPS proxy path

## Step 2: establish real upstream TLS policy
- validate current upstream TLS config and runtime behavior
- verify SNI behavior against real targets
- improve certificate validation UX and error surfacing
- improve fallback from MITM to passthrough when needed (beyond current static rules + temporary failure cache)

## Step 3: improve captured session fidelity
- record byte counts for CONNECT tunnels
- distinguish tunnel errors vs target server errors
- persist sessions to SQLite
- add filter/search API
