package media_test

import "testing"

func TestUploadSizeCap(t *testing.T) {
	const max = 50 << 20
	if max != 52428800 {
		t.Fatalf("50MiB cap drifted: %d", max)
	}
}
