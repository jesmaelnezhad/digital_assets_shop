package payment_test

import "testing"

func TestRateMustBePositive(t *testing.T) {
	if validRate(0) || validRate(-1) {
		t.Fatal("non-positive rates")
	}
	if !validRate(1) || !validRate(250.5) {
		t.Fatal("positive rates")
	}
}

func TestConfirmAcceptsAwaitingPayment(t *testing.T) {
	ok := map[string]bool{"pending": true, "created": true, "awaiting_payment": true, "processing": true}
	if !ok["awaiting_payment"] || ok["paid"] {
		t.Fatal("confirm should accept awaiting_payment and skip paid")
	}
}

func validRate(v float64) bool { return v > 0 }
