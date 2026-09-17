package main

import (
	"fmt"
	"regexp"
	"strings"
)

func main() {
	// The SQL the handler sends (as fmt.Sprintf would produce with \n\t formatting)
	handlerSQL := `SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd,
       p.asset_path, p.asset_hash, p.status, p.download_count_limit,
       p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type,
       p.created_at, p.updated_at
FROM products p
LEFT JOIN categories c ON p.category_id = c.id
WHERE p.status = $1
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3`

	fmt.Println("=== Handler SQL (raw) ===")
	fmt.Printf("%q\n\n", handlerSQL)

	// Pattern we're using in the mock
	pattern := `(?s)SELECT\s+p\.id,\s+p\.title,\s+p\.slug,\s+p\.description,\s+p\.category_id,\s+COALESCE\(c\.name,\s+'',\s+\)\s+p\.price_usd,\s+p\.asset_path,\s+p\.asset_hash,\s+p\.status,\s+p\.download_count_limit,\s+p\.max_downloads_per_user,\s+p\.file_size_bytes,\s+p\.file_mime_type,\s+p\.created_at,\s+p\.updated_at\s+FROM\s+products\s+p\s+LEFT\s+JOIN\s+categories\s+c\s+ON\s+p\.category_id\s*=\s*c\.id\s+WHERE\s+p\.status\s*=\s*\$1\s+ORDER\s+BY\s+p\.created_at\s+DESC\s+LIMIT\s+\$2\s+OFFSET\s+\$3`

	fmt.Println("=== Pattern ===")
	fmt.Println(pattern)
	fmt.Println()

	re := regexp.MustCompile(pattern)
	fmt.Printf("Match: %v\n\n", re.MatchString(handlerSQL))

	// Try with literal whitespace (the handler uses tabs for indentation)
	// The handler SQL has: \n\t\t (newline + 2 tabs) for continuation lines
	handlerSQL2 := strings.ReplaceAll(handlerSQL, "       ", "\t\t")
	handlerSQL2 = strings.ReplaceAll(handlerSQL2, "  ", "\t")
	
	fmt.Println("=== Handler SQL (with tabs) ===")
	fmt.Printf("%q\n\n", handlerSQL2)
	fmt.Printf("Match with tabs: %v\n\n", re.MatchString(handlerSQL2))

	// Try simpler pattern
	simple := `(?s)SELECT.*?FROM products p.*?LEFT JOIN.*?WHERE.*ORDER BY.*LIMIT.*OFFSET.*`
	re2 := regexp.MustCompile(simple)
	fmt.Printf("Simple match: %v\n", re2.MatchString(handlerSQL))
	fmt.Printf("Simple match (tabs): %v\n", re2.MatchString(handlerSQL2))

	// Try pattern with literal \t 
	litTabs := `(?s)SELECT\s+p\.id,\s+p\.title,\s+p\.slug,\s+p\.description,\s+p\.category_id,\s+COALESCE\(c\.name,\s+'',\s+\)\s+p\.price_usd,\s+p\.asset_path,\s+p\.asset_hash,\s+p\.status,\s+p\.download_count_limit,\s+p\.max_downloads_per_user,\s+p\.file_size_bytes,\s+p\.file_mime_type,\s+p\.created_at,\s+p\.updated_at\s+FROM\s+products\s+p\s+LEFT\s+JOIN\s+categories\s+c\s+ON\s+p\.category_id\s*=\s*c\.id\s+WHERE\s+p\.status\s*=\s*\$1\s+ORDER\s+BY\s+p\.created_at\s+DESC\s+LIMIT\s+\$2\s+OFFSET\s+\$3`
	re3 := regexp.MustCompile(litTabs)
	fmt.Printf("\nLiteral tabs match: %v\n", re3.MatchString(handlerSQL))
	fmt.Printf("Literal tabs match (with real tabs): %v\n", re3.MatchString(handlerSQL2))
}
