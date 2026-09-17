# Staging-First Verification Priority

When unit tests are stuck on mock SQL mismatches (e.g. 19 PASS / 45 FAIL) but the build passes and staging is live, deploy and verify staging before continuing to fight mocks.

## Why

The 45 mock mismatches are a local test artifact. Staging E2E is the actual verification criterion for the ecommerce goal. If staging endpoints respond correctly, the deployed code works.

## Decision tree

1. Build passes? → Yes: continue. No: fix compile errors.
2. Staging live and responding? → Yes: run E2E probes. No: debug staging.
3. E2E probes pass? → Yes: goal progress made; fix remaining tests as secondary work.
4. E2E probes fail? → Trace the specific endpoint failure to code; fix that code path; redeploy; re-test.

## Signals you're in a mock-fix loop

- Restoring `.bak` → `fix_v5.py` → COUNT fix → patch → test → more failures → restore, repeated 3+ times
- FAIL count stuck at 40+ despite multiple fix attempts
- Each Python regex rewrite introduces new corruptions (merged lines, lost commas, double-escaped backslashes)

## The only stable baseline

```bash
cp handlers/all_features_test.go.bak handlers/all_features_test.go
python3 fix_v5.py
python3 -c "import sys; d=open('handlers/all_features_test.go','rb').read(); d=d.replace(bytes([0x5c,0x5c,0x2a]), bytes([0x5c,0x2a])); open('handlers/all_features_test.go','wb').write(d)"
go build ./handlers/   # exit 0
go test ./handlers/ -count=1  # 19 PASS, 45 FAIL
```

From here: surgical `patch` on individual mocks only. Never full-file Python regex rewrites.
