package identity_test

import (
	"os"
	"testing"

	"github.com/pawradise/shared/auth"
	"golang.org/x/crypto/bcrypt"
)

func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET", "test-jwt-secret")
	os.Exit(m.Run())
}

func TestRegisterLoginJWTCarriesCustomerRole(t *testing.T) {
	tok, err := auth.GenerateJWT(9, "maya@example.com", "customer")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := auth.ValidateJWT(tok, os.Getenv("JWT_SECRET"))
	if err != nil {
		t.Fatal(err)
	}
	if claims.Role != "customer" || claims.Email != "maya@example.com" {
		t.Fatalf("%+v", claims)
	}
}

func TestStaffJWTKeepsTabs(t *testing.T) {
	tok, err := auth.GenerateJWT(2, "leo@example.com", "staff", "Products,Banner")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := auth.ValidateJWT(tok, os.Getenv("JWT_SECRET"))
	if err != nil {
		t.Fatal(err)
	}
	if claims.Role != "staff" || claims.Tabs != "Products,Banner" {
		t.Fatalf("%+v", claims)
	}
}

func TestPasswordHashRoundTrip(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("nia"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	if bcrypt.CompareHashAndPassword(hash, []byte("nia")) != nil {
		t.Fatal("verify failed")
	}
	if bcrypt.CompareHashAndPassword(hash, []byte("wrong")) == nil {
		t.Fatal("wrong password should fail")
	}
}

func TestLegacyUserRoleNormalizesToCustomer(t *testing.T) {
	if auth.NormalizeRole("user") != "customer" {
		t.Fatal("user -> customer")
	}
}
