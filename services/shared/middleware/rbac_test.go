package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pawradise/shared/auth"
)

func TestMain(m *testing.M) {
	os.Setenv("ADMIN_TOKEN", "test-admin-token")
	os.Setenv("JWT_SECRET", "test-jwt-secret")
	os.Exit(m.Run())
}

func rbacRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/api/v1")
	g.Use(AdminAuthMiddleware())
	g.GET("/admin/stats", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	g.GET("/admin/products", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	g.PUT("/products/banner", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	g.PUT("/admin/users/:id/access", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	return r
}

func do(r *gin.Engine, method, path string, headers map[string]string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for _, ck := range cookies {
		req.AddCookie(ck)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAdminAuthRejectsAnonymous(t *testing.T) {
	w := do(rbacRouter(), "GET", "/api/v1/admin/stats", nil)
	if w.Code != 401 {
		t.Fatalf("got %d %s", w.Code, w.Body.String())
	}
}

func TestAdminAuthAcceptsStaticToken(t *testing.T) {
	w := do(rbacRouter(), "GET", "/api/v1/admin/stats", map[string]string{
		"Authorization": "Bearer test-admin-token",
	})
	if w.Code != 200 {
		t.Fatalf("got %d %s", w.Code, w.Body.String())
	}
}

func TestAdminAuthRejectsCustomerJWT(t *testing.T) {
	tok, err := auth.GenerateJWT(9, "maya@example.com", "customer")
	if err != nil {
		t.Fatal(err)
	}
	w := do(rbacRouter(), "GET", "/api/v1/admin/stats", map[string]string{
		"Authorization": "Bearer " + tok,
	})
	if w.Code != 403 {
		t.Fatalf("got %d %s", w.Code, w.Body.String())
	}
}

func TestAdminAuthStaffTabs(t *testing.T) {
	tok, err := auth.GenerateJWT(2, "leo@example.com", "staff", "Products,Banner")
	if err != nil {
		t.Fatal(err)
	}
	hdr := map[string]string{"Authorization": "Bearer " + tok}
	r := rbacRouter()
	if w := do(r, "GET", "/api/v1/admin/products", hdr); w.Code != 200 {
		t.Fatalf("products got %d %s", w.Code, w.Body.String())
	}
	if w := do(r, "PUT", "/api/v1/products/banner", hdr); w.Code != 200 {
		t.Fatalf("banner got %d %s", w.Code, w.Body.String())
	}
	if w := do(r, "GET", "/api/v1/admin/stats", hdr); w.Code != 403 {
		t.Fatalf("stats got %d %s", w.Code, w.Body.String())
	}
}

func TestAdminAuthCookieStaff(t *testing.T) {
	tok, err := auth.GenerateJWT(2, "leo@example.com", "staff", "Products")
	if err != nil {
		t.Fatal(err)
	}
	w := do(rbacRouter(), "GET", "/api/v1/admin/products", nil, &http.Cookie{
		Name:  "pawradise_session",
		Value: tok,
	})
	if w.Code != 200 {
		t.Fatalf("got %d %s", w.Code, w.Body.String())
	}
}

func TestAdminAuthAccessNeedsOperatorToken(t *testing.T) {
	tok, err := auth.GenerateJWT(1, "nia@example.com", "admin")
	if err != nil {
		t.Fatal(err)
	}
	r := rbacRouter()
	hdr := map[string]string{"Authorization": "Bearer " + tok}
	if w := do(r, "PUT", "/api/v1/admin/users/2/access", hdr); w.Code != 403 {
		t.Fatalf("without operator got %d %s", w.Code, w.Body.String())
	}
	hdr["X-Admin-Token"] = "test-admin-token"
	if w := do(r, "PUT", "/api/v1/admin/users/2/access", hdr); w.Code != 200 {
		t.Fatalf("with operator got %d %s", w.Code, w.Body.String())
	}
}

func TestStaffCannotHitAccessEvenWithToken(t *testing.T) {
	tok, err := auth.GenerateJWT(2, "leo@example.com", "staff", "Users,Products")
	if err != nil {
		t.Fatal(err)
	}
	w := do(rbacRouter(), "PUT", "/api/v1/admin/users/3/access", map[string]string{
		"Authorization": "Bearer " + tok,
		"X-Admin-Token": "test-admin-token",
	})
	if w.Code != 403 {
		t.Fatalf("got %d %s", w.Code, w.Body.String())
	}
}

func TestSanitizeStaffTabsDropsAccess(t *testing.T) {
	got := JoinTabs([]string{"Products", "Access", "bogus", "Banner", "Products"})
	if got != "Products,Banner" {
		t.Fatalf("got %q", got)
	}
}

func TestTabForRequestMapsNewSurfaces(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		method, path, want string
	}{
		{"PUT", "/api/v1/products/appearance", "Appearance"},
		{"PUT", "/api/v1/products/banner", "Banner"},
		{"GET", "/api/v1/admin/order-steps", "OrderStepsRead"},
		{"POST", "/api/v1/admin/order-steps", "Steps"},
		{"PUT", "/api/v1/admin/orders/3/status", "Orders"},
		{"PUT", "/api/v1/admin/users/2/access", "Users"},
		{"GET", "/api/v1/admin/community/posts", "Community"},
		{"POST", "/api/v1/admin/export/emails", "Export"},
		{"GET", "/api/v1/admin/events", "Events"},
		{"PUT", "/api/v1/admin/events/ttl", "Events"},
	}
	for _, tc := range cases {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(tc.method, tc.path, nil)
		if got := TabForRequest(c); got != tc.want {
			t.Fatalf("%s %s -> %q want %q", tc.method, tc.path, got, tc.want)
		}
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/api/v1/admin/users/2/access", nil)
	if !PrivilegedAccessPath(c) {
		t.Fatal("access path should be privileged")
	}
}

func TestStaffHasTabOrderStepsRead(t *testing.T) {
	if !StaffHasTab("Orders,Community", "OrderStepsRead") {
		t.Fatal("Orders implies order-steps read")
	}
	if !StaffHasTab("Steps", "OrderStepsRead") {
		t.Fatal("Steps implies order-steps read")
	}
	if StaffHasTab("Products", "OrderStepsRead") {
		t.Fatal("Products should not imply order-steps")
	}
	if !StaffHasTab("Orders,Steps", "Steps") {
		t.Fatal("explicit Steps grant")
	}
}
