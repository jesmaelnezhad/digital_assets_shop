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

func TestPipelineSlug(t *testing.T) {
	if pipelineSlug("Waiting for payment") != "waiting_for_payment" {
		t.Fatalf("got %q", pipelineSlug("Waiting for payment"))
	}
	if pipelineSlug("  Paid! ") != "paid" {
		t.Fatalf("got %q", pipelineSlug("  Paid! "))
	}
}

func TestDeleteSystemStepRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := &AdminHandler{db: db, commerceDB: db, stepsReady: true}
	mock.ExpectQuery("SELECT slug, is_system").WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"slug", "is_system"}).AddRow("paid", true))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/admin/order-steps/1", nil)
	h.DeleteOrderStep(c)
	if w.Code != 400 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "system") {
		t.Fatalf("body %s", w.Body.String())
	}
}

func TestMoveUnknownStepRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := &AdminHandler{db: db, commerceDB: db, stepsReady: true}
	mock.ExpectQuery("SELECT id, slug, label").WillReturnRows(
		sqlmock.NewRows([]string{"id", "slug", "label", "sort_order", "is_system", "is_terminal"}).
			AddRow(1, "paid", "Paid", 30, true, false),
	)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "9"}}
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/orders/9/status",
		strings.NewReader(`{"status":"not-a-real-step"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.UpdateOrderStatus(c)
	if w.Code != 400 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
}

func TestMoveSameStepIsNoop(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := &AdminHandler{db: db, commerceDB: db, stepsReady: true}
	mock.ExpectQuery("SELECT id, slug, label").WillReturnRows(
		sqlmock.NewRows([]string{"id", "slug", "label", "sort_order", "is_system", "is_terminal"}).
			AddRow(1, "paid", "Paid", 30, true, false),
	)
	mock.ExpectQuery("SELECT status FROM orders").WithArgs(4).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("paid"))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "4"}}
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/orders/4/status",
		strings.NewReader(`{"status":"paid"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.UpdateOrderStatus(c)
	if w.Code != 200 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
}

func TestListOrderSteps(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := &AdminHandler{db: db, commerceDB: db, stepsReady: true}
	mock.ExpectQuery("SELECT id, slug, label").WillReturnRows(
		sqlmock.NewRows([]string{"id", "slug", "label", "sort_order", "is_system", "is_terminal"}).
			AddRow(1, "created", "Created", 10, true, false).
			AddRow(2, "preparation", "Preparation", 40, false, false),
	)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/order-steps", nil)
	h.ListOrderSteps(c)
	if w.Code != 200 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "preparation") {
		t.Fatalf("body %s", w.Body.String())
	}
}

func TestCreateOrderStepConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := &AdminHandler{db: db, commerceDB: db, stepsReady: true}
	mock.ExpectQuery("SELECT id, slug, label").WillReturnRows(
		sqlmock.NewRows([]string{"id", "slug", "label", "sort_order", "is_system", "is_terminal"}).
			AddRow(4, "preparation", "Preparation", 40, false, false),
	)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/order-steps",
		strings.NewReader(`{"label":"Preparation"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.CreateOrderStep(c)
	if w.Code != 409 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
}

func TestListAllOrdersFiltersStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := &AdminHandler{db: db, commerceDB: db, stepsReady: true}
	ts := time.Now()
	mock.ExpectQuery("SELECT id, slug, label").WillReturnRows(
		sqlmock.NewRows([]string{"id", "slug", "label", "sort_order", "is_system", "is_terminal"}).
			AddRow(3, "paid", "Paid", 30, true, false),
	)
	mock.ExpectQuery("FROM orders o").WillReturnRows(
		sqlmock.NewRows([]string{"id", "user_id", "email", "status", "total_usd", "crypto_chain", "crypto_amount", "payment_tx_hash", "payment_confirmations", "paid_at", "created_at", "updated_at"}).
			AddRow(7, 1, "nia@example.com", "paid", 19.0, "BSC", "0.03", "", 0, ts, ts, ts),
	)
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("FROM order_steps s").WillReturnRows(
		sqlmock.NewRows([]string{"slug", "label", "is_terminal", "count"}).
			AddRow("paid", "Paid", false, 1),
	)
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/orders?status=paid", nil)
	h.ListAllOrders(c)
	if w.Code != 200 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"status_label":"Paid"`) {
		t.Fatalf("body %s", w.Body.String())
	}
}

func TestOrderNotFoundOnMove(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	h := &AdminHandler{db: db, commerceDB: db, stepsReady: true}
	mock.ExpectQuery("SELECT id, slug, label").WillReturnRows(
		sqlmock.NewRows([]string{"id", "slug", "label", "sort_order", "is_system", "is_terminal"}).
			AddRow(1, "paid", "Paid", 30, true, false),
	)
	mock.ExpectQuery("SELECT status FROM orders").WithArgs(99).WillReturnError(sql.ErrNoRows)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "99"}}
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/orders/99/status",
		strings.NewReader(`{"status":"paid"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.UpdateOrderStatus(c)
	if w.Code != 404 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
}
