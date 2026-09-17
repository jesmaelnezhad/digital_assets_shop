package handlers

import (
	"database/sql"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pawradise/shared/middleware"
	"github.com/pawradise/shared/models"
)

type CommunityHandler struct {
	db *sql.DB
}

func NewCommunityHandler(db *sql.DB) *CommunityHandler {
	return &CommunityHandler{db: db}
}

func (h *CommunityHandler) GetFeed(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 { page = 1 }
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if perPage < 1 || perPage > 50 { perPage = 20 }
	offset := (page - 1) * perPage
	filter := c.DefaultQuery("filter", "recent")
	profileUser := c.Query("user_id")

	var rows *sql.Rows
	var err error
	var total int

	userID, _ := middleware.GetUserIDFromContext(c)

	if profileUser != "" {
		uid, _ := strconv.Atoi(profileUser)
		h.db.QueryRow("SELECT COUNT(*) FROM community_posts WHERE user_id = $1", uid).Scan(&total)
		rows, err = h.db.Query(
			`SELECT cp.id, cp.user_id, cp.content, cp.community_type, cp.is_public, cp.is_pinned,
			cp.like_count, cp.comment_count, cp.created_at, cp.updated_at,
			COALESCE(u.email, ''), COALESCE(u.name, ''), COALESCE(up.display_name, ''), COALESCE(up.avatar_url, ''),
			(SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = cp.user_id)) as following,
			(SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = cp.user_id AND following_id = $1)) as follows_you,
			(SELECT EXISTS(SELECT 1 FROM post_likes WHERE user_id = $1 AND post_id = cp.id)) as liked
			FROM community_posts cp
			LEFT JOIN users u ON cp.user_id = u.id
			LEFT JOIN user_profiles up ON cp.user_id = up.user_id
			WHERE cp.user_id = $2
			ORDER BY cp.created_at DESC LIMIT $3 OFFSET $4`,
			userID, uid, perPage, offset,
		)
	} else if filter == "following" && userID > 0 {
		h.db.QueryRow("SELECT COUNT(*) FROM community_posts WHERE user_id IN (SELECT following_id FROM follows WHERE follower_id = $1)", userID).Scan(&total)
		rows, err = h.db.Query(
			`SELECT cp.id, cp.user_id, cp.content, cp.community_type, cp.is_public, cp.is_pinned,
			cp.like_count, cp.comment_count, cp.created_at, cp.updated_at,
			COALESCE(u.email, ''), COALESCE(u.name, ''), COALESCE(up.display_name, ''), COALESCE(up.avatar_url, ''),
			(SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = cp.user_id)) as following,
			(SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = cp.user_id AND following_id = $1)) as follows_you,
			(SELECT EXISTS(SELECT 1 FROM post_likes WHERE user_id = $1 AND post_id = cp.id)) as liked
			FROM community_posts cp
			LEFT JOIN users u ON cp.user_id = u.id
			LEFT JOIN user_profiles up ON cp.user_id = up.user_id
			WHERE cp.user_id IN (SELECT following_id FROM follows WHERE follower_id = $1)
			ORDER BY cp.created_at DESC LIMIT $2 OFFSET $3`,
			userID, perPage, offset,
		)
	} else {
		if filter == "following" {
			total = 0
			rows, err = h.db.Query(
				`SELECT cp.id, cp.user_id, cp.content, cp.community_type, cp.is_public, cp.is_pinned,
				cp.like_count, cp.comment_count, cp.created_at, cp.updated_at,
				COALESCE(u.email, ''), COALESCE(u.name, ''), COALESCE(up.display_name, ''), COALESCE(up.avatar_url, ''),
				false, false, false
				FROM community_posts cp
				LEFT JOIN users u ON cp.user_id = u.id
				LEFT JOIN user_profiles up ON cp.user_id = up.user_id
				WHERE false
				LIMIT 0`)
		} else {
			h.db.QueryRow("SELECT COUNT(*) FROM community_posts WHERE is_public = true").Scan(&total)
			rows, err = h.db.Query(
				`SELECT cp.id, cp.user_id, cp.content, cp.community_type, cp.is_public, cp.is_pinned,
				cp.like_count, cp.comment_count, cp.created_at, cp.updated_at,
				COALESCE(u.email, ''), COALESCE(u.name, ''), COALESCE(up.display_name, ''), COALESCE(up.avatar_url, ''),
				(SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = cp.user_id)) as following,
				(SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = cp.user_id AND following_id = $1)) as follows_you,
				(SELECT EXISTS(SELECT 1 FROM post_likes WHERE user_id = $1 AND post_id = cp.id)) as liked
				FROM community_posts cp
				LEFT JOIN users u ON cp.user_id = u.id
				LEFT JOIN user_profiles up ON cp.user_id = up.user_id
				WHERE cp.is_public = true
				ORDER BY cp.created_at DESC LIMIT $2 OFFSET $3`,
				userID, perPage, offset,
			)
		}
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch feed"})
		return
	}
	defer rows.Close()

	type postWithAuthor struct {
		ID          int    `json:"id"`
		UserID      int    `json:"user_id"`
		Content     string `json:"content"`
		Type        string `json:"type"`
		IsPublic    bool   `json:"is_public"`
		IsPinned    bool   `json:"is_pinned"`
		LikeCount   int    `json:"like_count"`
		CommentCount int   `json:"comment_count"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
		Author      struct {
			ID        int    `json:"id"`
			Email     string `json:"email"`
			Name      string `json:"name"`
			Display   string `json:"display_name"`
			Avatar    string `json:"avatar_url"`
			Following bool   `json:"following"`
			FollowsYou bool  `json:"follows_you"`
		} `json:"author"`
		Liked bool `json:"liked"`
	}
	var posts []postWithAuthor
	for rows.Next() {
		var p postWithAuthor
		var createdAt, updatedAt sql.NullTime
		var avatar, display sql.NullString
		var following, followsYou, liked bool
		if err := rows.Scan(&p.ID, &p.UserID, &p.Content, &p.Type, &p.IsPublic, &p.IsPinned,
			&p.LikeCount, &p.CommentCount, &createdAt, &updatedAt,
			&p.Author.Email, &p.Author.Name, &display, &avatar, &following, &followsYou, &liked); err != nil { continue }
		if createdAt.Valid { p.CreatedAt = createdAt.Time.Format(time.RFC3339) }
		if updatedAt.Valid { p.UpdatedAt = updatedAt.Time.Format(time.RFC3339) }
		if avatar.Valid { p.Author.Avatar = avatar.String }
		if display.Valid { p.Author.Display = display.String }
		p.Author.Following = following
		p.Author.FollowsYou = followsYou
		p.Liked = liked
		p.Author.ID = p.UserID
		posts = append(posts, p)
	}
	if posts == nil { posts = []postWithAuthor{} }

	c.JSON(http.StatusOK, gin.H{"posts": posts, "total": total, "page": page, "per_page": perPage})
}

func (h *CommunityHandler) CreatePost(c *gin.Context) {
	uid, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Content  string `json:"content" binding:"required"`
		Type     string `json:"type"`
		IsPublic *bool  `json:"is_public"`
		IsPinned *bool  `json:"is_pinned"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Content) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content too long"})
		return
	}

	if req.Type == "" { req.Type = "post" }
	if req.IsPublic == nil { req.IsPublic = new(bool); *req.IsPublic = true }

	// Ensure user exists in community DB
	var existingUserID int
	err := h.db.QueryRow("SELECT id FROM users WHERE id = $1", uid).Scan(&existingUserID)
	if err == sql.ErrNoRows {
		// User doesn't exist, create a placeholder
		_, _ = h.db.Exec("INSERT INTO users (id, email, name) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING", uid, "user@example.com", "User")
		_, _ = h.db.Exec("INSERT INTO user_profiles (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING", uid)
	}

	var id int
	err = h.db.QueryRow(
		`INSERT INTO community_posts (user_id, content, community_type, is_public, is_pinned)
		 VALUES ($1, $2, $3, $4, COALESCE($5, false)) RETURNING id`,
		uid, req.Content, req.Type, *req.IsPublic, req.IsPinned,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create post"})
		return
	}

	h.db.Exec("UPDATE user_profiles SET updated_at = NOW() WHERE user_id = $1", uid)

	c.JSON(http.StatusCreated, gin.H{"id": id, "user_id": uid, "content": req.Content, "type": req.Type})
}

