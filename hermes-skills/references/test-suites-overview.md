# Test Structure and Running

## Three Test Suites

Pawradise has three distinct test suites:

### 1. Unit Tests (Go, sqlmock-based)
- Location: `/root/project/tests/unit/<service-name>/`
- Pattern: `*_test.go` files with mocked SQL (sqlmock)
- Run: `cd /root/project/tests && go test ./unit/...`
- Status: All passing (172 tests across 8 services)

### 2. Integration Tests (Go, cross-service)
- Location: `/root/project/tests/integration/`
- Pattern: `*_test.go` files testing workflows across multiple services
- Run: `cd /root/project/tests && go test ./integration/...`
- Status: All passing (28 tests)

### 3. E2E Tests (Node.js, live API)
- Location: `/root/project/tests/e2e-suite.js`
- Pattern: 104 tests using Node.js http module against live deployment
- Run: `cd /root/project/tests && API_HOST=194.5.206.106 API_PORT=30758 ADMIN_TOKEN=<token> node e2e-suite.js`
- Status: 77/104 passing (as of 2026-09-09)

## E2E Test Configuration

Environment variables:
- `API_HOST`: RED IP (194.5.206.106)
- `API_PORT`: NodePort for ingress-nginx (30758)
- `ADMIN_TOKEN`: Admin bearer token from secrets

## E2E Test Patterns

### Response Format Assertions
- `assertArray(data, 'field', 'context')` — checks if `data[field]` exists and is an array
- `assertField(data, 'field', 'context')` — checks if `data[field]` is not undefined/null
- `assertStatus(res, expected, 'context')` — checks HTTP status code

### Common E2E Failures
1. **Response key mismatch**: Handler returns `{"cart":[]}` but test expects `items` key
2. **Auth context not populated**: Handler gets `user_id:0` when middleware doesn't inject it
3. **Stale Docker image**: Pods run old code despite new deployment
4. **Missing routes**: Endpoint registered in code but not in ingress
5. **DB schema mismatch**: Handler queries column that doesn't exist in DB

### Running E2E Tests (Recommended)

```bash
cd /root/project/tests
API_HOST=194.5.206.106 API_PORT=30758 ADMIN_TOKEN=admin_secret_2026_prod \
  timeout 120 node e2e-suite.js 2>&1 | tail -80
```

Use `timeout` to prevent hanging on network issues.
