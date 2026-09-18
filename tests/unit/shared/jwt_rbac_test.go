package shared_test

import (
	"os"
	"testing"

	"github.com/pawradise/shared/auth"
	"github.com/pawradise/shared/middleware"
)

func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET", "test-jwt-secret")
	os.Setenv("ADMIN_TOKEN", "test-admin-token")
	os.Exit(m.Run())
}

func TestAdminJWTDropsStaffTabs(t *testing.T) {
	tok, err := auth.GenerateJWT(1, "nia@example.com", "admin", "Products,Access")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := auth.ValidateJWT(tok, os.Getenv("JWT_SECRET"))
	if err != nil {
		t.Fatal(err)
	}
	if claims.Role != "admin" {
		t.Fatalf("role %s", claims.Role)
	}
	if claims.Tabs != "" {
		t.Fatalf("admin tabs should be empty, got %q", claims.Tabs)
	}
}

func TestJoinTabsKeepsSteps(t *testing.T) {
	got := middleware.JoinTabs([]string{"Orders", "Steps", "Access"})
	if got != "Orders,Steps" {
		t.Fatalf("got %q", got)
	}
}

func TestGarbageJWTRejected(t *testing.T) {
	if _, err := auth.ValidateJWT("not-a-jwt", os.Getenv("JWT_SECRET")); err == nil {
		t.Fatal("expected error")
	}
}