func (h *CommunityHandler) GetPost(c *gin.Context) {
	idStr := c.Param("id")
	postID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	var p models.CommunityPost
	var createdAt, updatedAt sql.NullTime
	err = h.db.QueryRow(
		`SELECT cp.id, cp.user_id, cp.content, cp.community_type, cp.is_public, cp.is_pinned,
		cp.like_count, cp.comment_count, cp.created_at, cp.updated_at
		FROM community_posts cp WHERE cp.id = $1`,
		postID,
	).Scan(&p.ID, &p.UserID, &p.Content, &p.Type, &p.IsPublic, &p.IsPinned,
		&p.LikeCount, &p.CommentCount, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if createdAt.Valid { p.CreatedAt = createdAt.Time.Format(time.RFC3339) }
	if updatedAt.Valid { p.UpdatedAt = updatedAt.Time.Format(time.RFC3339) }

	var author struct {
		ID      int    `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Display string `json:"display_name"`
		Avatar  string `json:"avatar_url"`
	}
	h.db.QueryRow(
		`SELECT u.id, COALESCE(u.email, ''), COALESCE(u.name, ''), COALESCE(up.display_name, ''), COALESCE(up.avatar_url, '')
		FROM users u LEFT JOIN user_profiles up ON u.id = up.user_id WHERE u.id = $1`,
		p.UserID,
	).Scan(&author.ID, &author.Email, &author.Name, &author.Display, &author.Avatar)

	commentRows, _ := h.db.Query(
		`SELECT pc.id, pc.post_id, pc.user_id, pc.content, pc.like_count, pc.created_at, pc.updated_at,
		        COALESCE(u.email, ''), COALESCE(u.name, ''), COALESCE(up.display_name, ''), COALESCE(up.avatar_url, '')
		FROM post_comments pc LEFT JOIN users u ON pc.user_id = u.id
		LEFT JOIN user_profiles up ON pc.user_id = up.user_id
		WHERE pc.post_id = $1 ORDER BY pc.created_at ASC`,
		postID,
	)
	defer commentRows.Close()

	type comment struct {
		ID        int    `json:"id"`
		PostID    int    `json:"post_id"`
		UserID    int    `json:"user_id"`
		Content   string `json:"content"`
		LikeCount int    `json:"like_count"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		User      struct {
			ID      int    `json:"id"`
			Email   string `json:"email"`
			Name    string `json:"name"`
			Display string `json:"display_name"`
			Avatar  string `json:"avatar_url"`
		} `json:"user"`
	}
	var comments []comment
	for commentRows.Next() {
		var c comment
		var createdAt, updatedAt sql.NullTime
		var display, avatar sql.NullString
		if err := commentRows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Content, &c.LikeCount, &createdAt, &updatedAt,
			&c.User.Email, &c.User.Name, &display, &avatar); err != nil {
			continue
		}
		if createdAt.Valid { c.CreatedAt = createdAt.Time.Format(time.RFC3339) }
		if updatedAt.Valid { c.UpdatedAt = updatedAt.Time.Format(time.RFC3339) }
		if display.Valid { c.User.Display = display.String }
		if avatar.Valid { c.User.Avatar = avatar.String }
		comments = append(comments, c)
	}
	if comments == nil { comments = []comment{} }

	c.JSON(http.StatusOK, gin.H{
		"post":    p,
		"author":  author,
		"comments": comments,
		"total_comments": len(comments),
	})
}

