package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/DATA-DOG/sqlmock"
	"github.com/gin-gonic/gin"
	"time"
)

// Minimal reproduction: run the GetProducts handler and capture the SQL
func main() {
	mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}
	defer mock.Close()

	// Set up mock expectations
	mock.ExpectQuery("SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// The regex pattern we plan to use
	pattern := `(?s)SELECT\s+p\.id,\s+p\.title,\s+p\.slug,\s+p\.description,\s+p\.category_id,\s+COALESCE\(c\.name,\s+'',\s+\)\s+p\.price_usd,\s+p\.asset_path,\s+p\.asset_hash,\s+p\.status,\s+p\.download_count_limit,\s+p\.max_downloads_per_user,\s+p\.file_size_bytes,\s+p\.file_mime_type,\s+p\.created_at,\s+p\.updated_at\s+FROM\s+products\s+p\s+LEFT\s+JOIN\s+categories\s+c\s+ON\s+p\.category_id\s*=\s*c\.id\s+WHERE\s+p\.status\s*=\s*\$1\s+ORDER\s+BY\s+p\.created_at\s+DESC\s+LIMIT\s+\$2\s+OFFSET\s+\$3`

	// Test if the regex matches the expected SQL
	re := regexp.MustCompile(pattern)
	
	// The SQL the handler would send (formatted as fmt.Sprintf would produce)
	handlerSQL := `SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd,
       p.asset_path, p.asset_hash, p.status, p.download_count_limit,
       p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type,
       p.created_at, p.updated_at
FROM products p
LEFT JOIN categories c ON p.category_id = c.id
WHERE p.status = $1
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3`

	fmt.Println("Handler SQL:")
	fmt.Println(handlerSQL)
	fmt.Println()
	
	match := re.MatchString(handlerSQL)
	fmt.Printf("Regex match: %v\n", match)
	
	if !match {
		// Try simpler pattern
		simple := `(?s)SELECT.*FROM products p.*LEFT JOIN.*WHERE.*ORDER BY.*LIMIT.*OFFSET.*`
		re2 := regexp.MustCompile(simple)
		fmt.Printf("Simple match: %v\n", re2.MatchString(handlerSQL))
		
		// Try without \s (use literal space and tab)
		lit := `SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd,.*FROM products p.*LEFT JOIN categories c ON p.category_id = c.id.*WHERE p.status = \$1.*ORDER BY p.created_at DESC.*LIMIT \$2 OFFSET \$3`
		re3 := regexp.MustCompile(lit)
		fmt.Printf("Literal match: %v\n", re3.MatchString(handlerSQL))
		
		// What if the handler SQL uses \t instead of spaces?
		withTabs := strings.ReplaceAll(handlerSQL, "       ", "\t\t")
		withTabs = strings.ReplaceAll(withTabs, "  ", "\t")
		fmt.Println("\nWith tabs:")
		fmt.Println(withTabs)
		fmt.Printf("Simple match (tabs): %v\n", re2.MatchString(withTabs))
		fmt.Printf("Literal match (tabs): %v\n", re3.MatchString(withTabs))
	}
}
