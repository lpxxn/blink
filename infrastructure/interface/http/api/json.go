package httpapi

import (
	"strconv"

	domaincategory "github.com/lpxxn/blink/domain/category"
	domainpost "github.com/lpxxn/blink/domain/post"
	domainpostreply "github.com/lpxxn/blink/domain/postreply"
	domainuser "github.com/lpxxn/blink/domain/user"
)

// PostJSON is the API wire shape: snowflake IDs use json ",string" so browsers do not lose precision.
type PostJSON struct {
	ID              int64   `json:"id,string"`
	UserID          int64   `json:"user_id,string"`
	UserName        string  `json:"user_name"`
	PostType        int     `json:"post_type"`
	Visibility      int     `json:"visibility"`
	Body            string  `json:"body"`
	Images          []string `json:"images"`
	Status          int     `json:"status"`
	ModerationFlag  int     `json:"moderation_flag"`
	ModerationNote  string  `json:"moderation_note"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
	CategoryID      *int64 `json:"category_id,string"`
}

// PostToJSON maps a domain post to the JSON response (IDs as strings in JSON).
func PostToJSON(p *domainpost.Post) PostJSON {
	return PostJSON{
		ID:             p.ID,
		UserID:         p.UserID,
		PostType:       p.PostType,
		Visibility:     p.Visibility,
		Body:           p.Body,
		Images:         p.Images,
		Status:         p.Status,
		ModerationFlag: p.ModerationFlag,
		ModerationNote: p.ModerationNote,
		CreatedAt:      p.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      p.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		CategoryID:     p.CategoryID,
	}
}

// CategoryJSON is the wire shape for categories.
type CategoryJSON struct {
	ID        int64  `json:"id,string"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

// CategoryToJSON maps a domain category.
func CategoryToJSON(c *domaincategory.Category) CategoryJSON {
	return CategoryJSON{
		ID:        c.ID,
		Slug:      c.Slug,
		Name:      c.Name,
		SortOrder: c.SortOrder,
	}
}

// CategoriesResponse is GET /api/categories.
type CategoriesResponse struct {
	Categories []CategoryJSON `json:"categories"`
}

// ReplyJSON is the wire shape for post replies.
type ReplyJSON struct {
	ID            int64  `json:"id,string"`
	PostID        int64  `json:"post_id,string"`
	UserID        int64  `json:"user_id,string"`
	UserName      string `json:"user_name"`
	Body          string `json:"body"`
	Status        int    `json:"status"`
	CreatedAt     string `json:"created_at"`
	ParentReplyID *int64 `json:"parent_reply_id,string"`
}

// ReplyToJSON maps a domain reply.
func ReplyToJSON(r *domainpostreply.Reply) ReplyJSON {
	return ReplyJSON{
		ID:            r.ID,
		PostID:        r.PostID,
		UserID:        r.UserID,
		Body:          r.Body,
		Status:        r.Status,
		CreatedAt:     r.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		ParentReplyID: r.ParentReplyID,
	}
}

// PostsPageJSON is the list response for GET /api/posts and GET /api/me/posts.
type PostsPageJSON struct {
	Posts      []PostJSON `json:"posts"`
	NextCursor *string    `json:"next_cursor,omitempty"`
}

// RepliesPageJSON is the list response for GET /api/posts/:id/replies.
type RepliesPageJSON struct {
	Replies    []ReplyJSON `json:"replies"`
	NextCursor *string     `json:"next_cursor,omitempty"`
}

// NextCursorString encodes a snowflake cursor for JSON (string, not Number).
func NextCursorString(id int64) *string {
	s := strconv.FormatInt(id, 10)
	return &s
}

// AdminUserJSON is the wire shape for GET /admin/api/users rows.
type AdminUserJSON struct {
	SnowflakeID       int64   `json:"id,string"`
	Email             string  `json:"email"`
	Name              string  `json:"name"`
	Status            int     `json:"status"`
	Role              string  `json:"role"`
	LastLoginAt       *string `json:"last_login_at,omitempty"`
	LastLoginIP       string  `json:"last_login_ip"`
	LastLoginDevice   string  `json:"last_login_device"`
	CreatedAt         string  `json:"created_at"`
}

// AdminUserToJSON maps an admin list entry.
func AdminUserToJSON(e domainuser.AdminListEntry) AdminUserJSON {
	j := AdminUserJSON{
		SnowflakeID:     e.SnowflakeID,
		Email:           e.Email,
		Name:            e.Name,
		Status:          e.Status,
		Role:            e.Role,
		LastLoginIP:     e.LastLoginIP,
		LastLoginDevice: e.LastLoginDevice,
		CreatedAt:       e.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
	if e.LastLoginAt != nil {
		s := e.LastLoginAt.UTC().Format("2006-01-02T15:04:05Z07:00")
		j.LastLoginAt = &s
	}
	return j
}

// AdminPostJSON extends PostJSON with optional deleted_at for admin listings.
type AdminPostJSON struct {
	PostJSON
	DeletedAt *string `json:"deleted_at,omitempty"`
}

// AdminPostToJSON maps a domain post for admin APIs.
func AdminPostToJSON(p *domainpost.Post) AdminPostJSON {
	base := PostToJSON(p)
	var del *string
	if p.DeletedAt != nil {
		s := p.DeletedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
		del = &s
	}
	return AdminPostJSON{PostJSON: base, DeletedAt: del}
}

// AdminPostsPageJSON is GET /admin/api/posts.
type AdminPostsPageJSON struct {
	Posts []AdminPostJSON `json:"posts"`
	Total int64           `json:"total"`
}

// OverviewJSON is GET /admin/api/overview.
type OverviewJSON struct {
	UserCount     int64  `json:"user_count,string"`
	PostCount     int64  `json:"post_count,string"`
	PostsToday    int64  `json:"posts_today,string"`
	CategoryCount *int64 `json:"category_count,string,omitempty"`
}

// AdminUsersResponse is GET /admin/api/users.
type AdminUsersResponse struct {
	Users []AdminUserJSON `json:"users"`
}
