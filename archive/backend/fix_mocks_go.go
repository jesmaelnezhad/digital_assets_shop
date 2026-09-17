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

	re := regexp.MustCompile(`(mock\.(?:ExpectQuery|ExpectExec)\()(".*?")(\)\..*)`)

	fixed := re.ReplaceAllStringFunc(string(data), func(match string) string {
		sub := re.FindStringSubmatch(match)
		if len(sub) < 4 {
			return match
		}
		prefix := sub[1]
		quoted := sub[2]
		suffix := sub[3]

		unquoted, err := strconv.Unquote(quoted)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unquote error for %q: %v\n", quoted, err)
			return match
		}

		// Escape $ for regex matching in backtick
		escaped := strings.ReplaceAll(unquoted, "$", "\\$")
		return fmt.Sprintf("%s`%s`%s", prefix, escaped, suffix)
	})

	os.WriteFile("handlers/all_features_test.go", []byte(fixed), 0644)
	fmt.Println("Done")
}
