package handlers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func TestGetExchangeRates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ts := time.Now()
	mock.ExpectQuery("SELECT id, chain, symbol").WillReturnRows(
		sqlmock.NewRows([]string{"id", "chain", "symbol", "rate_to_usd", "updated_at"}).
			AddRow(1, "bsc", "USDT", 1.0, ts),
	)
	h := NewPaymentHandler(db)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/exchange-rates", nil)
	h.GetExchangeRates(c)
	if w.Code != 200 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
}

func TestGetExchangeRateMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT id, chain, symbol").WithArgs("nope").WillReturnError(sql.ErrNoRows)
	h := NewPaymentHandler(db)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "chain", Value: "nope"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/exchange-rates/nope", nil)
	h.GetExchangeRate(c)
	if w.Code != 404 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
}
