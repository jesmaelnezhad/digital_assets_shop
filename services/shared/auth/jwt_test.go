package auth

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET", "test-jwt-secret")
	os.Exit(m.Run())
}

func TestNormalizeRole(t *testing.T) {
	cases := map[string]string{
		"": "customer", "user": "customer", "USER": "customer",
		"customer": "customer", "staff": "staff", "Admin": "admin",
	}
	for in, want := range cases {
		if got := NormalizeRole(in); got != want {
			t.Fatalf("NormalizeRole(%q)=%q want %q", in, got, want)
		}
	}
}

func TestGenerateAndValidateJWTCarriesRoleAndTabs(t *testing.T) {
	tok, err := GenerateJWT(2, "leo@example.com", "staff", "Products,Banner")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := ValidateJWT(tok, os.Getenv("JWT_SECRET"))
	if err != nil || claims == nil {
		t.Fatalf("validate: %v", err)
	}
	if claims.UserID != 2 || claims.Email != "leo@example.com" {
		t.Fatalf("identity %+v", claims)
	}
	if claims.Role != "staff" || claims.Tabs != "Products,Banner" {
		t.Fatalf("role/tabs %+v", claims)
	}
}

func TestGenerateJWTDropsTabsForNonStaff(t *testing.T) {
	tok, err := GenerateJWT(1, "nia@example.com", "admin", "Products")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := ValidateJWTWithDefaultSecret(tok)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Role != "admin" || claims.Tabs != "" {
		t.Fatalf("admin should not keep staff tabs: %+v", claims)
	}
}

func TestValidateJWTRejectsGarbage(t *testing.T) {
	if _, err := ValidateJWT("not-a-jwt", os.Getenv("JWT_SECRET")); err == nil {
		t.Fatal("expected error")
	}
}

func TestHashTokenStable(t *testing.T) {
	a := HashToken("abc")
	b := HashToken("abc")
	c := HashToken("abcd")
	if a == "" || a != b || a == c {
		t.Fatalf("hash a=%s b=%s c=%s", a, b, c)
	}
}
