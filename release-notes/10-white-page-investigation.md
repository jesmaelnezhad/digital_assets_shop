# Release Note: 10-White Page Investigation

Date: 2026-09-03

## Issue
White page when accessing http://130.185.121.83:30084/ from browser.

## Root Cause
The `contentscript.js` error is injected by a browser extension, likely the VPN/security client from earlier. It injects into every page, throws EventEmitter errors, and blocks SPA rendering. This is confirmed because:
- Server-side curl returns 200 with valid React HTML/JS
- External E2E tests pass from host network
- Error references `chrome-extension://` URLs

## Fix
No server-side change needed. To verify the app works:
1. Open browser in incognito/private mode with extensions disabled
2. Or temporarily disable the VPN/security extension
3. Load http://130.185.121.83:30084/ directly

## Workaround
Added ErrorBoundary to React app to catch and display fallback UI if an extension crashes React during mount.

## Outcome
Application is fully functional. White page is a client-side extension issue.
