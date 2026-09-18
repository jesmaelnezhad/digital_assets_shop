package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func TestNormalizeStoredRole(t *testing.T) {
	if normalizeStoredRole("user") != "customer" {
		t.Fatal("legacy user -> customer")
	}
	if normalizeStoredRole("STAFF") != "staff" {
		t.Fatal("staff")
	}
}

func TestLoginReturnsRoleAndTabs(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-jwt-secret")
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	hash, err := bcrypt.GenerateFromPassword([]byte("leo"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT id,email,password_hash").WithArgs("leo@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "name", "role", "staff_tabs"}).
			AddRow(2, "leo@example.com", string(hash), "Leo Park", "staff", "Products,Banner"))
	h := &AuthHandler{db: db}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/login", strings.NewReader(`{"email":"leo@example.com","password":"leo"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Login(c)
	if w.Code != 200 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, `"role":"staff"`) || !strings.Contains(body, "Products,Banner") {
		t.Fatalf("expected staff tabs in login: %s", body)
	}
}

func TestLoginRejectsBadPassword(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-jwt-secret")
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	hash, _ := bcrypt.GenerateFromPassword([]byte("leo"), bcrypt.DefaultCost)
	mock.ExpectQuery("SELECT id,email,password_hash").WithArgs("leo@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "name", "role", "staff_tabs"}).
			AddRow(2, "leo@example.com", string(hash), "Leo Park", "staff", "Products"))
	h := &AuthHandler{db: db}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/login", strings.NewReader(`{"email":"leo@example.com","password":"wrong"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Login(c)
	if w.Code != 401 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
}

func TestGetReferralsUsesContextWithoutBearer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := &AuthHandler{db: db}
	mock.ExpectQuery("SELECT COALESCE\\(code").WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"code"}).AddRow("maya-ref"))
	mock.ExpectQuery("SELECT rl.id").WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "created_at", "user_id", "created_at"}))
	mock.ExpectQuery("SELECT COALESCE\\(SUM").WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(0))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", 7)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/referrals", nil)
	h.GetReferrals(c)
	if w.Code != 200 {
		t.Fatalf("cookie-session referrals %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "maya-ref") {
		t.Fatalf("body %s", w.Body.String())
	}
}
