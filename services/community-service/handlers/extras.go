package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pawradise/shared/middleware"
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
	if q != "" {
		sqlStr += " WHERE u.name ILIKE $1 OR COALESCE(up.bio,'') ILIKE $1 OR u.email ILIKE $1"
		args = append(args, "%"+q+"%")
	}
	sqlStr += " ORDER BY (SELECT COUNT(*) FROM follows f WHERE f.following_id = u.id) DESC, u.id ASC LIMIT 50"
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
	c.JSON(http.StatusOK, gin.H{"users": users, "total": len(users)})
}

func (h *CommunityHandler) listFollowGraph(c *gin.Context, following bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	viewer, _ := middleware.GetUserIDFromContext(c)
	var rows *sql.Rows
	if following {
		rows, err = h.db.Query("SELECT following_id FROM follows WHERE follower_id = $1", id)
	} else {
		rows, err = h.db.Query("SELECT follower_id FROM follows WHERE following_id = $1", id)
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
	c.JSON(http.StatusOK, gin.H{"users": users, "total": len(users)})
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
