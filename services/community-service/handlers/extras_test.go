package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func TestListPeople(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := NewCommunityHandler(db)
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT u.id FROM users").WithArgs(12, 0).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectQuery("SELECT COALESCE").WithArgs(2).WillReturnRows(
		sqlmock.NewRows([]string{"email", "name", "avatar_url", "bio"}).AddRow("nia@example.com", "Nia", "", "maker"),
	)
	mock.ExpectQuery("SELECT COUNT").WithArgs(2).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectQuery("SELECT COUNT").WithArgs(2).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/community/users", nil)
	h.ListPeople(c)
	if w.Code != 200 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"total":1`) {
		t.Fatalf("expected total in people payload: %s", w.Body.String())
	}
}

func TestListPeoplePage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := NewCommunityHandler(db)
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(40))
	mock.ExpectQuery("SELECT u.id FROM users").WithArgs(5, 5).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(9))
	mock.ExpectQuery("SELECT COALESCE").WithArgs(9).WillReturnRows(
		sqlmock.NewRows([]string{"email", "name", "avatar_url", "bio"}).AddRow("c@example.com", "Cora", "", "bio"),
	)
	mock.ExpectQuery("SELECT COUNT").WithArgs(9).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT COUNT").WithArgs(9).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/community/users?page=2&per_page=5", nil)
	h.ListPeople(c)
	if w.Code != 200 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
}

func TestListPeopleSearch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := NewCommunityHandler(db)
	mock.ExpectQuery("SELECT COUNT").WithArgs("%maya%").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT u.id FROM users").WithArgs("%maya%", 12, 0).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(4))
	mock.ExpectQuery("SELECT COALESCE").WithArgs(4).WillReturnRows(
		sqlmock.NewRows([]string{"email", "name", "avatar_url", "bio"}).AddRow("maya@example.com", "Maya Chen", "", "buyer"),
	)
	mock.ExpectQuery("SELECT COUNT").WithArgs(4).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT COUNT").WithArgs(4).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/community/users?q=maya", nil)
	h.ListPeople(c)
	if w.Code != 200 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Maya Chen") {
		t.Fatalf("body %s", w.Body.String())
	}
}

func TestListFollowersRejectsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCommunityHandler(nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "x"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/community/users/x/followers", nil)
	h.ListFollowers(c)
	if w.Code != 400 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
}
