package gormdb

import (
	"context"
	"errors"
	"time"

	domainpost "github.com/lpxxn/blink/domain/post"
	"gorm.io/gorm"
)

type PostRepository struct {
	DB *gorm.DB
}

func postModelToDomain(m *PostModel) (*domainpost.Post, error) {
	imgs, err := decodeImageSlice(m.Images)
	if err != nil {
		return nil, err
	}
	var deleted *time.Time
	if m.DeletedAt.Valid {
		t := m.DeletedAt.Time
		deleted = &t
	}
	return &domainpost.Post{
		ID:               m.ID,
		UserID:           m.UserID,
		PostType:         m.PostType,
		ReplyToPostID:    m.ReplyToPostID,
		ReferencedPostID: m.ReferencedPostID,
		Visibility:       m.Visibility,
		AudienceListID:   m.AudienceListID,
		CategoryID:       m.CategoryID,
		Body:             m.Body,
		Images:           imgs,
		Status:           m.Status,
		ModerationFlag:   m.ModerationFlag,
		ModerationNote:   m.ModerationNote,
		AppealBody:       m.AppealBody,
		AppealStatus:     m.AppealStatus,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
		DeletedAt:        deleted,
	}, nil
}

func domainToPostModel(p *domainpost.Post) (*PostModel, error) {
	imgJSON, err := encodeImageSlice(p.Images)
	if err != nil {
		return nil, err
	}
	return &PostModel{
		ID:               p.ID,
		UserID:           p.UserID,
		PostType:         p.PostType,
		ReplyToPostID:    p.ReplyToPostID,
		ReferencedPostID: p.ReferencedPostID,
		Visibility:       p.Visibility,
		AudienceListID:   p.AudienceListID,
		CategoryID:       p.CategoryID,
		Body:             p.Body,
		Images:           imgJSON,
		Status:           p.Status,
		ModerationFlag:   p.ModerationFlag,
		ModerationNote:   p.ModerationNote,
		AppealBody:       p.AppealBody,
		AppealStatus:     p.AppealStatus,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}, nil
}

func (r *PostRepository) Create(ctx context.Context, p *domainpost.Post) error {
	m, err := domainToPostModel(p)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = now
	}
	return r.DB.WithContext(ctx).Create(m).Error
}

func (r *PostRepository) Update(ctx context.Context, p *domainpost.Post) error {
	m, err := domainToPostModel(p)
	if err != nil {
		return err
	}
	m.UpdatedAt = time.Now().UTC()
	return r.DB.WithContext(ctx).Model(&PostModel{}).Where("id = ?", p.ID).Updates(map[string]interface{}{
		"user_id":            m.UserID,
		"post_type":          m.PostType,
		"reply_to_post_id":   m.ReplyToPostID,
		"referenced_post_id": m.ReferencedPostID,
		"visibility":         m.Visibility,
		"audience_list_id":   m.AudienceListID,
		"category_id":        m.CategoryID,
		"body":               m.Body,
		"images":             m.Images,
		"status":             m.Status,
		"moderation_flag":    m.ModerationFlag,
		"moderation_note":    m.ModerationNote,
		"appeal_body":        m.AppealBody,
		"appeal_status":      m.AppealStatus,
		"updated_at":         m.UpdatedAt,
	}).Error
}

func (r *PostRepository) SoftDelete(ctx context.Context, id int64) error {
	return r.DB.WithContext(ctx).Where("id = ?", id).Delete(&PostModel{}).Error
}

func (r *PostRepository) GetByID(ctx context.Context, id int64) (*domainpost.Post, error) {
	var m PostModel
	err := r.DB.WithContext(ctx).Unscoped().Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domainpost.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return postModelToDomain(&m)
}

func (r *PostRepository) publicFeedBaseQuery(ctx context.Context) *gorm.DB {
	return r.DB.WithContext(ctx).Model(&PostModel{}).
		Where("posts.deleted_at IS NULL AND posts.status = ? AND posts.moderation_flag = ? AND posts.post_type = ? AND posts.visibility = ?",
			domainpost.StatusPublished, domainpost.ModerationNormal, domainpost.TypeOriginal, domainpost.VisibilityPublic)
}

func applyBlockFilter(q *gorm.DB, viewerID int64) *gorm.DB {
	if viewerID == 0 {
		return q
	}
	return q.Where(`posts.user_id NOT IN (
		SELECT blocked_id FROM user_blocks WHERE blocker_id = ? AND deleted_at IS NULL
		UNION
		SELECT blocker_id FROM user_blocks WHERE blocked_id = ? AND deleted_at IS NULL
	)`, viewerID, viewerID)
}

func derefViewerID(viewerID *int64) int64 {
	if viewerID == nil {
		return 0
	}
	return *viewerID
}

