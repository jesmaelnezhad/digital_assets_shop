package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	data, _ := os.ReadFile("handlers/all_features_test.go")

	// Fix: products mock line 125 - replace the regex pattern
	// The handler sends multi-line SQL with tabs and newlines
	// Go regex: \s matches whitespace including \n
	// We need the pattern to match the full query
	
	// Build the correct regex pattern
	// Use \s+ for whitespace (matches newlines too)
	// Escape dots, dollars, parens
	pattern := `(?s)SELECT\s+p\.id,\s+p\.title,\s+p\.slug,\s+p\.description,\s+p\.category_id,\s+COALESCE\(c\.name,\s+'',\s+\)\s+p\.price_usd,\s+p\.asset_path,\s+p\.asset_hash,\s+p\.status,\s+p\.download_count_limit,\s+p\.max_downloads_per_user,\s+p\.file_size_bytes,\s+p\.file_mime_type,\s+p\.created_at,\s+p\.updated_at\s+FROM\s+products\s+p\s+LEFT\s+JOIN\s+categories\s+c\s+ON\s+p\.category_id\s*=\s*c\.id\s+WHERE\s+p\.status\s*=\s*\$1\s+ORDER\s+BY\s+p\.created_at\s+DESC\s+LIMIT\s+\$2\s+OFFSET\s+\$3`

	// Also fix line 123 COUNT pattern  
	countPattern := `SELECT COUNT(*) FROM products p LEFT JOIN categories c ON p.category_id = c.id WHERE p.status = $1`

	// Find and replace the product list mock line
	lines := strings.Split(string(data), "\n")
	
	for i, line := range lines {
		if strings.Contains(line, "SELECT p.id, p.title.*") && strings.Contains(line, "mock.ExpectQuery") {
			// Replace with correct pattern
			lines[i] = fmt.Sprintf("\tmock.ExpectQuery(`%s`).", pattern)
			fmt.Printf("Fixed line %d (product list)\n", i+1)
		}
		if strings.Contains(line, "SELECT COUNT\\(\\*\\) FROM products.*") && strings.Contains(line, "mock.ExpectQuery") {
			// Replace with correct COUNT pattern
			lines[i] = fmt.Sprintf("\tmock.ExpectQuery(`%s`).", countPattern)
			fmt.Printf("Fixed line %d (COUNT)\n", i+1)
		}
	}

	result := strings.Join(lines, "\n")
	os.WriteFile("handlers/all_features_test.go", []byte(result), 0644)
	fmt.Println("Done writing file")

	// Verify
	data2, _ := os.ReadFile("handlers/all_features_test.go")
	if strings.Contains(string(data2), pattern) {
		fmt.Println("Pattern found in file - OK")
	} else {
		fmt.Println("Pattern NOT found - ERROR")
	}
	
	// Build and test
	os.Chdir("/root/project/backend")
	fmt.Println("Building...")
	build := execCommand("go", "build", "./handlers/")
	fmt.Printf("Build: exit=%d\n", build.ExitCode())
	
	fmt.Println("Testing...")
	test := execCommand("go", "test", "./handlers/", "-v", "-count=1", "-run", "TestProducts_GetProducts$")
	fmt.Println(test.Stdout())
}

func execCommand(name string, args ...string) *execResult {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return &execResult{
		ExitCode: 0,
		Stdout:   string(out),
	}
}

type execResult struct {
	ExitCode int
	Stdout   string
}
