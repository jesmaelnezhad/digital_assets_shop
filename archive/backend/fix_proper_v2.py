#!/usr/bin/env python3
"""
Clean fix: read backup, properly interpret Go double-quoted escape sequences,
write correct backtick raw strings. One shot.
"""
import re

with open("handlers/all_features_test.go.bak", "rb") as f:
    data = f.read()

lines = data.split(b'\n')
out = []

def interpret_go_string(sq_bytes):
    """
    Interpret Go double-quoted string escape sequences.
    Input: raw bytes between the double quotes in Go source.
    Output: actual string value bytes.
    """
    result = bytearray()
    i = 0
    while i < len(sq_bytes):
        b = sq_bytes[i]
        if b == 0x5c:  # backslash - start of escape
            if i + 1 >= len(sq_bytes):
                result.append(b)  # trailing backslash, keep as-is
                break
            next_b = sq_bytes[i + 1]
            if next_b == 0x5c:  # \\
                result.append(0x5c)  # single backslash
                i += 2
            elif next_b == 0x22:  # "
                result.append(0x22)  # literal quote
                i += 2
            elif next_b == 0x6e:  # \n
                result.append(0x0a)  # newline
                i += 2
            elif next_b == 0x74:  # \t
                result.append(0x09)  # tab
                i += 2
            elif next_b == 0x72:  # \r
                result.append(0x0d)  # carriage return
                i += 2
            else:
                # Unknown escape: keep both bytes as-is (this handles \$ and \*)
                # In Go, unknown escapes are compile errors, but we'll preserve them
                result.append(b)
                result.append(next_b)
                i += 2
        else:
            result.append(b)
            i += 1
    return bytes(result)

for line in lines:
    # Match: whitespace + mock.ExpectQuery("SQL")  or  mock.ExpectExec("SQL") + optional trailing .
    m = re.match(rb'^(\s*mock\.(ExpectQuery|ExpectExec)\()"(.*)"(\)\.?\s*)$', line)
    if m:
        prefix = m.group(1)  # e.g. b'\tmock.ExpectQuery(' 
        method = m.group(2)  # b'ExpectQuery' or b'ExpectExec'
        sql_raw = m.group(3) # Raw bytes between double quotes (Go source with escapes)
        suffix = m.group(4)  # closing "). or ")."
        
        # Interpret Go escape sequences to get actual string value
        sql_value = interpret_go_string(sql_raw)
        
        # Wrap in backtick
        new_line = prefix + b'`' + sql_value + b'`' + suffix
        out.append(new_line)
        continue
    
    out.append(line)

result = b'\n'.join(out)

# Summary
backtick = result.count(b'mock.ExpectQuery(`') + result.count(b'mock.ExpectExec(`')
double_q = result.count(b'mock.ExpectQuery("') + result.count(b'mock.ExpectExec("')
print(f"Backtick: {backtick}, Double-quoted remaining: {double_q}")

# Check key lines
lines_out = result.split(b'\n')
for idx in [13, 26, 122, 137]:
    if idx < len(lines_out):
        line = lines_out[idx]
        text = line.decode('latin-1')
        print(f"\nLine {idx+1}: {text[:120]}")
        # Find backtick content
        bt_start = line.find(b'`')
        if bt_start >= 0:
            bt_end = line.find(b'`', bt_start + 1)
            if bt_end >= 0:
                sql = line[bt_start+1:bt_end]
                print(f"  SQL ({len(sql)} bytes): {sql[:80]}")
                print(f"  Hex: {sql[:50].hex()}")
                if b'COUNT' in sql:
                    c = sql.find(b'COUNT')
                    s = sql[c:c+15]
                    print(f"  COUNT hex: {s.hex()}")
                    for i, b in enumerate(s):
                        ch = chr(b) if 32 <= b < 127 else '?'
                        print(f"    [{i}] 0x{b:02x} ({ch})")
                if b'$' in sql:
                    d_idx = sql.find(b'$')
                    ctx = sql[max(0,d_idx-5):d_idx+5]
                    print(f"  $ ctx hex: {ctx.hex()}")

with open("handlers/all_features_test.go", "wb") as f:
    f.write(result)
print(f"\nWrote {len(result)} bytes")
