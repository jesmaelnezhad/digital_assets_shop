package handlers

import "strings"

func (h *AuthHandler) ensureRoles() {
	_, _ = h.db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(32) NOT NULL DEFAULT 'customer'`)
	_, _ = h.db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS staff_tabs TEXT NOT NULL DEFAULT ''`)
	_, _ = h.db.Exec(`UPDATE users SET role = 'customer' WHERE role IS NULL OR role = '' OR role = 'user'`)
	var n int
	_ = h.db.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'admin'`).Scan(&n)
	if n == 0 {
		_, _ = h.db.Exec(`UPDATE users SET role = 'admin', staff_tabs = '' WHERE lower(email) = 'nia@example.com'`)
	}
	_, _ = h.db.Exec(`UPDATE users SET role = 'staff', staff_tabs = $1
		WHERE lower(email) = 'leo@example.com' AND role = 'customer'`,
		"Products,Banner,Orders,Community")
}

func normalizeStoredRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin":
		return "admin"
	case "staff":
		return "staff"
	default:
		return "customer"
	}
}