func (r *PostRepository) ListPublicFeed(ctx context.Context, categoryID *int64, uncategorizedOnly bool, beforeID *int64, limit int, viewerID *int64) ([]*domainpost.Post, error) {
	q := applyBlockFilter(r.publicFeedBaseQuery(ctx), derefViewerID(viewerID))
	if categoryID != nil {
		q = q.Where("category_id = ?", *categoryID)
	} else if uncategorizedOnly {
		q = q.Where("category_id IS NULL")
	}
	if beforeID != nil {
		q = q.Where("id < ?", *beforeID)
	}
	var rows []PostModel
	if err := q.Order("id DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domainpost.Post, 0, len(rows))
	for i := range rows {
		p, err := postModelToDomain(&rows[i])
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *PostRepository) ListFollowingFeed(ctx context.Context, followerID int64, beforeID *int64, limit int) ([]*domainpost.Post, error) {
	q := applyBlockFilter(r.publicFeedBaseQuery(ctx), followerID).
		Joins("INNER JOIN user_follows uf ON uf.followee_id = posts.user_id AND uf.follower_id = ? AND uf.deleted_at IS NULL", followerID)
	if beforeID != nil {
		q = q.Where("posts.id < ?", *beforeID)
	}
	var rows []PostModel
	if err := q.Order("posts.id DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domainpost.Post, 0, len(rows))
	for i := range rows {
		p, err := postModelToDomain(&rows[i])
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *PostRepository) SearchPublic(ctx context.Context, query string, beforeID *int64, limit int, viewerID *int64) ([]*domainpost.Post, error) {
	pattern := likeContainsPattern(query)
	q := applyBlockFilter(r.publicFeedBaseQuery(ctx), derefViewerID(viewerID)).Where("posts.body LIKE ? ESCAPE '\\'", pattern)
	if beforeID != nil {
		q = q.Where("posts.id < ?", *beforeID)
	}
	var rows []PostModel
	if err := q.Order("posts.id DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domainpost.Post, 0, len(rows))
	for i := range rows {
		p, err := postModelToDomain(&rows[i])
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *PostRepository) ListPublicByUserID(ctx context.Context, userID int64, beforeID *int64, limit int) ([]*domainpost.Post, error) {
	q := r.DB.WithContext(ctx).Model(&PostModel{}).
		Where("user_id = ? AND deleted_at IS NULL AND status = ? AND moderation_flag = ? AND post_type = ? AND visibility = ?",
			userID, domainpost.StatusPublished, domainpost.ModerationNormal, domainpost.TypeOriginal, domainpost.VisibilityPublic)
	if beforeID != nil {
		q = q.Where("id < ?", *beforeID)
	}
	var rows []PostModel
	if err := q.Order("id DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domainpost.Post, 0, len(rows))
	for i := range rows {
		p, err := postModelToDomain(&rows[i])
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *PostRepository) ListByUserID(ctx context.Context, userID int64, includeDraft bool, beforeID *int64, limit int) ([]*domainpost.Post, error) {
	q := r.DB.WithContext(ctx).Model(&PostModel{}).Where("user_id = ? AND deleted_at IS NULL", userID)
	if !includeDraft {
		q = q.Where("status = ?", domainpost.StatusPublished)
	}
	if beforeID != nil {
		q = q.Where("id < ?", *beforeID)
	}
	var rows []PostModel
	if err := q.Order("id DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domainpost.Post, 0, len(rows))
	for i := range rows {
		p, err := postModelToDomain(&rows[i])
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *PostRepository) adminBaseQuery(ctx context.Context, f domainpost.AdminListFilters) *gorm.DB {
	var q *gorm.DB
	if f.IncludeDeleted {
		q = r.DB.WithContext(ctx).Unscoped().Model(&PostModel{})
	} else {
		q = r.DB.WithContext(ctx).Model(&PostModel{})
	}
	if f.UserID != nil {
		q = q.Where("user_id = ?", *f.UserID)
	}
	if f.CategoryID != nil {
		q = q.Where("category_id = ?", *f.CategoryID)
	}
	if f.ModerationFlag != nil {
		q = q.Where("moderation_flag = ?", *f.ModerationFlag)
	}
	if f.AppealPending {
		q = q.Where("appeal_status = ?", domainpost.AppealPending)
	}
	if f.SensitiveHitPending {
		q = q.Where("moderation_flag = ?", domainpost.ModerationFlagged).
			Where("(moderation_note LIKE ? OR moderation_note LIKE ?)", "有敏感词：%", "sensitive_hit:%")
	}
	return q
}

func (r *PostRepository) CountAdmin(ctx context.Context, f domainpost.AdminListFilters) (int64, error) {
	var n int64
	err := r.adminBaseQuery(ctx, f).Count(&n).Error
	return n, err
}

func (r *PostRepository) AdminList(ctx context.Context, f domainpost.AdminListFilters, offset, limit int) ([]*domainpost.Post, int64, error) {
	var total int64
	if err := r.adminBaseQuery(ctx, f).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []PostModel
	if err := r.adminBaseQuery(ctx, f).Order("id DESC").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]*domainpost.Post, 0, len(rows))
	for i := range rows {
		p, err := postModelToDomain(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, p)
	}
	return out, total, nil
}

func (r *PostRepository) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.DB.WithContext(ctx).Model(&PostModel{}).Where("deleted_at IS NULL").Count(&n).Error
	return n, err
}

func (r *PostRepository) CountCreatedSince(ctx context.Context, t time.Time) (int64, error) {
	var n int64
	err := r.DB.WithContext(ctx).Model(&PostModel{}).
		Where("deleted_at IS NULL AND created_at >= ?", t).
		Count(&n).Error
	return n, err
}

func (r *PostRepository) TopPosters(ctx context.Context, since, until time.Time, limit int) ([]domainpost.UserPostCount, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	type row struct {
		UserID    int64 `gorm:"column:user_id"`
		PostCount int64 `gorm:"column:post_count"`
	}
	var rows []row
	err := r.DB.WithContext(ctx).Model(&PostModel{}).
		Select("user_id, COUNT(*) AS post_count").
		Where("deleted_at IS NULL AND status = ? AND created_at >= ? AND created_at < ?",
			domainpost.StatusPublished, since, until).
		Group("user_id").
		Order("post_count DESC, user_id ASC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domainpost.UserPostCount, len(rows))
	for i, r := range rows {
		out[i] = domainpost.UserPostCount{UserID: r.UserID, PostCount: r.PostCount}
	}
	return out, nil
}
