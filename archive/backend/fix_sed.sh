#!/bin/bash
# Directly fix each specific mock SQL pattern in all_features_test.go
# Based on exact strings found via grep

FILE="/root/project/backend/handlers/all_features_test.go"

# Fix double-escaped stars: \\\* → \*  (in Go raw strings)
# In the file: COUNT\(\*\\\\)  →  COUNT\(\*\)
sed -i 's/COUNT(\\*\\\\)/COUNT(\\*/g' "$FILE"

# Fix double-escaped dollars: \\\$ → \$
sed -i 's/\\\\\$/\\$/g' "$FILE"

# Fix product images query - add ORDER BY
sed -i 's/FROM product_images WHERE product_id = \\\$1`)/FROM product_images WHERE product_id = \\\$1 ORDER BY is_primary DESC, id`)/g' "$FILE"

# Fix community feed comment alias: c.id → cm.id
sed -i 's/COUNT(DISTINCT c\.id)/COUNT(DISTINCT cm.id)/g' "$FILE"

# Fix guest order CHECK - table name and columns
sed -i 's/SELECT id, email, status.*FROM guest_orders WHERE id = \\\$1/SELECT email, status, total_usd, crypto_chain FROM orders WHERE id = \\\$1 AND guest_order = true/g' "$FILE"

# Fix wishlist LIST query - replace .* wildcards
sed -i 's/SELECT w\.product_id, p\.title, p\.slug\.\*FROM wishlists w\.\*/SELECT p.id, p.title, p.slug, CAST(p.price_usd AS NUMERIC) as price_usd, p.stock_quantity, COALESCE(pi.url, '\'\'') as primary_image FROM wishlists w JOIN products p ON p.id = w.product_id LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.is_primary = true WHERE w.user_id = \\\$1 ORDER BY w.created_at DESC/g' "$FILE"

echo "Done with sed fixes"
