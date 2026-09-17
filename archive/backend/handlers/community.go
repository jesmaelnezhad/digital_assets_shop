package handlers

import (
	"backend/database"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GetFeed(c *gin.Context) {
	uid := getUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 { page = 1 }
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if perPage < 1 || perPage > 50 { perPage = 20 }
	offset := (page - 1) * perPage
	filter := c.DefaultQuery("filter", "recent")

	var rows *sql.Rows
	var err error
	if filter == "following" && uid > 0 {
		rows, err = database.DB.Query(`
			SELECT cp.id, cp.user_id, cp.content, cp.created_at,
			       u.email, u.name, up.avatar_url,
			       (SELECT COUNT(*) FROM post_likes WHERE post_id = cp.id),
			       (SELECT COUNT(*) FROM post_comments WHERE post_id = cp.id),
			       EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = cp.user_id)
			FROM community_posts cp
			JOIN users u ON cp.user_id = u.id
			LEFT JOIN user_profiles up ON cp.user_id = up.user_id
			WHERE cp.user_id IN (SELECT following_id FROM follows WHERE follower_id = $1)
			ORDER BY cp.created_at DESC LIMIT $2 OFFSET $3
		`, uid, perPage, offset)
	} else {
		rows, err = database.DB.Query(`
			SELECT cp.id, cp.user_id, cp.content, cp.created_at,
			       u.email, u.name, up.avatar_url,
			       (SELECT COUNT(*) FROM post_likes WHERE post_id = cp.id),
			       (SELECT COUNT(*) FROM post_comments WHERE post_id = cp.id),
			       EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = cp.user_id)
			FROM community_posts cp
			JOIN users u ON cp.user_id = u.id
			LEFT JOIN user_profiles up ON cp.user_id = up.user_id
			ORDER BY cp.created_at DESC LIMIT $2 OFFSET $3
		`, uid, perPage, offset)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch feed"})
		return
	}
	defer rows.Close()

	posts := []map[string]interface{}{}
	for rows.Next() {
		var id, userID int
		var content string
		var createdAt time.Time
		var email, name, avatarURL sql.NullString
		var likeCount, commentCount int
		var isFollowed bool
		if rows.Scan(&id, &userID, &content, &createdAt, &email, &name, &avatarURL, &likeCount, &commentCount, &isFollowed) != nil {
			continue
		}
		posts = append(posts, map[string]interface{}{
			"id": id, "user_id": userID, "content": content,
			"created_at": createdAt.Format(time.RFC3339),
			"user": map[string]interface{}{"id": userID, "email": email.String, "name": name.String, "avatar_url": avatarURL.String},
			"like_count": likeCount, "comment_count": commentCount, "is_followed": isFollowed,
		})
	}
	c.JSON(http.StatusOK, gin.H{"posts": posts, "filter": filter})
}

func CreatePost(c *gin.Context) {
	uid, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	userID := uid.(int)
	var req struct {
		Content string `json:"content" binding:"required,max=500"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	if len(req.Content) == 0 || len(req.Content) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content must be 1-500 characters"})
		return
	}
	var postID int
	if err := database.DB.QueryRow(
		"INSERT INTO community_posts (user_id, content) VALUES ($1, $2) RETURNING id",
		userID, req.Content,
	).Scan(&postID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create post"})
		return
	}
	var post struct {
		ID int `json:"id"`; UserID int `json:"user_id"`; Content string `json:"content"`
		CreatedAt time.Time `json:"created_at"`; Email string `json:"email"`; Name string `json:"name"`
		AvatarURL *string `json:"avatar_url"`; LikeCount int `json:"like_count"`; CommentCount int `json:"comment_count"`
	}
	database.DB.QueryRow(`SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, up.avatar_url,
		(SELECT COUNT(*) FROM post_likes WHERE post_id = cp.id), (SELECT COUNT(*) FROM post_comments WHERE post_id = cp.id)
		FROM community_posts cp JOIN users u ON cp.user_id = u.id LEFT JOIN user_profiles up ON cp.user_id = up.user_id WHERE cp.id = $1`, postID,
	).Scan(&post.ID, &post.UserID, &post.Content, &post.CreatedAt, &post.Email, &post.Name, &post.AvatarURL, &post.LikeCount, &post.CommentCount)
	c.JSON(http.StatusCreated, gin.H{"post": post, "message": "post created"})
}

func GetPost(c *gin.Context) {
	postID, _ := strconv.Atoi(c.Param("id"))
	var post struct {
		ID int `json:"id"`; UserID int `json:"user_id"`; Content string `json:"content"`
		CreatedAt time.Time `json:"created_at"`; Email string `json:"email"`; Name string `json:"name"`
		AvatarURL *string `json:"avatar_url"`; LikeCount int `json:"like_count"`; CommentCount int `json:"comment_count"`
	}
	err := database.DB.QueryRow(`SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.email, u.name, up.avatar_url,
		(SELECT COUNT(*) FROM post_likes WHERE post_id = cp.id), (SELECT COUNT(*) FROM post_comments WHERE post_id = cp.id)
		FROM community_posts cp JOIN users u ON cp.user_id = u.id LEFT JOIN user_profiles up ON cp.user_id = up.user_id WHERE cp.id = $1`, postID,
	).Scan(&post.ID, &post.UserID, &post.Content, &post.CreatedAt, &post.Email, &post.Name, &post.AvatarURL, &post.LikeCount, &post.CommentCount)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch post"})
		return
	}
	var isLiked bool
	if uid, ok := c.Get("user_id"); ok {
		database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM post_likes WHERE user_id = $1 AND post_id = $2)", uid.(int), postID).Scan(&isLiked)
	}
	rows, cerr := database.DB.Query(`SELECT pc.id, pc.user_id, pc.content, pc.created_at, u.email, u.name, up.avatar_url,
		EXISTS(SELECT 1 FROM post_likes WHERE post_id = pc.id AND user_id = pc.user_id) FROM post_comments pc
		JOIN users u ON pc.user_id = u.id LEFT JOIN user_profiles up ON pc.user_id = up.user_id
		WHERE pc.post_id = $1 ORDER BY pc.created_at ASC`, postID)
	comments := []map[string]interface{}{}
	if cerr == nil {
		defer rows.Close()
		for rows.Next() {
			var cid, cuid int; var ccontent string; var ccreated time.Time
			var cmail, cname, cavatar sql.NullString; var cLiked bool
			if rows.Scan(&cid, &cuid, &ccontent, &ccreated, &cmail, &cname, &cavatar, &cLiked) != nil { continue }
			comments = append(comments, map[string]interface{}{
				"id": cid, "user_id": cuid, "content": ccontent, "created_at": ccreated.Format(time.RFC3339),
				"user": map[string]interface{}{"id": cuid, "email": cmail.String, "name": cname.String, "avatar_url": cavatar.String},
				"user_liked": cLiked,
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"post": post, "is_liked": isLiked, "comments": comments})
}

func LikePost(c *gin.Context) {
	uid, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	userID := uid.(int)
	postID, _ := strconv.Atoi(c.Param("id"))

	// Verify post exists
	var postExists int
	database.DB.QueryRow("SELECT COUNT(*) FROM community_posts WHERE id = $1", postID).Scan(&postExists)
	if postExists == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}

	var existing int
	database.DB.QueryRow("SELECT COUNT(*) FROM post_likes WHERE user_id = $1 AND post_id = $2", userID, postID).Scan(&existing)
	if existing > 0 {
		database.DB.Exec("DELETE FROM post_likes WHERE user_id = $1 AND post_id = $2", userID, postID)
		c.JSON(http.StatusOK, gin.H{"liked": false, "message": "like removed"})
	} else {
		database.DB.Exec("INSERT INTO post_likes (user_id, post_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, postID)
		c.JSON(http.StatusOK, gin.H{"liked": true, "message": "post liked"})
	}
}

func UnlikePost(c *gin.Context) {
	uid, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	userID := uid.(int)
	postID, _ := strconv.Atoi(c.Param("id"))

	// Verify post exists
	var postExists int
	database.DB.QueryRow("SELECT COUNT(*) FROM community_posts WHERE id = $1", postID).Scan(&postExists)
	if postExists == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}

	result, err := database.DB.Exec("DELETE FROM post_likes WHERE user_id = $1 AND post_id = $2", userID, postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unlike"})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "like not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"liked": false, "message": "like removed"})
}

func AddComment(c *gin.Context) {
	uid, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	userID := uid.(int)
	postID, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Content string `json:"content" binding:"required,max=500"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if len(req.Content) == 0 || len(req.Content) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content must be 1-500 characters"})
		return
	}
	var postExists int
	database.DB.QueryRow("SELECT COUNT(*) FROM community_posts WHERE id = $1", postID).Scan(&postExists)
	if postExists == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}
	var commentID int
	if err := database.DB.QueryRow(
		"INSERT INTO post_comments (user_id, post_id, content) VALUES ($1, $2, $3) RETURNING id",
		userID, postID, req.Content,
	).Scan(&commentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add comment"})
		return
	}
	var comment struct {
		ID int `json:"id"`; UserID int `json:"user_id"`; PostID int `json:"post_id"`
		Content string `json:"content"`; CreatedAt time.Time `json:"created_at"`
		Email string `json:"email"`; Name string `json:"name"`; AvatarURL *string `json:"avatar_url"`
	}
	database.DB.QueryRow(`SELECT pc.id, pc.user_id, pc.post_id, pc.content, pc.created_at, u.email, u.name, up.avatar_url
		FROM post_comments pc JOIN users u ON pc.user_id = u.id LEFT JOIN user_profiles up ON pc.user_id = up.user_id WHERE pc.id = $1`, commentID,
	).Scan(&comment.ID, &comment.UserID, &comment.PostID, &comment.Content, &comment.CreatedAt, &comment.Email, &comment.Name, &comment.AvatarURL)
	c.JSON(http.StatusCreated, gin.H{"comment": comment, "message": "comment added"})
}

