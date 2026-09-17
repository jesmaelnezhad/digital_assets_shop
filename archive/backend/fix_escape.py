#!/usr/bin/env python3
"""Fix all_features_test.go: correct backslash-dollar escaping."""
PATH = "/root/project/backend/handlers/all_features_test.go"
with open(PATH, "rb") as f:
    raw = f.read()

# Fix: replace 3-byte sequence \ \ $ with 2-byte \ $
# ONLY inside backtick-delimited raw strings
result = bytearray()
i = 0
depth = 0  # odd = inside raw string
bt = ord('`')  # 96
bs = ord('\\')  # 92
dl = ord('$')  # 36

while i < len(raw):
    b = raw[i]
    if b == bt:  # backtick
        depth += 1
        result.append(bt)
        i += 1
        while i < len(raw) and raw[i] != bt:
            if depth % 2 == 1 and i + 2 < len(raw) and raw[i] == bs and raw[i+1] == bs and raw[i+2] == dl:
                # Found \\$ inside raw string -> replace with \$
                result.append(bs)
                result.append(dl)
                i += 3
            else:
                result.append(raw[i])
                i += 1
        if i < len(raw):
            result.append(bt)
            i += 1
    else:
        result.append(b)
        i += 1

with open(PATH, "wb") as f:
    f.write(result)

print("Fixed backslash-dollar escaping in raw strings")