func (h *CommunityHandler) DeletePost(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid := userID.(int)

	postIDStr := c.Param("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	var ownerID int
	err = h.db.QueryRow("SELECT user_id FROM community_posts WHERE id = $1", postID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") || uid != ownerID {
		adminToken := os.Getenv("ADMIN_TOKEN")
		if adminToken == "" { adminToken = "admin-secret-token-change-in-production" }
		if !strings.HasPrefix(authHeader, "Bearer ") || strings.TrimPrefix(authHeader, "Bearer ") != adminToken {
			c.JSON(http.StatusForbidden, gin.H{"error": "not your post"})
			return
		}
	}

	result, err := h.db.Exec("DELETE FROM community_posts WHERE id = $1", postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "post deleted"})
}

func (h *CommunityHandler) LikePost(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid := userID.(int)

	postIDStr := c.Param("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	var existing int
	h.db.QueryRow("SELECT 1 FROM post_likes WHERE user_id = $1 AND post_id = $2", uid, postID).Scan(&existing)
	if existing == 1 {
		c.JSON(http.StatusConflict, gin.H{"error": "already liked"})
		return
	}

	_, err = h.db.Exec("INSERT INTO post_likes (user_id, post_id) VALUES ($1, $2)", uid, postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "like failed"})
		return
	}

	h.db.Exec("UPDATE community_posts SET like_count = like_count + 1, updated_at = NOW() WHERE id = $1", postID)

	c.JSON(http.StatusOK, gin.H{"message": "post liked"})
}

func (h *CommunityHandler) UnlikePost(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid := userID.(int)

	postIDStr := c.Param("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	result, err := h.db.Exec("DELETE FROM post_likes WHERE user_id = $1 AND post_id = $2", uid, postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unlike failed"})
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "like not found"})
		return
	}

	h.db.Exec("UPDATE community_posts SET like_count = GREATEST(0, like_count - 1), updated_at = NOW() WHERE id = $1", postID)

	c.JSON(http.StatusOK, gin.H{"message": "post unliked"})
}

