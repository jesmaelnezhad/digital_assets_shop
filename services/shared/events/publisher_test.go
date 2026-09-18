package events

import "testing"

func TestSanitizeForNotifyDoublesQuotes(t *testing.T) {
	if got := sanitizeForNotify("O'Brien"); got != "O''Brien" {
		t.Fatalf("got %q", got)
	}
	if got := sanitizeForNotify("plain"); got != "plain" {
		t.Fatalf("got %q", got)
	}
}
