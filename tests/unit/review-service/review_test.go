package review_test

import "testing"

func TestRatingBounds(t *testing.T) {
	for _, r := range []int{1, 3, 5} {
		if !validRating(r) {
			t.Fatalf("%d should be valid", r)
		}
	}
	for _, r := range []int{0, 6, -1} {
		if validRating(r) {
			t.Fatalf("%d should be invalid", r)
		}
	}
}

func validRating(r int) bool { return r >= 1 && r <= 5 }
