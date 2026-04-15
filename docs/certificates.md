# NetClaw Certificate Authority Notes

NetClaw now includes a local certificate authority scaffold.

## What it does
- Creates a persistent local root CA under `.netclaw-data/certs/`
- Stores:
  - `netclaw-root-ca.pem`
  - `netclaw-root-ca.key`
- Provides per-host leaf certificate generation in memory
- Exposes CA info over the local API:
  - `GET /api/certificate-authority`

## Current stage
The proxy now has a first-pass MITM handshake scaffold:
1. On CONNECT for `example.com:443`
2. Generate or fetch a leaf cert for `example.com`
3. Perform TLS handshake with the client using that leaf cert
4. Read decrypted HTTP requests from the client-side TLS stream
5. Forward captured requests through the existing HTTP pipeline

## Important limitation
This is still a scaffold, not a production-ready MITM engine. It needs real build-and-runtime validation, upstream TLS hardening, header/body correctness checks, and fallback logic.

## macOS trust flow
For real HTTPS interception, the root CA certificate must be imported into Keychain Access and marked as trusted.