func DeleteComment(c *gin.Context) {
	uid, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	userID := uid.(int)
	commentID, _ := strconv.Atoi(c.Param("commentId"))
	var ownerID int
	database.DB.QueryRow("SELECT user_id FROM post_comments WHERE id = $1", commentID).Scan(&ownerID)
	if ownerID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}
	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your comment"})
		return
	}
	result, err := database.DB.Exec("DELETE FROM post_comments WHERE id = $1", commentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "comment deleted"})
}

func FollowUser(c *gin.Context) {
	uid, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	followerID := uid.(int)
	targetID, _ := strconv.Atoi(c.Param("userId"))
	if targetID == followerID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot follow yourself"})
		return
	}
	var targetExists int
	database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE id = $1", targetID).Scan(&targetExists)
	if targetExists == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	var already int
	database.DB.QueryRow("SELECT COUNT(*) FROM follows WHERE follower_id = $1 AND following_id = $2", followerID, targetID).Scan(&already)
	if already > 0 {
		c.JSON(http.StatusOK, gin.H{"following": true, "message": "already following"})
		return
	}
	database.DB.Exec("INSERT INTO follows (follower_id, following_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", followerID, targetID)
	c.JSON(http.StatusOK, gin.H{"following": true, "message": "now following"})
}

