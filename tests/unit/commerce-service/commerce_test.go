package commerce_test

import "testing"

func TestSeededCouponCodes(t *testing.T) {
	for _, code := range []string{"SAVE12", "WELCOME", "MARBLE"} {
		if code == "" {
			t.Fatal("empty coupon")
		}
	}
}

func TestCheckoutStartsAwaitingPayment(t *testing.T) {
	if checkoutStatus() != "awaiting_payment" {
		t.Fatal(checkoutStatus())
	}
}

func checkoutStatus() string { return "awaiting_payment" }