func (h *CommunityHandler) AddComment(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid := userID.(int)

	postIDStr := c.Param("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var commentID int
	err = h.db.QueryRow(
		`INSERT INTO post_comments (post_id, user_id, content) VALUES ($1, $2, $3) RETURNING id`,
		postID, uid, req.Content,
	).Scan(&commentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add comment"})
		return
	}

	h.db.Exec("UPDATE community_posts SET comment_count = comment_count + 1, updated_at = NOW() WHERE id = $1", postID)

	c.JSON(http.StatusCreated, gin.H{"comment": models.PostComment{ID: commentID, PostID: postID, Content: req.Content}})
}

func (h *CommunityHandler) DeleteComment(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid := userID.(int)

	postIDStr := c.Param("id")
	commentIDStr := c.Param("commentId")
	postID, err1 := strconv.Atoi(postIDStr)
	commentID, err2 := strconv.Atoi(commentIDStr)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ids"})
		return
	}

	var ownerID int
	err := h.db.QueryRow("SELECT user_id FROM post_comments WHERE id = $1 AND post_id = $2", commentID, postID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if ownerID != uid {
		authHeader := c.GetHeader("Authorization")
		adminToken := os.Getenv("ADMIN_TOKEN")
		if adminToken == "" { adminToken = "admin-secret-token-change-in-production" }
		if !strings.HasPrefix(authHeader, "Bearer ") || strings.TrimPrefix(authHeader, "Bearer ") != adminToken {
			c.JSON(http.StatusForbidden, gin.H{"error": "not your comment"})
			return
		}
	}

	result, err := h.db.Exec("DELETE FROM post_comments WHERE id = $1 AND post_id = $2", commentID, postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}

	h.db.Exec("UPDATE community_posts SET comment_count = GREATEST(0, comment_count - 1), updated_at = NOW() WHERE id = $1", postID)

	c.JSON(http.StatusOK, gin.H{"message": "comment deleted"})
}

func (h *CommunityHandler) FollowUser(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid := userID.(int)

	targetIDStr := c.Param("userId")
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if targetID == uid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot follow yourself"})
		return
	}

	// Ensure both users exist in community DB
	var existing int
	err = h.db.QueryRow("SELECT id FROM users WHERE id = $1", uid).Scan(&existing)
	if err == sql.ErrNoRows {
		_, _ = h.db.Exec("INSERT INTO users (id, email, name) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING", uid, "user@example.com", "User")
		_, _ = h.db.Exec("INSERT INTO user_profiles (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING", uid)
	}

	err = h.db.QueryRow("SELECT id FROM users WHERE id = $1", targetID).Scan(&existing)
	if err == sql.ErrNoRows {
		_, _ = h.db.Exec("INSERT INTO users (id, email, name) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING", targetID, "user@example.com", "User")
		_, _ = h.db.Exec("INSERT INTO user_profiles (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING", targetID)
	}

	var alreadyFollowing int
	h.db.QueryRow("SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = $2", uid, targetID).Scan(&alreadyFollowing)
	if alreadyFollowing == 1 {
		c.JSON(http.StatusConflict, gin.H{"error": "already following"})
		return
	}

	_, err = h.db.Exec("INSERT INTO follows (follower_id, following_id) VALUES ($1, $2)", uid, targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "follow failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user followed"})
}

func (h *CommunityHandler) UnfollowUser(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid := userID.(int)

	targetIDStr := c.Param("userId")
	targetID, err := strconv.Atoi(targetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	result, err := h.db.Exec("DELETE FROM follows WHERE follower_id = $1 AND following_id = $2", uid, targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unfollow failed"})
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not following"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user unfollowed"})
}

