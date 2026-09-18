package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/store4bots/shared/middleware"
)

type memberJSON struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	AvatarURL      string `json:"avatar_url"`
	Bio            string `json:"bio"`
	FollowerCount  int    `json:"follower_count"`
	FollowingCount int    `json:"following_count"`
	Following      bool   `json:"following"`
	FollowsYou     bool   `json:"follows_you"`
	IsSelf         bool   `json:"is_self"`
}

func (h *CommunityHandler) loadMember(id, viewer int) (memberJSON, error) {
	var m memberJSON
	m.ID = id
	err := h.db.QueryRow(
		`SELECT COALESCE(u.email,''), COALESCE(u.name,''), COALESCE(up.avatar_url,''), COALESCE(up.bio,'')
		 FROM users u LEFT JOIN user_profiles up ON up.user_id = u.id WHERE u.id = $1`,
		id,
	).Scan(&m.Email, &m.Name, &m.AvatarURL, &m.Bio)
	if err != nil {
		return m, err
	}
	h.db.QueryRow("SELECT COUNT(*) FROM follows WHERE following_id = $1", id).Scan(&m.FollowerCount)
	h.db.QueryRow("SELECT COUNT(*) FROM follows WHERE follower_id = $1", id).Scan(&m.FollowingCount)
	if viewer > 0 {
		var n int
		h.db.QueryRow("SELECT 1 FROM follows WHERE follower_id=$1 AND following_id=$2", viewer, id).Scan(&n)
		m.Following = n == 1
		n = 0
		h.db.QueryRow("SELECT 1 FROM follows WHERE follower_id=$1 AND following_id=$2", id, viewer).Scan(&n)
		m.FollowsYou = n == 1
		m.IsSelf = viewer == id
	}
	return m, nil
}

func (h *CommunityHandler) ListPeople(c *gin.Context) {
	viewer, _ := middleware.GetUserIDFromContext(c)
	q := c.Query("q")
	sqlStr := `SELECT u.id FROM users u LEFT JOIN user_profiles up ON up.user_id = u.id`
	args := []interface{}{}
	n := 0
	if q != "" {
		n++
		sqlStr += fmt.Sprintf(" WHERE u.name ILIKE $%d OR COALESCE(up.bio,'') ILIKE $%d OR u.email ILIKE $%d", n, n, n)
		args = append(args, "%"+q+"%")
	}
	countSQL := "SELECT COUNT(*) FROM users u LEFT JOIN user_profiles up ON up.user_id = u.id"
	if q != "" {
		countSQL += " WHERE u.name ILIKE $1 OR COALESCE(up.bio,'') ILIKE $1 OR u.email ILIKE $1"
	}
	var total int
	if q != "" {
		_ = h.db.QueryRow(countSQL, "%"+q+"%").Scan(&total)
	} else {
		_ = h.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&total)
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "12"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 12
	}
	offset := (page - 1) * perPage
	n++
	limP := n
	n++
	offP := n
	sqlStr += fmt.Sprintf(" ORDER BY (SELECT COUNT(*) FROM follows f WHERE f.following_id = u.id) DESC, u.id ASC LIMIT $%d OFFSET $%d", limP, offP)
	args = append(args, perPage, offset)
	rows, err := h.db.Query(sqlStr, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	defer rows.Close()
	users := []memberJSON{}
	for rows.Next() {
		var id int
		if rows.Scan(&id) != nil {
			continue
		}
		m, err := h.loadMember(id, viewer)
		if err == nil {
			users = append(users, m)
		}
	}
	if users == nil {
		users = []memberJSON{}
	}
	c.JSON(http.StatusOK, gin.H{"users": users, "total": total, "page": page, "per_page": perPage})
}

func pageParams(c *gin.Context, defPer, maxPer int) (page, per, offset int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	per, _ = strconv.Atoi(c.DefaultQuery("per_page", strconv.Itoa(defPer)))
	if page < 1 {
		page = 1
	}
	if per < 1 || per > maxPer {
		per = defPer
	}
	return page, per, (page - 1) * per
}

func (h *CommunityHandler) listFollowGraph(c *gin.Context, following bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	viewer, _ := middleware.GetUserIDFromContext(c)
	page, perPage, offset := pageParams(c, 12, 50)
	var total int
	var rows *sql.Rows
	if following {
		_ = h.db.QueryRow("SELECT COUNT(*) FROM follows WHERE follower_id = $1", id).Scan(&total)
		rows, err = h.db.Query("SELECT following_id FROM follows WHERE follower_id = $1 ORDER BY id DESC LIMIT $2 OFFSET $3", id, perPage, offset)
	} else {
		_ = h.db.QueryRow("SELECT COUNT(*) FROM follows WHERE following_id = $1", id).Scan(&total)
		rows, err = h.db.Query("SELECT follower_id FROM follows WHERE following_id = $1 ORDER BY id DESC LIMIT $2 OFFSET $3", id, perPage, offset)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	defer rows.Close()
	users := []memberJSON{}
	for rows.Next() {
		var oid int
		if rows.Scan(&oid) != nil {
			continue
		}
		m, err := h.loadMember(oid, viewer)
		if err == nil {
			users = append(users, m)
		}
	}
	if users == nil {
		users = []memberJSON{}
	}
	c.JSON(http.StatusOK, gin.H{"users": users, "total": total, "page": page, "per_page": perPage})
}

func (h *CommunityHandler) ListFollowers(c *gin.Context) {
	h.listFollowGraph(c, false)
}

func (h *CommunityHandler) ListFollowing(c *gin.Context) {
	h.listFollowGraph(c, true)
}

func (h *CommunityHandler) Suggestions(c *gin.Context) {
	viewer, _ := middleware.GetUserIDFromContext(c)
	q := `SELECT u.id FROM users u WHERE ($1 = 0 OR u.id <> $1)
		AND ($1 = 0 OR u.id NOT IN (SELECT following_id FROM follows WHERE follower_id = $1))
		ORDER BY (SELECT COUNT(*) FROM follows f WHERE f.following_id = u.id) DESC, u.id ASC LIMIT 12`
	rows, err := h.db.Query(q, viewer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	defer rows.Close()
	users := []memberJSON{}
	for rows.Next() {
		var id int
		if rows.Scan(&id) != nil {
			continue
		}
		m, err := h.loadMember(id, viewer)
		if err == nil {
			users = append(users, m)
		}
	}
	if users == nil {
		users = []memberJSON{}
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}
