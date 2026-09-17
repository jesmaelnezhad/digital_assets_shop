#!/usr/bin/env python3
"""
Fix all broken mock chains in all_features_test.go.
Reads the file, finds mock.Expect* calls missing their .WithArgs/.WillReturnRows chains,
and fixes them by re-chaining properly.
"""
import re, os

test_file = "/root/project/backend/handlers/all_features_test.go"
with open(test_file) as f:
    content = f.read()

fixes = 0

# Fix 1: Line 129-130: Products_GetProducts - the 2nd mock chain is broken
# mock.ExpectQuery(GET_PRODUCTS_SQL).\n=mock.ExpectQuery(IMAGES_SQL).\n\t\tWithArgs(1)...
# should be:
# mock.ExpectQuery(GET_PRODUCTS_SQL).\n\t\tWithArgs(...)....
# mock.ExpectQuery(IMAGES_SQL).\n\t\tWithArgs(1)...
old = """mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = \\$1 ORDER BY p.created_at DESC LIMIT \\$2 OFFSET \\$3`).
	mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = \\$1`)."""

new = """mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = \\$1 ORDER BY p.created_at DESC LIMIT \\$2 OFFSET \\$3`).
		WithArgs("active", 12, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "description", "category_id", "category_name", "price_usd", "asset_path", "asset_hash", "status", "download_count_limit", "max_downloads_per_user", "file_size_bytes", "file_mime_type", "created_at", "updated_at"}).
			AddRow(1, "Test Product", "test-product", "A test product", nil, "Icons", "10.00", "", "", "active", 100, 5, nil, "", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)).
			AddRow(2, "Another Product", "another-product", "Another test", nil, "", "20.00", "", "", "active", 50, 5, nil, "", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))
	mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = \\$1`)."

if old in content:
    content = content.replace(old, new)
    fixes += 1
    print("FIXED: Products_GetProducts mock chain")
else:
    print("MISS: Products_GetProducts mock chain")

# Fix 2: Line 145-146: Products_GetProductBySlug - same issue
old2 = """mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.slug = \\$1 AND p.status = 'active'`).
	mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = \\$1`)."

if old2 in content:
    content = content.replace(old2, old2.replace(
        ".\n\t\tmock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = \\$1`).\n\t\tWithArgs(1).",
        ".\n\t\tWithArgs(\"test-product\").\n\t\tWillReturnRows(sqlmock.NewRows([]string{\"id\", \"title\", \"slug\", \"description\", \"category_id\", \"category_name\", \"price_usd\", \"asset_path\", \"asset_hash\", \"status\", \"download_count_limit\", \"max_downloads_per_user\", \"file_size_bytes\", \"file_mime_type\", \"created_at\", \"updated_at\"}).\n\t\t\tAddRow(1, \"Test Product\", \"test-product\", \"A test product\", nil, \"Icons\", \"10.00\", \"\", \"\", \"active\", 100, 5, nil, \"\", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))\n\tmock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = \\$1`).\n\t\tWithArgs(1)."
    ))
    fixes += 1
    print("FIXED: Products_GetProductBySlug mock chain")
else:
    print("MISS: Products_GetProductBySlug mock chain")

# Fix 3: Line 161-165: Products_SearchProducts - same issue
old3 = """mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = \\$1 AND (p.name LIKE \\$2 OR p.description LIKE \\$2 OR p.slug LIKE \\$2) ORDER BY p.created_at DESC LIMIT \\$3 OFFSET \\$4`).
		WithArgs("active", "%test%", 12, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "description", "category_id", "category_name", "price_usd", "asset_path", "asset_hash", "status", "download_count_limit", "max_downloads_per_user", "file_size_bytes", "file_mime_type", "created_at", "updated_at"}).
			AddRow(1, "Test Product", "test-product", "A test product", nil, "Icons", "10.00", "", "", "active", 100, 5, nil, "", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))
	mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = \\$1`).
\t\tWithArgs(1).
\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "url", "is_primary", "created_at"}).
\t\t\tAddRow(1, 1, "http://example.com/img.png", true, mockTime(2024, time.January, 1)))"""

new3 = """mock.ExpectQuery(`SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd, p.asset_path, p.asset_hash, p.status, p.download_count_limit, p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type, p.created_at, p.updated_at FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = \\$1 AND (p.name LIKE \\$2 OR p.description LIKE \\$2 OR p.slug LIKE \\$2) ORDER BY p.created_at DESC LIMIT \\$3 OFFSET \\$4`).
		WithArgs("active", "%test%", 12, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug", "description", "category_id", "category_name", "price_usd", "asset_path", "asset_hash", "status", "download_count_limit", "max_downloads_per_user", "file_size_bytes", "file_mime_type", "created_at", "updated_at"}).
			AddRow(1, "Test Product", "test-product", "A test product", nil, "Icons", "10.00", "", "", "active", 100, 5, nil, "", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))
	mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = \\$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "url", "is_primary", "created_at"}).
			AddRow(1, 1, "http://example.com/img.png", true, mockTime(2024, time.January, 1)))"""