func UnfollowUser(c *gin.Context) {
	uid, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	followerID := uid.(int)
	targetID, _ := strconv.Atoi(c.Param("userId"))
	result, err := database.DB.Exec("DELETE FROM follows WHERE follower_id = $1 AND following_id = $2", followerID, targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unfollow"})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not following"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"following": false, "message": "unfollowed"})
}

func GetPublicProfile(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var user struct {
		ID int `json:"id"`; Name string `json:"name"`
	}
	var profile struct {
		Bio *string `json:"bio"`; AvatarURL *string `json:"avatar_url"`; WalletAddr *string `json:"wallet_address"`
		CreatedAt time.Time `json:"created_at"`
	}
	database.DB.QueryRow("SELECT id, name FROM users WHERE id = $1", id).Scan(&user.ID, &user.Name)
	if user.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	database.DB.QueryRow("SELECT bio, avatar_url, wallet_address, created_at FROM user_profiles WHERE user_id = $1", id).Scan(&profile.Bio, &profile.AvatarURL, &profile.WalletAddr, &profile.CreatedAt)
	var posts, followers, following int
	database.DB.QueryRow("SELECT COUNT(*) FROM community_posts WHERE user_id = $1", id).Scan(&posts)
	database.DB.QueryRow("SELECT COUNT(*) FROM follows WHERE following_id = $1", id).Scan(&followers)
	database.DB.QueryRow("SELECT COUNT(*) FROM follows WHERE follower_id = $1", id).Scan(&following)
	var followingMe int
	if uid, ok := c.Get("user_id"); ok {
		database.DB.QueryRow("SELECT COUNT(*) FROM follows WHERE follower_id = $1 AND following_id = $2", uid.(int), id).Scan(&followingMe)
	}
	c.JSON(http.StatusOK, gin.H{"user": user, "profile": profile, "post_count": posts, "followers": followers, "following": following, "is_following": followingMe > 0})
}

