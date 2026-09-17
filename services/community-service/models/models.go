package models

import "time"

// Post represents a community post.
type Post struct {
	ID           int       `json:"id" db:"id"`
	UserID       int       `json:"user_id" db:"user_id"`
	Type         string    `json:"type" db:"type"`
	CommunityType string   `json:"community_type" db:"community_type"`
	Content      string    `json:"content" db:"content"`
	IsPublic     bool      `json:"is_public" db:"is_public"`
	IsPinned     bool      `json:"is_pinned" db:"is_pinned"`
	LikeCount    int       `json:"like_count" db:"like_count"`
	CommentCount int       `json:"comment_count" db:"comment_count"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Comment represents a comment on a post.
type Comment struct {
	ID        int       `json:"id" db:"id"`
	PostID    int       `json:"post_id" db:"post_id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Content   string    `json:"content" db:"content"`
	LikeCount int       `json:"like_count" db:"like_count"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// PostLike tracks which users liked which posts.
type PostLike struct {
	ID        int       `json:"id" db:"id"`
	PostID    int       `json:"post_id" db:"post_id"`
	UserID    int       `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// CommentLike tracks which users liked which comments.
type CommentLike struct {
	ID        int       `json:"id" db:"id"`
	CommentID int       `json:"comment_id" db:"comment_id"`
	UserID    int       `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Follow represents a user following another user.
type Follow struct {
	ID          int       `json:"id" db:"id"`
	FollowerID  int       `json:"follower_id" db:"follower_id"`
	FollowingID int       `json:"following_id" db:"following_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// PublicProfile is a user's public-facing community profile.
type PublicProfile struct {
	UserID       int    `json:"user_id"`
	Name         string `json:"name"`
	Bio          string `json:"bio"`
	AvatarURL    string `json:"avatar_url"`
	FollowerCount int   `json:"follower_count"`
	FollowingCount int `json:"following_count"`
	PostCount    int    `json:"post_count"`
}

// CreatePostRequest is the payload for creating a post.
type CreatePostRequest struct {
	Title   string `json:"title" binding:"required,max=255"`
	Content string `json:"content" binding:"required"`
}

// CreateCommentRequest is the payload for creating a comment.
type CreateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

// PostListResponse wraps a paginated list of posts.
type PostListResponse struct {
	Posts    []Post `json:"posts"`
	Total    int    `json:"total"`
	Page     int    `json:"page"`
	PerPage  int    `json:"per_page"`
}

// CommentListResponse wraps a list of comments.
type CommentListResponse struct {
	Comments []Comment `json:"comments"`
	Total    int       `json:"total"`
}