if old3 in content:
    content = content.replace(old3, new3)
    fixes += 1
    print("FIXED: Products_SearchProducts mock chain")
else:
    # Try alternate
    old3b = """mock.ExpectQuery(`SELECT id, product_id, url, is_primary, created_at FROM product_images WHERE product_id = \\$1`).
\t\tWithArgs(1).
\t\tWillReturnRows(sqlmock.NewRows([]string{\"id\", \"product_id\", \"url\", \"is_primary\", \"created_at\"}).
\t\t\tAddRow(1, 1, \"http://example.com/img.png\", true, mockTime(2024, time.January, 1)))"""
    if old3b in content:
        print("  Already fixed?")
    else:
        print("MISS: Products_SearchProducts mock chain - checking content...")
        idx = content.find('WHERE p.status = \\\\$1 AND (p.name LIKE')
        if idx >= 0:
            print(repr(content[idx:idx+500]))

# Fix 4: Community_Feed - broken chain
old4 = """mock.ExpectQuery(`SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, up.avatar_url, COUNT\\(DISTINCT l.id\\) as count_likes, COUNT\\(DISTINCT cm.id\\) as count_comments, EXISTS\\(SELECT 1 FROM follows WHERE follower_id = \\$1 AND followee_id = cp.user_id\\) as following FROM community_posts cp JOIN users u ON u.id = cp.user_id LEFT JOIN user_profiles up ON up.user_id = u.id LEFT JOIN post_likes l ON l.post_id = cp.id LEFT JOIN post_comments cm ON cm.post_id = cp.id WHERE cp.user_id = \\$1 GROUP BY cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, up.avatar_url`).""

if old4 in content:
    print("  Community_Feed chain OK (just a question mark)")
else:
    # Check what's there
    idx = content.find('COUNT\\\\(DISTINCT l.id\\\\')
    if idx >= 0:
        end_idx = content.find('WillReturnRows', idx)
        if end_idx >= 0:
            print(f"  Community_Feed: found WillReturnRows at offset {end_idx}")
            print(repr(content[idx:end_idx+100]))
        else:
            print(f"  Community_Feed: MISSING WillReturnRows after offset {idx}")
            print(repr(content[idx:idx+300]))

# Fix 5: Community_Profile user_profiles - broken chain
old5 = """mock.ExpectQuery(`SELECT id, user_id, bio, avatar_url, wallet_address, created_at, updated_at FROM user_profiles WHERE user_id = \\$1`).
\t\tWithArgs(1).
\t\tWillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "bio", "avatar_url", "wallet_address", "created_at", "updated_at"}).
\t\t\tAddRow(1, 1, "Hello world", nil, "0x1234...", mockTime(2024, time.January, 1), mockTime(2024, time.January, 1)))"""

if old5 in content:
    print("  Community_Profile user_profiles: OK")
else:
    idx = content.find('SELECT id, user_id, bio, avatar_url, wallet_address, created_at, updated_at FROM user_profiles WHERE user_id = \\$1')
    if idx >= 0:
        print(f"  Community_Profile user_profiles: offset={idx}")
        print(repr(content[idx:idx+300]))

# Fix 6: Cart_AddItem - broken chain (stock query)
old6 = """mock.ExpectQuery(`SELECT stock_quantity, CAST(price_usd AS INTEGER) FROM products WHERE id = \\$1`).WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"stock_quantity", "price_usd"}).AddRow(100, 10))"""

if old6 in content:
    print("  Cart_AddItem stock: OK")
else:
    idx = content.find('SELECT stock_quantity, CAST(price_usd')
    if idx >= 0:
        print(f"  Cart_AddItem stock: offset={idx}")
        print(repr(content[idx:idx+300]))

# Fix 7: Cart_AddItem - INSERT INTO cart_items broken chain
old7 = """mock.ExpectExec(`INSERT INTO cart_items (cart_id, product_id, quantity) VALUES (\\\\$1, \\\\$2, \\\\$3) ON CONFLICT (cart_id, product_id) DO UPDATE SET quantity = \\\\$3, added_at = CURRENT_TIMESTAMP`).WithArgs(1, 1, 3).WillReturnResult(sqlmock.NewResult(1, 1))"""

if old7 in content:
    print("  Cart_AddItem INSERT: OK")
else:
    idx = content.find('INSERT INTO cart_items (cart_id, product_id, quantity)')
    if idx >= 0:
        print(f"  Cart_AddItem INSERT: offset={idx}")
        print(repr(content[idx:idx+300]))
    else:
        idx = content.find('INSERT INTO cart_items')
        if idx >= 0:
            print(f"  Cart_AddItem INSERT (short): offset={idx}")
            print(repr(content[idx:idx+200]))

with open(test_file, "w") as f:
    f.write(content)

print(f"\nTotal fixes: {fixes}")
