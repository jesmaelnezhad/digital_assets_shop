#!/usr/bin/env python3
"""Staging E2E probe - test live staging endpoints."""
import subprocess, json, sys

BASE = "http://194.5.206.106:30083"
PASS, FAIL = 0, 0

def check(name, url, expected_code=200, method="GET", data=None, headers=None):
    global PASS, FAIL
    cmd = ["curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "--connect-timeout", "3"]
    if headers:
        for k, v in headers.items():
            cmd += ["-H", f"{k}: {v}"]
    if method == "POST" or data:
        cmd += ["-X", method if method == "POST" else "GET"]
    if data:
        cmd += ["-d", data]
    cmd.append(url)
    
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=5)
        code = int(result.stdout.strip()) if result.stdout.strip().isdigit() else 0
        status = "PASS" if code == expected_code else "FAIL"
        if status == "PASS":
            PASS += 1
        else:
            FAIL += 1
        print(f"  [{status}] {name}: HTTP {code} (expected {expected_code})")
    except Exception as e:
        FAIL += 1
        print(f"  [FAIL] {name}: ERROR {e}")

print("=== Staging E2E Probe ===\n")

# Public endpoints
print("--- Public API ---")
check("Health", f"{BASE}/health", 200)
check("Products list", f"{BASE}/api/v1/products?limit=1", 200)
check("Product categories", f"{BASE}/api/v1/categories", 200)
check("Community feed", f"{BASE}/api/v1/community/feed?limit=1", 200)

# Auth endpoints
print("\n--- Auth ---")
check("Register", f"{BASE}/api/v1/register", 201, "POST", 
      '{"email":"test-e2e@test.com","password":"test1234","name":"E2E Test"}')
check("Login", f"{BASE}/api/v1/login", 200, "POST",
      '{"email":"test-e2e@test.com","password":"test1234"}')

# Feature endpoints  
print("\n--- Features ---")
check("Cart (no auth)", f"{BASE}/api/v1/cart", 401)
check("Wishlist (no auth)", f"{BASE}/api/v1/wishlist", 401)
check("Exchange rates", f"{BASE}/api/v1/exchange-rates", 200)
check("Settings", f"{BASE}/api/v1/settings/site_name", 200)

# Admin endpoints
print("\n--- Admin ---")
check("Admin users", f"{BASE}/api/v1/admin/users", 401)
check("Admin settings", f"{BASE}/api/v1/admin/settings", 401)

print(f"\n=== Result: {PASS} PASS, {FAIL} FAIL ===")
sys.exit(0 if FAIL == 0 else 1)
