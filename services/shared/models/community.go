package models

import "time"

// CommunityPost represents a community post with type and visibility.
type CommunityPost struct {
	ID           int    `json:"id"`
	UserID       int    `json:"user_id"`
	Content      string `json:"content"`
	Type             string `json:"type"`
	IsPublic         bool   `json:"is_public"`
	IsPinned         bool   `json:"is_pinned"`
	ImageURL         string `json:"image_url"`
	LinkURL          string `json:"link_url"`
	LinkTitle        string `json:"link_title"`
	LinkDescription  string `json:"link_description"`
	LinkImage        string `json:"link_image"`
	LikeCount        int    `json:"like_count"`
	CommentCount     int    `json:"comment_count"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// PostComment represents a comment on a community post.
type PostComment struct {
	ID      int    `json:"id"`
	PostID  int    `json:"post_id"`
	Content string `json:"content"`
}

// Comment represents a comment on a post.
type Comment struct {
	ID          int       `json:"id" db:"id"`
	UserID      int       `json:"user_id" db:"user_id"`
	PostID      int       `json:"post_id" db:"post_id"`
	Content     string    `json:"content" db:"content"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	AuthorName  string    `json:"author_name,omitempty" db:"author_name"`
	AuthorAvatar string   `json:"author_avatar,omitempty" db:"author_avatar"`
}

// Like represents a user's like on a post.
type Like struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	PostID    int       `json:"post_id" db:"post_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Follow represents a follow relationship between users.
type Follow struct {
	ID          int       `json:"id" db:"id"`
	FollowerID  int       `json:"follower_id" db:"follower_id"`
	FollowingID int       `json:"following_id" db:"following_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// PublicProfile represents a user's public-facing community profile.
type PublicProfile struct {
	UserID         int    `json:"user_id"`
	Name           string `json:"name"`
	Bio            string `json:"bio,omitempty"`
	AvatarURL      string `json:"avatar_url,omitempty"`
	PostCount      int    `json:"post_count"`
	FollowerCount  int    `json:"follower_count"`
	FollowingCount int    `json:"following_count"`
}

// CreatePostRequest is the payload for creating a post.
type CreatePostRequest struct {
	UserID  int    `json:"user_id" binding:"required"`
	Content string `json:"content" binding:"required,max=5000"`
}

// LikeResponse is returned after toggling a like.
type LikeResponse struct {
	Liked     bool `json:"liked"`
	LikeCount int  `json:"like_count"`
}

// FollowResponse is returned after toggling a follow.
type FollowResponse struct {
	Following     bool `json:"following"`
	FollowerCount int  `json:"follower_count"`
}

// AddCommentRequest is the payload for adding a comment.
type AddCommentRequest struct {
	UserID  int    `json:"user_id" binding:"required"`
	PostID  int    `json:"post_id" binding:"required"`
	Content string `json:"content" binding:"required,max=2000"`
}

// PostFilter defines query parameters for post listing.
type PostFilter struct {
	Page    int `form:"page"`
	PerPage int `form:"per_page"`
	UserID  int `form:"user_id"`
}
