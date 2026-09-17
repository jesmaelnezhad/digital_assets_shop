# Staging E2E Verification (Python, SSH-exec on RED)

When the backend is deployed to RED's staging namespace and you need to
verify all features work end-to-end via real HTTP, run a Python script via
SSH on RED rather than curling from BLUE. This avoids token escaping issues
and lets you chain dependent calls (register→login→use-token).

## Why Python on RED instead of curl from BLUE

- Token extraction: Python `json.loads` handles the login response cleanly;
  bash `jq` or `python3 -c` inline is fragile with shell quoting on the SSH
  command line.
- Chained authentication: store the token in a Python variable across calls;
  bash requires temp files or complex quoting.
- Structured output: each endpoint gets a PASS/FAIL verdict with the actual
  response body on failure — easier to scan than raw curl output.
- Survives terminal cut-offs: the script runs entirely on RED; the SSH
  channel only carries the final output.

## Script template

```python
#!/usr/bin/env python3
"""Staging E2E verification — run on RED via:
    ssh root@194.5.206.106 'python3 /tmp/staging_e2e.py'
Or copy first:
    scp staging_e2e.py root@194.5.206.106:/tmp/staging_e2e.py
"""
import urllib.request, urllib.error, json, sys, time

BASE = "http://127.0.0.1/staging/api/v1"  # use 127.0.0.1 on RED, not RED IP

def req(method, path, data=None, token=None):
    url = f"{BASE}/{path}"
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    body = json.dumps(data).encode() if data else None
    try:
        rq = urllib.request.Request(url, data=body, headers=headers,
                                     method=method)
        with urllib.request.urlopen(rq, timeout=10) as resp:
            raw = resp.read()
            try:
                return resp.status, json.loads(raw)
            except json.JSONDecodeError:
                return resp.status, raw.decode("utf-8", errors="replace")
    except urllib.error.HTTPError as e:
        raw = e.read()
        try:
            return e.code, json.loads(raw)
        except json.JSONDecodeError:
            return e.code, raw.decode("utf-8", errors="replace")
    except Exception as e:
        return 0, str(e)

def check(label, code, expected, body):
    ok = code == expected
    status = "PASS" if ok else "FAIL"
    preview = ""
    if isinstance(body, dict):
        preview = json.dumps(body)[:200]
    elif isinstance(body, str) and len(body) > 200:
        preview = body[:200]
    else:
        preview = str(body)[:200]
    print(f"  [{status}] {label}: HTTP {code} (expected {expected})")
    if not ok:
        print(f"         Body: {preview}")
    return ok

# --- Run tests ---
passed = failed = 0

# 1. Register
code, d = req("POST", "register",
              {"email": "e2e_test_xyz@example.com", "password": "test1234",
               "name": "E2E Tester"})
if check("Register", code, 201, d):
    passed += 1
    token = d.get("token", "")
else:
    failed += 1
    token = ""
    print("  -> Cannot continue without registration token")
    sys.exit(1)

if token:
    time.sleep(0.5)

# 2. Login (with same credentials)
code, d = req("POST", "login",
              {"email": "e2e_test_xyz@example.com", "password": "test1234"})
if check("Login", code, 200, d):
    passed += 1
    login_token = d.get("token", "")
else:
    failed += 1

# 3. GetMe (requires JWT)
code, d = req("GET", "users/me", token=token)
if check("GetMe (JWT)", code, 200, d):
    passed += 1
else:
    failed += 1

# ... continue for all endpoints ...

print(f"\nResults: {passed} passed, {failed} failed out of {passed+failed}")
sys.exit(0 if failed == 0 else 1)
```

## Key decisions baked into the template

- **Use `127.0.0.1` on RED, not RED's IP.** The script runs on RED itself;
  `127.0.0.1` hits the host nginx directly without network hops.
- **`urllib.request` not `requests`.** It's stdlib; no pip install needed on
  RED (which may not have pip/python-requests).
- **`json.loads` with fallback.** API endpoints that return HTML (e.g. product
  by slug returning the product page) will fail JSON parsing — catch it and
  report the raw body instead of crashing.
- **Sleep between dependent calls.** A small `time.sleep(0.5)` between
  register and login avoids any DB replication delay on RED.
- **Exit code 1 on any failure.** Makes the SSH command's exit code reflect
  test success — useful for CI gates.

## Running it

```bash
# Copy to RED
scp /root/project/backend/staging_e2e.py root@194.5.206.106:/tmp/staging_e2e.py

# Execute on RED
ssh root@194.5.206.106 "python3 /tmp/staging_e2e.py"
```

## Interpreting results

- **PASS on auth endpoints (register/login/getMe) but FAIL on feature
  endpoints (cart/wishlist/etc) with 401:** the feature routes lack JWT
  middleware. Fix: wrap them in a `JwtAuthMiddleware()` group in main.go.
- **404 on specific endpoints:** route not registered. Check main.go and the
  handler's `Register*Routes` function.
- **500 on an endpoint:** handler has a runtime bug. Check backend pod logs:
  `k3s kubectl logs -n staging -l app=backend --tail=30`.
- **Product by slug returns 200 but body is HTML/None:** the handler is
  returning the wrong content type or the slug lookup is failing silently.

## When to add new endpoints to the script

Every new feature handler added to the codebase should get a test block in
this script before the deployment is considered verified. Keep the script
under ~300 lines; split into multiple files if it grows beyond that.

## Pitfall — script runs from wrong directory

If you run `python3 staging_e2e.py` without `scp`ing it to RED first, you're
running it on BLUE against BLUE's localhost — which has no backend. Always
SCP first, then SSH-execute on RED.

## Pitfall — hardcoded test email collision

Use a unique suffix per run (e.g. timestamp or random string) in the test
email to avoid "email already exists" errors on re-runs. Clean up test users
afterward via admin API if available.
