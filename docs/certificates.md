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

## Important
At this stage, certificate generation is ready, but the proxy does **not yet** use these certificates to MITM HTTPS traffic. CONNECT still operates in passthrough mode.

## Planned usage in the next step
1. On CONNECT for `example.com:443`
2. Generate or fetch a leaf cert for `example.com`
3. Perform TLS handshake with the client using that leaf cert
4. Establish upstream TLS to the real server
5. Parse decrypted HTTP messages and capture them

## macOS trust flow
For real HTTPS interception, the root CA certificate must be imported into Keychain Access and marked as trusted.
