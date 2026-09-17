package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	data, _ := os.ReadFile("handlers/all_features_test.go")

	// Pattern: mock.ExpectQuery("SQL"). or mock.ExpectExec("SQL").
	// The SQL is in a Go double-quoted string with escape sequences.
	// We need to:
	// 1. Extract SQL between the double quotes
	// 2. Interpret Go escape sequences to get actual SQL text
	// 3. Add regex escaping ($ -> \$, etc.) for backtick
	// 4. Wrap in backtick

	re := regexp.MustCompile(`(mock\.(?:ExpectQuery|ExpectExec)\()(".*")(\)\.)`)

	fixed := re.ReplaceAllStringFunc(string(data), func(match string) string {
		sub := re.FindStringSubmatch(match)
		if len(sub) < 4 {
			return match
		}
		prefix := sub[1]
		quoted := sub[2] // includes outer " "
		suffix := sub[3]

		// Extract content between quotes
		inner := quoted[1 : len(quoted)-1]

		// The inner text has Go double-quoted escape sequences.
		// These are ILLEGAL in Go (e.g., \$ is not a valid escape),
		// but the file was written with them. We need to convert to
		// what the handler actually sends and then add regex escaping.
		//
		// The patterns in the file use:
		//   \\\\\\$1  = literal text \\\\\$1 (4 backslashes + dollar + 1 in raw bytes)
		//   \\\\*     = literal text \\\\* (4 backslashes + star)
		//   \\\\(     = literal text \\\\( (2 backslashes + paren)
		//
		// When Go interprets these as double-quoted strings:
		//   \\ -> \  (escape sequence)
		//   \$ -> $  (not a valid escape, so compile error)
		//
		// Since these ARE compile errors, the file can't be compiled as-is.
		// We need to treat them as raw text and produce the correct regex.

		// Normalize: The raw bytes have sequences like \\\\\$1 that need
		// to become \$1 in the backtick (for regex matching of literal $1).
		
		// Reduce multiple backslashes before special chars to single:
		// \\\\\$ -> \$ (4 backslashes + dollar -> 1 backslash + dollar)
		// \\\\* -> \* (4 backslashes + star -> 1 backslash + star)  
		// \\\( -> \( (2 backslashes + paren -> 1 backslash + paren)
		
		// Replace 4+ backslashes before $ with 1
		inner = strings.ReplaceAll(inner, "\\\\\\$", "\$")
		// Replace 4+ backslashes before * with 1
		inner = strings.ReplaceAll(inner, "\\\\\\*", "\*")
		// Replace 2+ backslashes before ( with 1
		inner = strings.ReplaceAll(inner, "\\\\(", "\(")
		// Replace 2+ backslashes before ' with 1
		inner = strings.ReplaceAll(inner, "\\\\\'", "\\\'")

		// Any remaining \\ -> \ (double backslash -> single in regex)
		// But be careful: \\\\ (4 backslashes) -> \\ (2 backslashes) shouldn't become \ (1)
		// Actually for regex, \\ in the pattern matches literal backslash.
		// The handler doesn't send literal backslashes in SQL, so we can simplify.
		
		// Replace any remaining double-backslash pairs with single
		inner = strings.ReplaceAll(inner, "\\\\", "\\")

		return prefix + "`" + inner + "`" + suffix
	})

	os.WriteFile("handlers/all_features_test.go", []byte(fixed), 0644)
	fmt.Println("Done converting")

	// Verify
	data2, _ := os.ReadFile("handlers/all_features_test.go")
	bt := strings.Count(string(data2), "mock.ExpectQuery(`")
	dq := strings.Count(string(data2), `mock.ExpectQuery("`)
	fmt.Printf("Backtick: %d, Double-quoted: %d\n", bt, dq)

	lines := strings.Split(string(data2), "\n")
	for _, idx := range []int{13, 26, 122, 124} {
		if idx < len(lines) {
			fmt.Printf("L%d: %s\n", idx+1, lines[idx])
		}
	}
}
