package community_test

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func setupTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	return db, mock
}

func TestCommunityService_GetPosts(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
	}{
		{
			name: "success_list_posts",
			setupMock: func() {
				mock.ExpectQuery(`SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.name, u.avatar_url FROM community_posts cp`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "created_at", "name", "avatar_url"}).
						AddRow(1, 1, "Test post", "2024-01-01", "User", "https://avatar.url"))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM post_likes`).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM post_comments`).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
			},
			wantErr: false,
		},
		{
			name: "success_empty_feed",
			setupMock: func() {
				mock.ExpectQuery(`SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.name, u.avatar_url FROM community_posts cp`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "created_at", "name", "avatar_url"}))
			},
			wantErr: false,
		},
		{
			name: "success_pagination",
			setupMock: func() {
				mock.ExpectQuery(`SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.name, u.avatar_url FROM community_posts cp`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "created_at", "name", "avatar_url"}).
						AddRow(1, 1, "Post 1", "2024-01-01", "User", "https://avatar.url"))
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			// TODO: Call CommunityService.GetPosts(db, page, per_page)
		})
	}
}

func TestCommunityService_CreatePost(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name      string
		content   string
		userID    int
		setupMock func()
		wantErr   bool
	}{
		{
			name:    "success_create_post",
			content: "Test post content",
			userID:  1,
			setupMock: func() {
				mock.ExpectQuery(`INSERT INTO community_posts \(user_id, content\) VALUES \(\$1, \$2\) RETURNING id`).
					WithArgs(1, "Test post content").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			wantErr: false,
		},
		{
			name:    "content_too_long",
			content: string(make([]byte, 501)),
			userID:  1,
			setupMock: func() {
				// Should fail validation before DB call
			},
			wantErr: true,
		},
		{
			name:    "empty_content",
			content: "",
			userID:  1,
			setupMock: func() {
				// Should fail validation before DB call
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			// TODO: Call CommunityService.CreatePost(db, tt.userID, tt.content)
		})
	}
}

