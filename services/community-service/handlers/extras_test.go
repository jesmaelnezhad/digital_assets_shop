package handlers

import (
	"net/http"
	"net/http/httptest"
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
	mock.ExpectQuery("SELECT u.id FROM users").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
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
}
