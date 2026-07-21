package transport

import (
	"context"

	"github.com/chashma/lms/internal/modules/courses/contract"
)

// --- request bodies (snake_case, matching the frontend) ---

type lessonRequest struct {
	Title           string  `json:"title"`
	Type            string  `json:"type"`
	ContentURL      string  `json:"content_url"`
	Content         string  `json:"content"`
	DurationSeconds int     `json:"duration_seconds"`
	Position        int     `json:"position"`
	Price           float64 `json:"price"`
	IsFree          bool    `json:"is_free"`
}

type moduleRequest struct {
	Title    string          `json:"title"`
	Position int             `json:"position"`
	Lessons  []lessonRequest `json:"lessons"`
}

type createCourseRequest struct {
	Title        string          `json:"title"`
	Description  string          `json:"description"`
	ThumbnailURL string          `json:"thumbnail_url"`
	CategoryID   *int64          `json:"category_id"`
	Lang         string          `json:"lang"`
	Price        float64         `json:"price"`
	IsPublished  bool            `json:"is_published"`
	Modules      []moduleRequest `json:"modules"`
}

type updateCourseRequest struct {
	Title        *string         `json:"title"`
	Description  *string         `json:"description"`
	ThumbnailURL *string         `json:"thumbnail_url"`
	CategoryID   *int64          `json:"category_id"`
	Lang         *string         `json:"lang"`
	Price        *float64        `json:"price"`
	IsPublished  *bool           `json:"is_published"`
	Modules      []moduleRequest `json:"modules"`
}

type categoryRequest struct {
	NameUz   *string `json:"name_uz"`
	NameRu   *string `json:"name_ru"`
	NameEn   *string `json:"name_en"`
	ParentID *int64  `json:"parent_id"`
}

// --- helpers ---

func orEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func toModules(in []moduleRequest) []contract.Module {
	modules := []contract.Module{}
	for _, m := range in {
		mod := contract.Module{Title: m.Title, Position: m.Position, Lessons: []contract.Lesson{}}
		for _, l := range m.Lessons {
			t := l.Type
			if t == "" {
				t = "video"
			}
			mod.Lessons = append(mod.Lessons, contract.Lesson{
				Title: l.Title, Type: t, ContentURL: l.ContentURL, Content: l.Content,
				DurationSeconds: l.DurationSeconds, Position: l.Position, Price: l.Price, IsFree: l.IsFree,
			})
		}
		modules = append(modules, mod)
	}
	return modules
}

// decorateOne decorates a single course through the value-slice API.
func (h *Handler) decorateOne(ctx context.Context, c *contract.CourseView) error {
	s := []contract.CourseView{*c}
	if err := h.svc.Decorate(ctx, s); err != nil {
		return err
	}
	*c = s[0]
	return nil
}
