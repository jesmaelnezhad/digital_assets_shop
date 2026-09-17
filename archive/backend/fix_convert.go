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

	// Convert all mock.ExpectQuery("...") and mock.ExpectExec("...") to backtick
	// with proper regex escaping

	re := regexp.MustCompile(`(mock\.(?:ExpectQuery|ExpectExec)\()(".*?")\)\.`)

	fixed := re.ReplaceAllStringFunc(string(data), func(match string) string {
		sub := re.FindStringSubmatch(match)
		if len(sub) < 4 {
			return match
		}
		prefix := sub[1]  // mock.ExpectQuery(
		quoted := sub[2]  // "SQL..."
		suffix := sub[3]  // ).

		// Use Go's strconv.Unquote to interpret Go double-quoted escapes
		unquoted, err := strconv.Unquote(quoted)
		if err != nil {
			// If unquote fails (illegal escapes like \$), treat the content
			// between quotes as raw and just wrap in backtick
			inner := quoted[1 : len(quoted)-1]
			// Normalize: reduce \\\\\$ to \$ (regex escaped dollar)
			inner = strings.ReplaceAll(inner, "\\$", "\$")
			inner = strings.ReplaceAll(inner, "\\\\", "\\")
			return prefix + "`" + inner + "`" + suffix
		}

		// unquoted is the actual SQL string the handler sends
		// For regex matching in backtick, escape regex metacharacters
		// $ -> \$ (regex escaped dollar)
		// But keep .* wildcards as-is (intentional)
		// Escape dots in column names: . -> \.
		
		// Strategy: escape $, (, ) for regex, keep . for column names
		// but the column names already have \. in many patterns
		
		// Escape $ as \$ for regex
		escaped := strings.ReplaceAll(unquoted, "$", "\\$")
		
		return prefix + "`" + escaped + "`" + suffix
	})

	os.WriteFile("handlers/all_features_test.go", []byte(fixed), 0644)
	fmt.Println("Converted mock SQL to backtick with regex escaping")

	// Verify
	data2, _ := os.ReadFile("handlers/all_features_test.go")
	bt := strings.Count(string(data2), "mock.ExpectQuery(`")
	dq := strings.Count(string(data2), `mock.ExpectQuery("`)
	fmt.Printf("Backtick: %d, Double-quoted: %d\n", bt, dq)

	// Show line 27
	lines := strings.Split(string(data2), "\n")
	if len(lines) > 26 {
		fmt.Printf("\nL27: %s\n", lines[26])
	}
}
