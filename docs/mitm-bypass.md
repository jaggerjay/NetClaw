# NetClaw MITM Bypass Notes

NetClaw now includes a simple static MITM bypass mechanism.

## Purpose
Some targets should not go through HTTPS interception during early development, for example:
- localhost services
- pinned clients
- hosts that break during MITM testing

## Current behavior
If the CONNECT host matches `MITMBypassHosts`, NetClaw will:
1. acknowledge CONNECT
2. skip TLS interception
3. tunnel bytes directly to the upstream server

## Default bypass list
- `localhost`
- `127.0.0.1`

## Future improvements
- UI-managed bypass list
- auto-bypass after repeated TLS handshake failures
- session labeling for bypass reason
- wildcard and regex policies