func (h *CommunityHandler) GetPublicProfile(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	// Ensure user exists
	var existingUserID int
	err = h.db.QueryRow("SELECT id FROM users WHERE id = $1", userID).Scan(&existingUserID)
	if err == sql.ErrNoRows {
		_, _ = h.db.Exec("INSERT INTO users (id, email, name) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING", userID, "user@example.com", "User")
		_, _ = h.db.Exec("INSERT INTO user_profiles (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING", userID)
	}

	var u models.User
	var createdAt, updatedAt sql.NullTime
	err = h.db.QueryRow(
		"SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1",
		userID,
	).Scan(&u.ID, &u.Email, &u.Name, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if createdAt.Valid { u.CreatedAt = createdAt.Time }
	if updatedAt.Valid { u.UpdatedAt = updatedAt.Time }

	var avatar, bio sql.NullString
	h.db.QueryRow("SELECT avatar_url, bio FROM user_profiles WHERE user_id = $1", userID).Scan(&avatar, &bio)
	if avatar.Valid { u.AvatarURL = avatar.String }
	bioStr := ""
	if bio.Valid { bioStr = bio.String }

	var postCount, followerCount, followingCount int
	h.db.QueryRow("SELECT COUNT(*) FROM community_posts WHERE user_id = $1", userID).Scan(&postCount)
	h.db.QueryRow("SELECT COUNT(*) FROM follows WHERE following_id = $1", userID).Scan(&followerCount)
	h.db.QueryRow("SELECT COUNT(*) FROM follows WHERE follower_id = $1", userID).Scan(&followingCount)

	viewer, _ := middleware.GetUserIDFromContext(c)
	following, followsYou, isSelf := false, false, false
	if viewer > 0 {
		var n int
		h.db.QueryRow("SELECT 1 FROM follows WHERE follower_id=$1 AND following_id=$2", viewer, userID).Scan(&n)
		following = n == 1
		n = 0
		h.db.QueryRow("SELECT 1 FROM follows WHERE follower_id=$1 AND following_id=$2", userID, viewer).Scan(&n)
		followsYou = n == 1
		isSelf = viewer == userID
	}

	c.JSON(http.StatusOK, gin.H{
		"id":              u.ID,
		"email":           u.Email,
		"name":            u.Name,
		"avatar_url":      u.AvatarURL,
		"bio":             bioStr,
		"post_count":      postCount,
		"follower_count":  followerCount,
		"following_count": followingCount,
		"following":       following,
		"follows_you":     followsYou,
		"is_self":         isSelf,
		"created_at":      u.CreatedAt,
		"updated_at":      u.UpdatedAt,
	})
}

func (h *CommunityHandler) GetMyProfile(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid := userID.(int)

	var u models.User
	var createdAt, updatedAt sql.NullTime
	h.db.QueryRow(
		"SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1",
		uid,
	).Scan(&u.ID, &u.Email, &u.Name, &createdAt, &updatedAt)
	if err := h.db.QueryRow("SELECT id FROM users WHERE id = $1", uid).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if createdAt.Valid { u.CreatedAt = createdAt.Time }
	if updatedAt.Valid { u.UpdatedAt = updatedAt.Time }

	var avatar, bio sql.NullString
	h.db.QueryRow("SELECT avatar_url, bio FROM user_profiles WHERE user_id = $1", uid).Scan(&avatar, &bio)
	if avatar.Valid { u.AvatarURL = avatar.String }

	var postCount, followerCount, followingCount int
	h.db.QueryRow("SELECT COUNT(*) FROM community_posts WHERE user_id = $1", uid).Scan(&postCount)
	h.db.QueryRow("SELECT COUNT(*) FROM follows WHERE following_id = $1", uid).Scan(&followerCount)
	h.db.QueryRow("SELECT COUNT(*) FROM follows WHERE follower_id = $1", uid).Scan(&followingCount)

	c.JSON(http.StatusOK, gin.H{
		"user":           u,
		"post_count":     postCount,
		"follower_count": followerCount,
		"following_count": followingCount,
	})
}

func (h *CommunityHandler) UpdateProfile(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid := userID.(int)

	var req struct {
		DisplayName string `json:"display_name"`
		AvatarURL   string `json:"avatar_url"`
		Bio         string `json:"bio"`
		Location    string `json:"location"`
		Website     string `json:"website"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing struct{ UserID int }
	err := h.db.QueryRow("SELECT user_id FROM user_profiles WHERE user_id = $1", uid).Scan(&existing.UserID)

	if err == sql.ErrNoRows {
		_, err = h.db.Exec(
			"INSERT INTO user_profiles (user_id, display_name, avatar_url, bio, location, website) VALUES ($1, $2, $3, $4, $5, $6)",
			uid, req.DisplayName, req.AvatarURL, req.Bio, req.Location, req.Website,
		)
	} else if err == nil {
		_, err = h.db.Exec(
			"UPDATE user_profiles SET display_name = COALESCE(NULLIF($1, ''), display_name), avatar_url = COALESCE(NULLIF($2, ''), avatar_url), bio = COALESCE(NULLIF($3, ''), bio), location = COALESCE(NULLIF($4, ''), location), website = COALESCE(NULLIF($5, ''), website), updated_at = NOW() WHERE user_id = $6",
			req.DisplayName, req.AvatarURL, req.Bio, req.Location, req.Website, uid,
		)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "profile update failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "profile updated"})
}
