#!/usr/bin/env python3
"""
Definitive fix: read backup, convert ALL mock SQL from double-quoted to backtick,
fix all escaping correctly, write result. One shot, no incremental steps.
"""
import re

with open("handlers/all_features_test.go.bak", "rb") as f:
    data = f.read()

lines = data.split(b'\n')
out = []

for line in lines:
    # Match: optional whitespace + mock.ExpectQuery("SQL") or mock.ExpectExec("SQL") + optional .
    m = re.match(rb'^(\s*mock\.(ExpectQuery|ExpectExec)\()"([^"]*)"(\)\.?\s*)$', line)
    if m:
        prefix = m.group(1)   # e.g. b'\tmock.ExpectQuery('
        sql = m.group(3)      # SQL bytes inside double quotes
        suffix = m.group(4)   # closing "). or ").
        
        # Fix SQL escaping for backtick context
        # In double-quoted Go source, \\ = one literal backslash
        # In backtick, we want the literal characters as-is for sqlmock regex matching
        # 
        # The backup file inside double quotes contains:
        # - \\\\$1 (4 chars: \, \, \, \, $, 1 in source; Go interprets \\\\=$+$$ \$ = \$): result = \\\$1
        #   Wait no. In Go double-quoted source:
        #   \\\\  = Go interprets each \\ as \, so \\\\ becomes \\ (2 literal backslashes)
        #   \$    = invalid escape, Go error
        #   
        # Hmm, but the test was compiling before with "unknown escape sequence" error.
        # So the backup HAS invalid Go escapes. Let me check what's actually there.
        
        # From earlier hex analysis of line 27:
        # SQL ending bytes: 5c5c5c2431 = \, \, \, $, 1 (3 backslashes + $1)
        # In Go double-quoted source: \\ (1 backslash) + \\ (1 backslash) + \$ (ERROR)
        # 
        # For backtick context, we want the SQL to contain $1 literally (for sqlmock to match $1 placeholder)
        # Backtick: \$1 = literal backslash + $1. As regex for sqlmock: \$ matches literal $, 1 matches 1.
        # But the handler sends just $1 (no backslash). So sqlmock needs to match $1, not \$1.
        # 
        # In sqlmock, the regex pattern is matched against the actual SQL string.
        # Handler sends: "SELECT ... WHERE email = $1"
        # Mock regex:    "SELECT ... WHERE email = $1"  (regex: $ matches end of line!)
        # To match literal $: use \$ in regex. So mock needs: WHERE email = \$1
        # 
        # In backtick string: `WHERE email = \$1` = bytes: ..., =, space, \, $, 1
        # This is the regex pattern: "WHERE email = \$1" which matches "WHERE email = $1" literally.
        #
        # So we need the backtick SQL to contain: backslash, dollar, 1  (bytes: 5c 24 31)
        # Currently the double-quoted source has: 5c 5c 5c 24 31 (3 backslashes + $1)
        # Go interprets: \\ = \, \\ = \, \$ = error (invalid escape)
        # 
        # So the backup is ALREADY broken Go. The fix needs to produce valid Go that gives
        # the right regex pattern.
        #
        # For backtick: we want bytes 5c 24 31 (\ $ 1) for \$1 pattern
        # Current bytes in double-quoted source: 5c 5c 5c 24 31 (3 backslashes + $1)
        # 
        # Transformation: replace sequence of 2+ backslashes followed by $ with single backslash + $
        # i.e., 5c{2,} 24 -> 5c 24
        
        sql_fixed = sql
        
        # Replace sequences of 2+ backslashes before $ with single backslash
        # Match: backslash backslash ... backslash dollar
        # Replace with: backslash dollar
        sql_fixed = re.sub(rb'\\\+(\$ phosphine)', rb'\\\ phosphine', sql_fixed)  # wrong
        # Actually: re.sub(rb'\\{2,}\$', b'\\$', sql_fixed)
        sql_fixed = re.sub(rb'\\{2,}\$', b'\\$', sql_fixed)
        
        # For COUNT(\\\\*) -> COUNT(\\*)  
        # Current: COUNT + ( + \\\\* (2 backslashes + star) + )
        # In backtick we want: COUNT + ( + \\* (1 backslash + star) + )
        sql_fixed = re.sub(rb'\\\+(\*)', rb'\\\1', sql_fixed)
        
        # Also fix any other \\\\  (2 backslashes in source) -> \\ (1 backslash in backtick)
        # But ONLY when not before $ (already handled)  
        sql_fixed = re.sub(rb'\\{2,}', b'\\', sql_fixed)
        
        # Final check: ensure we have \$1 not just $1 for placeholder matching
        # Actually no - in backtick, \$1 = \, $, 1 which as regex matches literal $1
        # But if we removed too many backslashes, we might have just $1 which matches end-of-line
        # Let me check: do we need \$ or just $ in the regex?
        # sqlmock uses regex matching. $ in regex = end of line. \$ in regex = literal $.
        # So we NEED the backslash before $ for placeholders.
        # Our transformation 2+ backslashes -> 1 backslash preserves \$ correctly.
        
        new_line = prefix + b'`' + sql_fixed + b'`' + suffix
        out.append(new_line)
        continue
    
    out.append(line)

result = b'\n'.join(out)

# Summary
backtick_count = result.count(b'mock.ExpectQuery(`') + result.count(b'mock.ExpectExec(`')
double_count = len(re.findall(rb'mock\.(ExpectQuery|ExpectExec)\("', result))
print(f"Backtick mocks: {backtick_count}")
print(f"Double-quoted mocks remaining: {double_count}")

# Show key lines
lines_out = result.split(b'\n')
for idx in [13, 26, 122, 137]:
    if idx < len(lines_out):
        text = lines_out[idx].decode('latin-1')
        print(f"\nLine {idx+1}: {text[:120]}")
        if b'COUNT' in lines_out[idx]:
            c_idx = lines_out[idx].find(b'COUNT')
            snippet = lines_out[idx][c_idx:c_idx+15]
            print(f"  COUNT hex: {snippet.hex()}")
            for j, b in enumerate(snippet):
                c = chr(b) if 32 <= b < 127 else '?'
                print(f"    [{j}] 0x{b:02x} ({c})")

with open("handlers/all_features_test.go", "wb") as f:
    f.write(result)

print(f"\nWrote {len(result)} bytes to handlers/all_features_test.go")
