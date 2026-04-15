# Immediate Next Implementation Step

To make NetClaw V1 genuinely usable as a Fiddler-like tool, the next coding milestone should be:

## Step 1: implement HTTPS CONNECT support

Minimum viable behavior:
- accept CONNECT host:443
- hijack client connection
- establish upstream TCP connection
- tunnel bytes in both directions
- record a minimal session entry with host, port, start/end time, and connect success/failure

This gives:
- working HTTPS passthrough through the proxy
- a foundation for later MITM interception

## Step 2: add local CA subsystem
- generate root CA once
- persist to disk
- issue per-host leaf certificates

## Step 3: switch CONNECT from passthrough to MITM
- terminate TLS from client
- create upstream TLS to server
- parse HTTP over both sides
- route through capture pipeline