func TestCommunityService_GetPost(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name      string
		postID    int
		setupMock func()
		wantErr   bool
	}{
		{
			name:   "success_get_post",
			postID: 1,
			setupMock: func() {
				mock.ExpectQuery(`SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.name, u.avatar_url FROM community_posts cp WHERE cp.id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "created_at", "name", "avatar_url"}).
						AddRow(1, 1, "Test post", "2024-01-01", "User", "https://avatar.url"))
				mock.ExpectQuery(`SELECT pc.id, pc.user_id, pc.content, pc.created_at, u.name, u.avatar_url FROM post_comments pc`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "created_at", "name", "avatar_url"}).
						AddRow(1, 1, "Test comment", "2024-01-01", "User", "https://avatar.url"))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM post_likes`).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM post_comments`).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			wantErr: false,
		},
		{
			name:   "post_not_found",
			postID: 99999,
			setupMock: func() {
				mock.ExpectQuery(`SELECT cp.id, cp.user_id, cp.content, cp.created_at, u.name, u.avatar_url FROM community_posts cp WHERE cp.id = \$1`).
					WithArgs(99999).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "content", "created_at", "name", "avatar_url"}))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			// TODO: Call CommunityService.GetPost(db, tt.postID)
		})
	}
}

func TestCommunityService_DeletePost(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name      string
		postID    int
		setupMock func()
		wantErr   bool
	}{
		{
			name:   "success_delete_post",
			postID: 1,
			setupMock: func() {
				mock.ExpectExec(`DELETE FROM community_posts WHERE id = \$1`).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:   "post_not_found",
			postID: 99999,
			setupMock: func() {
				mock.ExpectExec(`DELETE FROM community_posts WHERE id = \$1`).
					WithArgs(99999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			// TODO: Call CommunityService.DeletePost(db, tt.postID)
		})
	}
}

func TestCommunityService_LikePost(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name      string
		postID    int
		userID    int
		setupMock func()
		wantLiked bool
		wantErr   bool
	}{
		{
			name:   "success_like_post",
			postID: 1,
			userID: 1,
			setupMock: func() {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM post_likes WHERE user_id = \$1 AND post_id = \$2`).
					WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectExec(`INSERT INTO post_likes \(user_id, post_id\) VALUES \(\$1, \$2\) ON CONFLICT DO NOTHING`).
					WithArgs(1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantLiked: true,
			wantErr:   false,
		},
		{
			name:   "success_unlike_post",
			postID: 1,
			userID: 1,
			setupMock: func() {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM post_likes WHERE user_id = \$1 AND post_id = \$2`).
					WithArgs(1, 1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				mock.ExpectExec(`DELETE FROM post_likes WHERE user_id = \$1 AND post_id = \$2`).
					WithArgs(1, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantLiked: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			// TODO: Call CommunityService.LikePost(db, tt.userID, tt.postID)
		})
	}
}

func TestCommunityService_AddComment(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name      string
		postID    int
		userID    int
		content   string
		setupMock func()
		wantErr   bool
	}{
		{
			name:    "success_add_comment",
			postID:  1,
			userID:  1,
			content: "Test comment",
			setupMock: func() {
				mock.ExpectQuery(`INSERT INTO post_comments \(user_id, post_id, content\) VALUES \(\$1, \$2, \$3\) RETURNING id`).
					WithArgs(1, 1, "Test comment").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			wantErr: false,
		},
		{
			name:    "content_too_long",
			postID:  1,
			userID:  1,
			content: string(make([]byte, 501)),
			setupMock: func() {
				// Should fail validation before DB call
			},
			wantErr: true,
		},
		{
			name:    "empty_content",
			postID:  1,
			userID:  1,
			content: "",
			setupMock: func() {
				// Should fail validation before DB call
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			// TODO: Call CommunityService.AddComment(db, tt.userID, tt.postID, tt.content)
		})
	}
}

func TestCommunityService_DeleteComment(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name      string
		commentID int
		userID    int
		setupMock func()
		wantErr   bool
	}{
		{
			name:      "success_delete_comment",
			commentID: 1,
			userID:    1,
			setupMock: func() {
				mock.ExpectExec(`DELETE FROM post_comments WHERE id = \$1 AND user_id = \$2`).
					WithArgs(1, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:      "comment_not_found",
			commentID: 99999,
			userID:    1,
			setupMock: func() {
				mock.ExpectExec(`DELETE FROM post_comments WHERE id = \$1 AND user_id = \$2`).
					WithArgs(99999, 1).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name:      "unauthorized_delete",
			commentID: 1,
			userID:    2,
			setupMock: func() {
				mock.ExpectExec(`DELETE FROM post_comments WHERE id = \$1 AND user_id = \$2`).
					WithArgs(1, 2).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			// TODO: Call CommunityService.DeleteComment(db, tt.commentID, tt.userID)
		})
	}
}

func TestCommunityService_FollowUser(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name        string
		followerID  int
		followingID int
		setupMock   func()
		wantErr     bool
	}{
		{
			name:        "success_follow_user",
			followerID:  1,
			followingID: 2,
			setupMock: func() {
				mock.ExpectExec(`INSERT INTO follows \(follower_id, following_id\) VALUES \(\$1, \$2\) ON CONFLICT DO NOTHING`).
					WithArgs(1, 2).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:        "already_following",
			followerID:  1,
			followingID: 2,
			setupMock: func() {
				mock.ExpectExec(`INSERT INTO follows \(follower_id, following_id\) VALUES \(\$1, \$2\) ON CONFLICT DO NOTHING`).
					WithArgs(1, 2).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: false,
		},
		{
			name:        "cannot_follow_self",
			followerID:  1,
			followingID: 1,
			setupMock: func() {
				// Should fail validation before DB call
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			// TODO: Call CommunityService.FollowUser(db, tt.followerID, tt.followingID)
		})
	}
}

func TestCommunityService_UnfollowUser(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name        string
		followerID  int
		followingID int
		setupMock   func()
		wantErr     bool
	}{
		{
			name:        "success_unfollow_user",
			followerID:  1,
			followingID: 2,
			setupMock: func() {
				mock.ExpectExec(`DELETE FROM follows WHERE follower_id = \$1 AND following_id = \$2`).
					WithArgs(1, 2).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:        "not_following",
			followerID:  1,
			followingID: 2,
			setupMock: func() {
				mock.ExpectExec(`DELETE FROM follows WHERE follower_id = \$1 AND following_id = \$2`).
					WithArgs(1, 2).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			// TODO: Call CommunityService.UnfollowUser(db, tt.followerID, tt.followingID)
		})
	}
}

func TestCommunityService_GetPublicProfile(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	tests := []struct {
		name      string
		userID    int
		setupMock func()
		wantErr   bool
	}{
		{
			name:   "success_get_profile",
			userID: 1,
			setupMock: func() {
				mock.ExpectQuery(`SELECT id, name, email FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
						AddRow(1, "Test User", "test@example.com"))
				mock.ExpectQuery(`SELECT bio, avatar_url, wallet_address FROM user_profiles WHERE user_id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"bio", "avatar_url", "wallet_address"}).
						AddRow("Bio", "https://avatar.url", "0x1234"))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM community_posts WHERE user_id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM follows WHERE following_id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM follows WHERE follower_id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
			},
			wantErr: false,
		},
		{
			name:   "user_not_found",
			userID: 99999,
			setupMock: func() {
				mock.ExpectQuery(`SELECT id, name, email FROM users WHERE id = \$1`).
					WithArgs(99999).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			// TODO: Call CommunityService.GetPublicProfile(db, tt.userID)
		})
	}
}
