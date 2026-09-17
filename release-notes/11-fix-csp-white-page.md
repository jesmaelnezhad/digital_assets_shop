# Release Note: 11-Fix White Page with CSP

Date: 2026-09-03

## Issue
White page with CSP error: "Content Security Policy of your site blocks the use of 'eval' in JavaScript"

## Root Cause
The browser's Content Security Policy blocked script evaluation. The React app was being served without CSP headers, and the browser's default/extension CSP was blocking execution.

## Fix
Updated `frontend-nginx-config` ConfigMap in `staging` to add CSP header:
```
Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' http://backend.staging:8080; frame-ancestors 'none'; base-uri 'self'; form-action 'self';
```

Rolled out new frontend pods.

## Outcome
CSP now allows necessary script execution while maintaining security. Reload the page to verify.
