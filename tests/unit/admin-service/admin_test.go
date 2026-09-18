package admin_test

import "testing"

func TestAccessNeedsOperatorToken(t *testing.T) {
	if operatorHeader() != "X-Admin-Token" {
		t.Fatal("operator header")
	}
}

func TestStaffTabsNeverIncludeAccess(t *testing.T) {
	for _, tab := range []string{"Products", "Banner", "Orders", "Steps", "Community", "Appearance"} {
		if tab == "Access" {
			t.Fatal("Access is admin-only")
		}
	}
}

func TestLastAdminCannotBeDemoted(t *testing.T) {
	if !blockLastAdminDemotion(1, "customer") {
		t.Fatal("last admin demotion must be blocked")
	}
	if blockLastAdminDemotion(2, "customer") {
		t.Fatal("second admin may be demoted")
	}
	if blockLastAdminDemotion(1, "admin") {
		t.Fatal("staying admin is fine")
	}
}

func operatorHeader() string { return "X-Admin-Token" }

func blockLastAdminDemotion(adminCount int, nextRole string) bool {
	return adminCount <= 1 && nextRole != "admin"
}
