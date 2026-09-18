package handlers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func TestSetUserAccessPromotesStaffTabs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	idb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer idb.Close()
	h := &AdminHandler{identityDB: idb}
	mock.ExpectQuery("SELECT COALESCE\\(role").WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("customer"))
	mock.ExpectExec("UPDATE users SET role").WithArgs("staff", "Products,Banner", 2).
		WillReturnResult(sqlmock.NewResult(0, 1))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/2/access",
		strings.NewReader(`{"role":"staff","staff_tabs":["Products","Banner","Access"]}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.SetUserAccess(c)
	if w.Code != 200 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"role":"staff"`) {
		t.Fatalf("body %s", w.Body.String())
	}
	if strings.Contains(w.Body.String(), "Access") {
		t.Fatalf("Access tab must be stripped: %s", w.Body.String())
	}
}

func TestSetUserAccessBlocksLastAdminDemotion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	idb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer idb.Close()
	h := &AdminHandler{identityDB: idb}
	mock.ExpectQuery("SELECT COALESCE\\(role").WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("admin"))
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/1/access",
		strings.NewReader(`{"role":"customer"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.SetUserAccess(c)
	if w.Code != 400 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
}

func TestSetUserAccessUnknownUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	idb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer idb.Close()
	h := &AdminHandler{identityDB: idb}
	mock.ExpectQuery("SELECT COALESCE\\(role").WithArgs(99).WillReturnError(sql.ErrNoRows)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "99"}}
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/99/access",
		strings.NewReader(`{"role":"staff"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.SetUserAccess(c)
	if w.Code != 404 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
}

func TestListUsersReturnsRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	idb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer idb.Close()
	h := &AdminHandler{identityDB: idb}
	now := time.Now()
	mock.ExpectQuery("SELECT id, email, name").WillReturnRows(
		sqlmock.NewRows([]string{"id", "email", "name", "role", "staff_tabs", "created_at", "updated_at"}).
			AddRow(1, "nia@example.com", "Nia", "admin", "", now, now).
			AddRow(2, "leo@example.com", "Leo", "staff", "Products,Banner", now, now),
	)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	h.ListUsers(c)
	if w.Code != 200 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, `"role":"admin"`) || !strings.Contains(body, `"role":"staff"`) {
		t.Fatalf("expected roles: %s", body)
	}
}
