package domain

import (
	"fmt"

	"github.com/chashma/lms/internal/modules/courses/contract"
	"github.com/chashma/lms/internal/platform/web"
)

// ValidateCourse checks a course and its curriculum. Messages match the
// frontend expectations byte-for-byte.
func ValidateCourse(v *web.Validator, c *contract.CourseView) {
	v.Check(c.Title != "", "title", "must be provided")
	v.Check(web.ByteLength(c.Title) <= 200, "title", "must not be more than 200 bytes long")
	v.Check(web.Permitted(c.Lang, "uz", "ru", "en"), "lang", "must be one of uz, ru, en")
	v.Check(c.Price >= 0, "price", "must not be negative")

	for mi, m := range c.Modules {
		v.Check(m.Title != "", fmt.Sprintf("modules[%d].title", mi), "must be provided")
		for li, l := range m.Lessons {
			key := fmt.Sprintf("modules[%d].lessons[%d]", mi, li)
			v.Check(l.Title != "", key+".title", "must be provided")
			v.Check(web.Permitted(l.Type, "video", "text"), key+".type", "must be one of video, text")
			v.Check(l.Price >= 0, key+".price", "must not be negative")
			v.Check(!l.IsFree || l.Price == 0, key+".price", "free lessons must have price 0")
			v.Check(l.DurationSeconds >= 0, key+".durationSeconds", "must not be negative")
		}
	}
}

// ValidateCategory checks a category's names.
func ValidateCategory(v *web.Validator, nameUz, nameRu, nameEn string) {
	v.Check(nameUz != "", "name_uz", "must be provided")
	v.Check(web.ByteLength(nameUz) <= 100, "name_uz", "must not be more than 100 bytes long")
	v.Check(nameRu != "", "name_ru", "must be provided")
	v.Check(web.ByteLength(nameRu) <= 100, "name_ru", "must not be more than 100 bytes long")
	v.Check(nameEn != "", "name_en", "must be provided")
	v.Check(web.ByteLength(nameEn) <= 100, "name_en", "must not be more than 100 bytes long")
}

// ValidateQuiz checks a quiz and its questions.
func ValidateQuiz(v *web.Validator, q *Quiz) {
	v.Check(q.Title != "", "title", "must be provided")
	v.Check(q.PassingScore >= 0 && q.PassingScore <= 100, "passing_score", "must be between 0 and 100")
	v.Check(q.TimeLimitMinutes > 0, "time_limit_minutes", "must be greater than zero")
	v.Check(len(q.Questions) > 0, "questions", "must contain at least one question")

	for i, qq := range q.Questions {
		key := fmt.Sprintf("questions[%d]", i)
		v.Check(qq.Question != "", key+".question", "must be provided")
		v.Check(len(qq.Options) >= 2, key+".options", "must have at least 2 options")
		v.Check(qq.CorrectIndex >= 0 && qq.CorrectIndex < len(qq.Options), key+".correct_index", "must point to an option")
	}
}
