package post

import (
	"context"
	"errors"
	"strings"

	appcategory "github.com/lpxxn/blink/application/category"
	appmoderation "github.com/lpxxn/blink/application/moderation"
	domaincategory "github.com/lpxxn/blink/domain/category"
	domainpost "github.com/lpxxn/blink/domain/post"
)

const (
	maxBodyLen = 20000
	maxImages  = 9
)

var (
	ErrForbidden      = errors.New("post: forbidden")
	ErrInvalidInput   = errors.New("post: invalid input")
	ErrNotVisible     = errors.New("post: not visible to viewer")
)

type Service struct {
	Posts      domainpost.Repository
	Categories domaincategory.Repository
	NewID      func() int64
}

func (s *Service) validateCategory(ctx context.Context, categoryID *int64) error {
	if categoryID == nil {
		return nil
	}
	_, err := s.Categories.GetByID(ctx, *categoryID)
	if err != nil {
		if errors.Is(err, domaincategory.ErrNotFound) {
			return appcategory.ErrInvalidCategory
		}
		return err
	}
	return nil
}

func (s *Service) validateBodyImages(body string, images []string) error {
	if len(body) > maxBodyLen {
		return ErrInvalidInput
	}
	if len(images) > maxImages {
		return ErrInvalidInput
	}
	return nil
}

// Create publishes or saves a draft original post (public visibility).
func (s *Service) Create(ctx context.Context, authorID int64, body string, categoryID *int64, images []string, draft bool) (*domainpost.Post, error) {
	body = strings.TrimSpace(body)
	if err := s.validateBodyImages(body, images); err != nil {
		return nil, err
	}
	if err := s.validateCategory(ctx, categoryID); err != nil {
		return nil, err
	}
	status := domainpost.StatusPublished
	if draft {
		status = domainpost.StatusDraft
	}
	words := appmoderation.SensitiveWords()
	hits := appmoderation.FindSensitiveHits(body, words)
	modFlag, modNote := appmoderation.PostModerationFromHits(hits)
	p := &domainpost.Post{
		ID:             s.NewID(),
		UserID:         authorID,
		PostType:       domainpost.TypeOriginal,
		Visibility:     domainpost.VisibilityPublic,
		CategoryID:     categoryID,
		Body:           body,
		Images:         images,
		Status:         status,
		ModerationFlag: modFlag,
		ModerationNote: modNote,
	}
	if err := s.Posts.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Patch updates the author's post.
type Patch struct {
	Body          *string
	CategoryID    *int64
	ClearCategory bool
	Images        *[]string
	Status        *int
}

func (s *Service) Patch(ctx context.Context, authorID, postID int64, patch Patch) (*domainpost.Post, error) {
	p, err := s.Posts.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if p.UserID != authorID {
		return nil, ErrForbidden
	}
	if p.DeletedAt != nil {
		return nil, domainpost.ErrNotFound
	}
	if patch.Body != nil {
		b := strings.TrimSpace(*patch.Body)
		if err := s.validateBodyImages(b, p.Images); err != nil {
			return nil, err
		}
		p.Body = b
	}
	if patch.Images != nil {
		if err := s.validateBodyImages(p.Body, *patch.Images); err != nil {
			return nil, err
		}
		p.Images = *patch.Images
	}
	if patch.ClearCategory {
		p.CategoryID = nil
	} else if patch.CategoryID != nil {
		if err := s.validateCategory(ctx, patch.CategoryID); err != nil {
			return nil, err
		}
		p.CategoryID = patch.CategoryID
	}
	if patch.Status != nil {
		switch *patch.Status {
		case domainpost.StatusDraft, domainpost.StatusPublished, domainpost.StatusHidden:
			p.Status = *patch.Status
		default:
			return nil, ErrInvalidInput
		}
	}
	// Re-evaluate sensitive-word moderation when body or images may have changed.
	w := appmoderation.SensitiveWords()
	h := appmoderation.FindSensitiveHits(p.Body, w)
	p.ModerationFlag, p.ModerationNote = appmoderation.PostModerationFromHits(h)
	if err := s.Posts.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) Delete(ctx context.Context, authorID, postID int64) error {
	p, err := s.Posts.GetByID(ctx, postID)
	if err != nil {
		return err
	}
	if p.UserID != authorID {
		return ErrForbidden
	}
	if p.DeletedAt != nil {
		return domainpost.ErrNotFound
	}
	return s.Posts.SoftDelete(ctx, postID)
}

// GetPublic returns a post visible on the public feed (or ErrNotVisible).
func (s *Service) GetPublic(ctx context.Context, postID int64) (*domainpost.Post, error) {
	p, err := s.Posts.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if p.DeletedAt != nil {
		return nil, domainpost.ErrNotFound
	}
	if p.Status != domainpost.StatusPublished || p.ModerationFlag != domainpost.ModerationNormal {
		return nil, ErrNotVisible
	}
	if p.PostType != domainpost.TypeOriginal || p.Visibility != domainpost.VisibilityPublic {
		return nil, ErrNotVisible
	}
	return p, nil
}

// GetForViewer returns a post for the optional viewer: authors always see their own non-deleted post;
// others only see public published normal posts.
func (s *Service) GetForViewer(ctx context.Context, postID int64, viewerID *int64) (*domainpost.Post, error) {
	p, err := s.Posts.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if p.DeletedAt != nil {
		return nil, domainpost.ErrNotFound
	}
	if viewerID != nil && *viewerID == p.UserID {
		return p, nil
	}
	if p.Status != domainpost.StatusPublished || p.ModerationFlag != domainpost.ModerationNormal {
		return nil, ErrNotVisible
	}
	if p.PostType != domainpost.TypeOriginal || p.Visibility != domainpost.VisibilityPublic {
		return nil, ErrNotVisible
	}
	return p, nil
}

// GetForAuthor returns the post if owned by author (including drafts); soft-deleted yields ErrNotFound.
func (s *Service) GetForAuthor(ctx context.Context, authorID, postID int64) (*domainpost.Post, error) {
	p, err := s.Posts.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if p.DeletedAt != nil {
		return nil, domainpost.ErrNotFound
	}
	if p.UserID != authorID {
		return nil, ErrForbidden
	}
	return p, nil
}

// ListFeed returns newest-first published posts for the public timeline.
func (s *Service) ListFeed(ctx context.Context, categoryID *int64, uncategorizedOnly bool, beforeID *int64, limit int) ([]*domainpost.Post, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.Posts.ListPublicFeed(ctx, categoryID, uncategorizedOnly, beforeID, limit)
}

// ListMine returns the author's posts (newest first).
func (s *Service) ListMine(ctx context.Context, userID int64, includeDraft bool, beforeID *int64, limit int) ([]*domainpost.Post, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.Posts.ListByUserID(ctx, userID, includeDraft, beforeID, limit)
}
