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

## Dynamic fallback behavior
NetClaw now also has a temporary in-memory fallback path for MITM failures.
If a host fails during MITM handling, it can be marked as temporarily bypassed for a backoff window so later connections go straight to passthrough.

## Current backoff window
- default: 10 minutes
- scope: in-memory only (cleared on process restart)

## Future improvements
- UI-managed bypass list
- persist temporary bypass state across restarts
- session labeling for bypass reason
- wildcard and regex policies