func GetMyProfile(c *gin.Context) {
	uid, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	userID := uid.(int)
	var p struct {
		ID int `json:"id"`; UserID int `json:"user_id"`; Bio *string `json:"bio"`; AvatarURL *string `json:"avatar_url"`
		WalletAddr *string `json:"wallet_address"`; CreatedAt time.Time `json:"created_at"`; UpdatedAt time.Time `json:"updated_at"`
	}
	err := database.DB.QueryRow("SELECT id, user_id, bio, avatar_url, wallet_address, created_at, updated_at FROM user_profiles WHERE user_id = $1", userID).Scan(&p.ID, &p.UserID, &p.Bio, &p.AvatarURL, &p.WalletAddr, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		database.DB.Exec("INSERT INTO user_profiles (user_id) VALUES ($1) ON CONFLICT DO NOTHING", userID)
		p.UserID = userID
	}
	var email, name string
	database.DB.QueryRow("SELECT email, name FROM users WHERE id = $1", userID).Scan(&email, &name)
	c.JSON(http.StatusOK, gin.H{"profile": p, "user": map[string]interface{}{"id": userID, "email": email, "name": name}})
}

func UpdateProfile(c *gin.Context) {
	uid, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	userID := uid.(int)
	var req struct {
		Name *string `json:"name"`; Bio *string `json:"bio"`; AvatarURL *string `json:"avatar_url"`; WalletAddr *string `json:"wallet_address"`
	}
	c.ShouldBindJSON(&req)
	if req.WalletAddr != nil && len(*req.WalletAddr) > 0 && len(*req.WalletAddr) < 20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet address"})
		return
	}
	if req.Name != nil {
		database.DB.Exec("UPDATE users SET name = $1 WHERE id = $2", *req.Name, userID)
	}
	_, err := database.DB.Exec(`INSERT INTO user_profiles (user_id, bio, avatar_url, wallet_address) VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET bio = COALESCE($2, user_profiles.bio), avatar_url = COALESCE($3, user_profiles.avatar_url),
		wallet_address = COALESCE($4, user_profiles.wallet_address), updated_at = NOW()`, userID, req.Bio, req.AvatarURL, req.WalletAddr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "profile updated"})
}

func RegisterCommunityRoutes(r *gin.RouterGroup) {
	community := r.Group("/community")
	community.Use(JwtAuthMiddleware())
	{
		community.GET("/feed", GetFeed)
		community.POST("/posts", CreatePost)
		community.GET("/posts/:id", GetPost)
		community.POST("/posts/:id/like", LikePost)
		community.DELETE("/posts/:id/like", UnlikePost)
		community.POST("/posts/:id/comments", AddComment)
		community.DELETE("/posts/:id/comments/:commentId", DeleteComment)
		community.POST("/follow/:userId", FollowUser)
		community.DELETE("/follow/:userId", UnfollowUser)
		community.GET("/users/:id", GetPublicProfile)
	}
	profile := r.Group("/profile")
	profile.Use(JwtAuthMiddleware())
	{
		profile.GET("", GetMyProfile)
		profile.PUT("", UpdateProfile)
	}
}

func getUserID(c *gin.Context) int {
	if uid, ok := c.Get("user_id"); ok {
		return uid.(int)
	}
	return 0
}
