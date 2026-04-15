# Immediate Next Implementation Step

NetClaw now has a scaffold for HTTPS CONNECT passthrough. The next coding milestone should be:

## Step 1: add local CA subsystem
- generate root CA once
- persist to disk
- issue per-host leaf certificates
- expose certificate path / trust guidance to the UI

## Step 2: switch CONNECT from passthrough to MITM
- terminate TLS from client
- create upstream TLS to server
- parse HTTP over both sides
- route decrypted requests through the same capture pipeline

## Step 3: improve captured session fidelity
- record byte counts for CONNECT tunnels
- distinguish tunnel errors vs target server errors
- persist sessions to SQLite
- add filter/search API
