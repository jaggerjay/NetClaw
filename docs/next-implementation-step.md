# Immediate Next Implementation Step

NetClaw has now been manually verified on macOS for the main path:
- app builds and runs
- proxy-core starts
- HTTP capture works
- HTTPS MITM works
- HAR export works

The next coding milestone should focus on polish rather than first-pass scaffolding.

## Step 1: clean up remaining UI rough edges
- improve response body preview behavior for very large bodies
- improve list-row status/capture indicators
- tighten export UX now that HAR export is functional

## Step 2: improve reproduction/debug workflows
- add curl export for individual sessions
- improve copy/share actions for captured requests
- make it easier to move from captured traffic to reproduction commands

## Step 3: reduce setup friction on macOS
- bundle proxy-core into the app
- reduce dependence on a local Go environment for normal testing
- improve guidance or automation for proxy and CA setup
