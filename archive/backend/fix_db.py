#!/usr/bin/env python3
"""Fix double-quoted mock strings -> backtick, fix escaping."""
import re

with open("handlers/all_features_test.go", "r") as f:
    lines = f.readlines()

out = []
for line in lines:
    # Convert " to ` for mock SQL strings
    # mock.ExpectQuery("SQL") -> mock.ExpectQuery(`SQL`)
    # But we need to handle the content too
    
    # Check if this line has a mock double-quoted string
    m = re.search(r'(mock\.(?:ExpectQuery|ExpectExec)\(")(.*)("(\)\.?))', line)
    if m:
        prefix = m.group(1)[:-1]  # mock.ExpectQuery(`
        content = m.group(2)      # SQL content (with Go escapes)
        suffix = m.group(3)       # ").
        
        # The content has Go double-quoted escapes. 
        # In the source file, \\$ is backslash + dollar (Go interprets \\ as \)
        # We want the backtick to contain \$ (literal backslash + dollar)
        # In the file bytes, the content has: possibly \\ then $N
        # We need to convert to: \$N (single backslash before dollar)
        
        # Decode Go string escapes to get actual content
        # Then re-encode for backtick (which needs no escaping for \$)
        # Actually for backtick, we just need the raw bytes that represent
        # the regex pattern \$1 (backslash + dollar + digit)
        
        # The content in the file currently has sequences like:
        # \\$1 = Go interprets as \$1 (backslash + dollar + 1) -- actually no
        # In Go double-quoted: \\ = \, \$ = literal $ (but \$ is invalid!)
        # 
        # Let me just handle it empirically:
        # The file has \\\\\ (or \\\) followed by $N
        # We want \$N in the backtick
        
        # Replace: one or more backslashes followed by $N -> \$N
        # But only if there are backslashes (indicating it was escaped in double-quoted)
        content = re.sub(r'(\\+)(\$[\d])', r'\\\2', content)
        # This replaces any number of backslashes before $digit with single backslash
        
        newline = prefix + '`' + content + '`' + suffix
        out.append(newline)
    else:
        out.append(line)

with open("handlers/all_features_test.go", "w") as f:
    f.writelines(out)

print("Fixed")
# Verify
with open("handlers/all_features_test.go", "r") as f:
    vl = f.readlines()
for i in [13, 26, 41, 70, 74, 77, 122, 127]:
    if i < len(vl):
        print(f"L{i+1}: {vl[i].rstrip()[:130]}")
