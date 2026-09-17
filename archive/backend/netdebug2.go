package main

import (
	"fmt"
	"regexp"
)

func main() {
	// The SQL the handler sends
	handlerSQL := "SELECT p.id, p.title, p.slug, p.description, p.category_id, COALESCE(c.name, ''), p.price_usd,\n       p.asset_path, p.asset_hash, p.status, p.download_count_limit,\n       p.max_downloads_per_user, p.file_size_bytes, p.file_mime_type,\n       p.created_at, p.updated_at\nFROM products p\nLEFT JOIN categories c ON p.category_id = c.id\nWHERE p.status = $1\nORDER BY p.created_at DESC\nLIMIT $2 OFFSET $3"

	// Pattern from file (with \\s+ = Go sees \s+)
	pattern := `(?s)SELECT\s+p\.id,\s+p\.title,\s+p\.slug,\s+p\.description,\s+p\.category_id,\s+COALESCE\(c\.name,\s+'',\s+\)\s+p\.price_usd,\s+p\.asset_path,\s+p\.asset_hash,\s+p\.status,\s+p\.download_count_limit,\s+p\.max_downloads_per_user,\s+p\.file_size_bytes,\s+p\.file_mime_type,\s+p\.created_at,\s+p\.updated_at\s+FROM\s+products\s+p\s+LEFT\s+JOIN\s+categories\s+c\s+ON\s+p\.category_id\s*=\s*c\.id\s+WHERE\s+p\.status\s*=\s*\$1\s+ORDER\s+BY\s+p\.created_at\s+DESC\s+LIMIT\s+\$2\s+OFFSET\s+\$3`

	re := regexp.MustCompile(pattern)
	fmt.Printf("Match with \\s+: %v\n", re.MatchString(handlerSQL))

	// Also test with the exact pattern string from the file
	// In the file: \\s -> Go sees \s, \\$ -> Go sees \$, etc.
	// But in Go source code string literal, \\\\s means \\s (two backslashes + s) in the string
	// Go regex then sees \\s which is literal backslash + s (NOT whitespace)
	// 
	// The file has \\s (two chars: backslash backslash s)
	// Go reads this as the string \s (backslash s) in the backtick
	// Go regex interprets \s as whitespace metacharacter
	// So \\s in the FILE = \s in Go regex = whitespace. Correct!
	//
	// But in my Go debug program, when I write the pattern as a Go string literal:
	// `(?s)SELECT\s+...`  — this is a raw string literal (backtick)
	// In raw string, \s = backslash + s (2 chars). Go regex sees \s = whitespace. Correct!
	//
	// But when I write: 
	// "(?s)SELECT\\s+..." — this is a regular string literal (double quote)
	// In regular string, \\s = backslash + s (2 chars). Same result.
	//
	// So the pattern should work. But it doesn't match. Why?
	//
	// Let me check character by character...
	
	fmt.Println("\n=== Character analysis ===")
	fmt.Printf("Handler SQL length: %d\n", len(handlerSQL))
	fmt.Printf("Pattern length: %d\n", len(pattern))
	
	// Check first few chars
	fmt.Printf("Handler SQL[:50]: %q\n", handlerSQL[:50])
	fmt.Printf("Pattern[:50]: %q\n", pattern[:50])
	
	// Try compiling the pattern differently
	// The pattern from file has \\s which in Go string becomes \s
	// Let me construct it the same way the file does
	filePattern := `(?s)SELECT\s+p\.id,\s+p\.title,\s+p\.slug,\s+p\.description,\s+p\.category_id,\s+COALESCE\(c\.name,\s+'',\s+\)\s+p\.price_usd,\s+p\.asset_path,\s+p\.asset_hash,\s+p\.status,\s+p\.download_count_limit,\s+p\.max_downloads_per_user,\s+p\.file_size_bytes,\s+p\.file_mime_type,\s+p\.created_at,\s+p\.updated_at\s+FROM\s+products\s+p\s+LEFT\s+JOIN\s+categories\s+c\s+ON\s+p\.category_id\s*=\s*c\.id\s+WHERE\s+p\.status\s*=\s*\$1\s+ORDER\s+BY\s+p\.created_at\s+DESC\s+LIMIT\s+\$2\s+OFFSET\s+\$3`
	
	re2 := regexp.MustCompile(filePattern)
	fmt.Printf("\nFile-style pattern match: %v\n", re2.MatchString(handlerSQL))
	
	// Hmm, same result. Let me check if COALESCE matching is the issue
	// Handler: COALESCE(c.name, '')  — two single quotes
	// Pattern: COALESCE\(c\.name,\s+'',\s+\) — matches COALESCE(c.name,  '',  )
	// The \s+ between ',' and '' might not match if there's no space
	
	// Check: in handler SQL, what's between "name," and "''"?
	idx := strings.Index(handlerSQL, "COALESCE")
	if idx >= 0:
		fmt.Printf("\nCOALESCE context: %q\n", handlerSQL[idx:idx+50])
	
	// Try pattern without the COALESCE whitespace requirements
	altPattern := `(?s)SELECT\s+p\.id,\s+p\.title,\s+p\.slug,\s+p\.description,\s+p\.category_id,\s+COALESCE\(c\.name,\s*''\s*\)\s+p\.price_usd,\s+p\.asset_path,\s+p\.asset_hash,\s+p\.status,\s+p\.download_count_limit,\s+p\.max_downloads_per_user,\s+p\.file_size_bytes,\s+p\.file_mime_type,\s+p\.created_at,\s+p\.updated_at\s+FROM\s+products\s+p\s+LEFT\s+JOIN\s+categories\s+c\s+ON\s+p\.category_id\s*=\s*c\.id\s+WHERE\s+p\.status\s*=\s*\$1\s+ORDER\s+BY\s+p\.created_at\s+DESC\s+LIMIT\s+\$2\s+OFFSET\s+\$3`
	re3 := regexp.MustCompile(altPattern)
	fmt.Printf("\nAlt pattern (with \\s* for COALESCE args) match: %v\n", re3.MatchString(handlerSQL))
	
	// Try even simpler: skip COALESCE entirely
	simple := `(?s)SELECT\s+p\.id,\s+p\.title,\s+p\.slug,\s+p\.description,\s+p\.category_id,\s+.*p\.price_usd.*FROM\s+products\s+p\s+LEFT\s+JOIN\s+categories\s+c\s+ON\s+p\.category_id\s*=\s*c\.id\s+WHERE\s+p\.status\s*=\s*\$1\s+ORDER\s+BY\s+p\.created_at\s+DESC\s+LIMIT\s+\$2\s+OFFSET\s+\$3`
	re4 := regexp.MustCompile(simple)
	fmt.Printf("Simple pattern match: %v\n", re4.MatchString(handlerSQL))
}
