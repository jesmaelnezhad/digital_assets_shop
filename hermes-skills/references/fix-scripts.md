# Fix Scripts for all_features_test.go

Catalogue of fix scripts for the test file repair workflow.

## Stable: fix_v5.py (69 lines)
Converts 77 double-quoted mock SQL strings to backtick raw strings.
Produces: BUILD:0, ~19 PASS, ~45 FAIL (SQL mismatches remain).
Usage: cp handlers/all_features_test.go.bak handlers/all_features_test.go && python3 fix_v5.py

## Byte-level fixes (after v5)
Fix COUNT(\\* -> COUNT(\* via bytes([0x5c,0x5c,0x2a]) -> bytes([0x5c,0x2a]).
Fix \\$ -> \$ via bytes([0x5c,0x5c,0x24]) -> bytes([0x5c,0x24]).
Result: BUILD:0, 19 PASS, 45 FAIL.

## Unreliable: all full-file Python regex rewrites
fix_all.py, fix_all_mocks.py, fix_remaining.py, fix_batch.py, fix_sql_mocks.py, fix_all_failures.py, fix_mocks_comprehensive.py, fix_mocks_batch.py, fix_mocks_final.py, fix_comprehensive.py, fix_definitive.py, fix_v6/v7/v8.py, fix_*.go.
Each caused file corruption requiring restore from backup. Pattern: regex replacements across newlines split backtick strings, miscount parens, cascade escaping.

## Recovery
cp handlers/all_features_test.go.bak handlers/all_features_test.go && python3 fix_v5.py && apply byte-level fixes && go build && go test. Always returns to 19 PASS / 45 FAIL baseline.
