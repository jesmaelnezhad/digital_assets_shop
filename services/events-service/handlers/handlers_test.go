package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestClampTTL(t *testing.T) {
	if ClampTTL(0) != DefaultTTLSeconds {
		t.Fatal("default")
	}
	if ClampTTL(10) != MinTTLSeconds {
		t.Fatal("min")
	}
	if ClampTTL(7200) != 7200 {
		t.Fatal("pass through")
	}
}

func TestAllowedEventNames(t *testing.T) {
	if !AllowedEvent("product_view") || !AllowedEvent("checkout_click") {
		t.Fatal("known events")
	}
	if AllowedEvent("page_view") || AllowedEvent("") {
		t.Fatal("unknown must fail")
	}
}

func TestIngestRejectsUnknown(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := New(NewMemoryStore())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/events", strings.NewReader(`{"name":"page_view"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Ingest(c)
	if w.Code != 400 {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
}

func TestIngestAndListRoundTrip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := NewMemoryStore()
	h := New(store)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/events", strings.NewReader(`{"name":"product_view","session_id":"s1","path":"/product/lunar-clay-characters","properties":{"product_id":1}}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Ingest(c)
	if w.Code != 202 {
		t.Fatalf("ingest %d %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/events", nil)
	h.List(c)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "product_view") {
		t.Fatalf("list %d %s", w.Code, w.Body.String())
	}
}

func TestSetTTLHours(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := New(NewMemoryStore())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/events/ttl", strings.NewReader(`{"hours":2}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.SetTTL(c)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"seconds":7200`) {
		t.Fatalf("ttl %d %s", w.Code, w.Body.String())
	}
}
