package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListCommunityPosts returns all community posts for moderation.
func ListCommunityPosts(c *gin.Context) {
	limit := 50
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed := parseInt(l); parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed := parseInt(o); parsed >= 0 {
			offset = parsed
		}
	}

	statusFilter := c.Query("status")

	query := `
		SELECT cp.id, u.email, cp.title, cp.content, cp.status, cp.created_at,
		       COUNT(cpc.id) AS comment_count
		FROM community_posts cp
		JOIN users u ON u.id = cp.user_id
		LEFT JOIN community_post_comments cpc ON cpc.post_id = cp.id
	`
	args := []interface{}{}
	if statusFilter != "" {
		query += " WHERE cp.status = $1"
		args = append(args, statusFilter)
	}
	query += " GROUP BY cp.id, u.email ORDER BY cp.created_at DESC"
	query += " LIMIT $" + itoa(len(args)+1) + " OFFSET $" + itoa(len(args)+2)
	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch community posts"})
		return
	}
	defer rows.Close()

	var posts []gin.H
	for rows.Next() {
		var (
			id           int
			email        string
			title        string
			content      string
			status       string
			createdAt    sql.NullTime
			commentCount int
		)
		if err := rows.Scan(&id, &email, &title, &content, &status, &createdAt, &commentCount); err != nil {
			continue
		}
		posts = append(posts, gin.H{
			"id":            id,
			"user_email":    email,
			"title":         title,
			"content":       content,
			"status":        status,
			"created_at":    nullTimeToString(createdAt),
			"comment_count": commentCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"posts":  posts,
		"limit":  limit,
		"offset": offset,
	})
}

// DeleteCommunityPost removes a community post (moderation action).
func DeleteCommunityPost(c *gin.Context) {
	postID := c.Param("id")

	// Delete comments first (foreign key constraint)
	_, err := db.Exec("DELETE FROM community_post_comments WHERE post_id = $1", postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post comments"})
		return
	}

	result, err := db.Exec("DELETE FROM community_posts WHERE id = $1", postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete community post"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Community post not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Community post deleted successfully"})
}
