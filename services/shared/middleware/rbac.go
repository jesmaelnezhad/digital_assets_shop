package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/store4bots/shared/auth"
)

// StaffAssignableTabs are admin desk sections that may be granted to staff.
var StaffAssignableTabs = []string{
	"Stats", "Users", "Products", "Banner", "Categories", "Bundles", "Coupons",
	"Orders", "Steps", "Guest", "Community", "Referrals", "Rates", "Settings", "SEO",
	"Export", "Requests", "Appearance", "Events",
}

var seoSettingKeys = map[string]bool{
	"site_title": true, "site_description": true, "site_keywords": true,
	"og_image": true, "canonical_host": true, "robots_index": true,
}

var appearanceSettingKeys = map[string]bool{
	"site_appearance": true, "theme_palette": true, "theme_font": true,
	"theme_radius": true, "theme_density": true,
}

func requestPath(c *gin.Context) string {
	p := c.FullPath()
	if p == "" && c.Request != nil {
		p = c.Request.URL.Path
	}
	return strings.ToLower(p)
}

// PrivilegedAccessPath is true for role/tab mutations. Those require an admin
// JWT plus the operator token (or the static ADMIN_TOKEN bearer).
func PrivilegedAccessPath(c *gin.Context) bool {
	p := requestPath(c)
	if !strings.Contains(p, "/admin/users/") {
		return false
	}
	return strings.HasSuffix(p, "/access") || strings.HasSuffix(p, "/role")
}

// TabForRequest maps an admin API call to a desk tab name.
func TabForRequest(c *gin.Context) string {
	p := requestPath(c)
	method := ""
	if c.Request != nil {
		method = strings.ToUpper(c.Request.Method)
	}
	key := ""
	if c.Params != nil {
		key = c.Param("key")
	}

	if strings.Contains(p, "/products/appearance") || strings.HasSuffix(p, "/appearance") {
		return "Appearance"
	}
	if strings.Contains(p, "/products/banner") {
		return "Banner"
	}
	if strings.Contains(p, "/admin/product-requests") {
		return "Requests"
	}
	if strings.Contains(p, "/admin/guest") {
		return "Guest"
	}
	if strings.Contains(p, "/admin/order-steps") {
		if method == "GET" {
			return "OrderStepsRead"
		}
		return "Steps"
	}
	if strings.Contains(p, "/admin/orders") {
		return "Orders"
	}
	if strings.Contains(p, "/admin/community") {
		return "Community"
	}
	if strings.Contains(p, "/admin/referrals") {
		return "Referrals"
	}
	if strings.Contains(p, "/exchange-rates") {
		return "Rates"
	}
	if strings.Contains(p, "/admin/events") {
		return "Events"
	}
	if strings.Contains(p, "/admin/coupons") {
		return "Coupons"
	}
	if strings.Contains(p, "/admin/bundles") || (strings.Contains(p, "/bundles") && method != "GET") {
		return "Bundles"
	}
	if strings.Contains(p, "/admin/categories") || (strings.Contains(p, "/categories") && method != "GET") {
		return "Categories"
	}
	if strings.Contains(p, "/admin/export") {
		return "Export"
	}
	if strings.Contains(p, "/admin/stats") || strings.Contains(p, "/products/stats") {
		return "Stats"
	}
	if strings.Contains(p, "/admin/users") {
		return "Users"
	}
	if strings.Contains(p, "/admin/settings") || (strings.Contains(p, "/settings") && method != "GET") {
		if method == "GET" && key == "" {
			return "SettingsRead"
		}
		if appearanceSettingKeys[key] {
			return "Appearance"
		}
		if seoSettingKeys[key] {
			return "SEO"
		}
		return "Settings"
	}
	if strings.Contains(p, "/pin") {
		return "Products"
	}
	if strings.Contains(p, "/admin/products") || (strings.Contains(p, "/products") && method != "GET") {
		return "Products"
	}
	return ""
}

// StaffHasTab reports whether comma-separated tabs include want.
func StaffHasTab(tabs, want string) bool {
	if want == "" {
		return false
	}
	if want == "SettingsRead" {
		return StaffHasAny(tabs, "Settings", "SEO", "Appearance")
	}
	if want == "OrderStepsRead" {
		return StaffHasAny(tabs, "Orders", "Steps")
	}
	for _, t := range strings.Split(tabs, ",") {
		if strings.EqualFold(strings.TrimSpace(t), want) {
			return true
		}
	}
	return false
}

// StaffHasAny is true if any of the named tabs is granted.
func StaffHasAny(tabs string, names ...string) bool {
	for _, name := range names {
		if StaffHasTab(tabs, name) {
			return true
		}
	}
	return false
}

// SanitizeStaffTabs keeps only assignable tab names.
func SanitizeStaffTabs(raw []string) []string {
	allow := map[string]bool{}
	for _, t := range StaffAssignableTabs {
		allow[t] = true
	}
	out := []string{}
	seen := map[string]bool{}
	for _, t := range raw {
		t = strings.TrimSpace(t)
		if t == "" || !allow[t] || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
	}
	return out
}

// SplitTabs turns a stored staff_tabs value into a slice.
func SplitTabs(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return SanitizeStaffTabs(out)
}

// JoinTabs stores staff tabs as a comma-separated string.
func JoinTabs(tabs []string) string {
	return strings.Join(SanitizeStaffTabs(tabs), ",")
}

// RoleOf returns a normalized role from gin context.
func RoleOf(c *gin.Context) string {
	v, _ := c.Get("role")
	s, _ := v.(string)
	return auth.NormalizeRole(s)
}
