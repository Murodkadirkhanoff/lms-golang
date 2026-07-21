package application

import (
	"context"

	userscontract "github.com/chashma/lms/internal/modules/users/contract"

	"github.com/chashma/lms/internal/modules/courses/contract"
	"github.com/chashma/lms/internal/modules/courses/domain"
)

// Service implements contract.CourseCatalog and the course use cases.
type Service struct {
	courses    CourseRepository
	categories CategoryRepository
	quizzes    QuizRepository
	reviews    ReviewRepository
	questions  QuestionRepository
	users      userscontract.UserDirectory
}

// NewService wires the courses service. Its only cross-context dependency is
// the users directory (for instructor names), matching the contract DAG.
func NewService(
	courses CourseRepository,
	categories CategoryRepository,
	quizzes QuizRepository,
	reviews ReviewRepository,
	questions QuestionRepository,
	users userscontract.UserDirectory,
) *Service {
	return &Service{courses: courses, categories: categories, quizzes: quizzes, reviews: reviews, questions: questions, users: users}
}

var _ contract.CourseCatalog = (*Service)(nil)

// Decorate fills the instructor object and UI-default fields on courses,
// resolving instructor names through the users directory.
func (s *Service) Decorate(ctx context.Context, list []contract.CourseView) error {
	idSet := map[int64]struct{}{}
	for _, c := range list {
		idSet[c.InstructorID] = struct{}{}
	}
	ids := make([]int64, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}

	byID := map[int64]userscontract.UserSummary{}
	if len(ids) > 0 {
		users, err := s.users.FindByIDs(ctx, ids)
		if err != nil {
			return err
		}
		for _, u := range users {
			byID[u.ID] = u
		}
	}

	for i := range list {
		c := &list[i]
		c.ThumbnailColor = domain.ThumbnailColor(c.ID)
		name := "Instructor"
		if u, ok := byID[c.InstructorID]; ok {
			name = u.Name
		}
		c.Instructor = &contract.Instructor{
			ID:          c.InstructorID,
			Name:        name,
			Headline:    "Instructor",
			AvatarColor: domain.AvatarColor(c.InstructorID),
		}
	}
	return nil
}

// List returns a decorated page of courses.
func (s *Service) List(ctx context.Context, f CourseFilters) ([]contract.CourseView, int, error) {
	list, total, err := s.courses.List(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	if err := s.Decorate(ctx, list); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// GetCourse returns a full course (with curriculum) by id or slug, undecorated.
func (s *Service) GetCourse(ctx context.Context, idOrSlug string) (*contract.CourseView, error) {
	return s.courses.FindByIDOrSlug(ctx, idOrSlug)
}

// ReviewsForCourse returns the latest reviews for a course.
func (s *Service) ReviewsForCourse(ctx context.Context, courseID int64, limit int) ([]contract.Review, error) {
	return s.reviews.ListForCourse(ctx, courseID, limit)
}

// CreateCourse persists a new course and its curriculum.
func (s *Service) CreateCourse(ctx context.Context, c *contract.CourseView) error {
	return s.courses.Insert(ctx, c)
}

// UpdateCourse updates a course's main fields.
func (s *Service) UpdateCourse(ctx context.Context, c *contract.CourseView) error {
	return s.courses.Update(ctx, c)
}

// ReplaceModules replaces a course's whole curriculum.
func (s *Service) ReplaceModules(ctx context.Context, courseID int64, modules []contract.Module) error {
	return s.courses.ReplaceModules(ctx, courseID, modules)
}

// DeleteCourse soft-deletes a course.
func (s *Service) DeleteCourse(ctx context.Context, id int64) error {
	return s.courses.Delete(ctx, id)
}

// --- categories ---

// ListCategories returns all categories with published-course counts.
func (s *Service) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return s.categories.List(ctx)
}

// CreateCategory persists a new category.
func (s *Service) CreateCategory(ctx context.Context, c *domain.Category) error {
	return s.categories.Insert(ctx, c)
}

// GetCategory returns a single category.
func (s *Service) GetCategory(ctx context.Context, id int64) (*domain.Category, error) {
	return s.categories.FindByID(ctx, id)
}

// UpdateCategory persists category changes.
func (s *Service) UpdateCategory(ctx context.Context, c *domain.Category) error {
	return s.categories.Update(ctx, c)
}

// DeleteCategory removes a category.
func (s *Service) DeleteCategory(ctx context.Context, id int64) error {
	return s.categories.Delete(ctx, id)
}

// --- quizzes ---

// QuizByCourse returns a course's quiz.
func (s *Service) QuizByCourse(ctx context.Context, courseID int64) (*domain.Quiz, error) {
	return s.quizzes.FindByCourseID(ctx, courseID)
}

// UpsertQuiz replaces a course's quiz.
func (s *Service) UpsertQuiz(ctx context.Context, q *domain.Quiz) error {
	return s.quizzes.Upsert(ctx, q)
}

// SubmitAttempt records a quiz attempt.
func (s *Service) SubmitAttempt(ctx context.Context, a *domain.QuizAttempt) error {
	return s.quizzes.InsertAttempt(ctx, a)
}

// ListAttempts returns a user's attempts for a course.
func (s *Service) ListAttempts(ctx context.Context, userID, courseID int64) ([]domain.QuizAttempt, error) {
	return s.quizzes.ListAttempts(ctx, userID, courseID)
}

// --- reviews ---

// UpsertReview creates or updates a course review.
func (s *Service) UpsertReview(ctx context.Context, r *contract.Review) error {
	return s.reviews.Upsert(ctx, r)
}

// --- questions ---

// ListQuestions returns recent questions for a lesson.
func (s *Service) ListQuestions(ctx context.Context, lessonID int64) ([]domain.LessonQuestion, error) {
	return s.questions.List(ctx, lessonID)
}

// AskQuestion records a learner question.
func (s *Service) AskQuestion(ctx context.Context, lessonID, userID int64, userName, question string) (domain.LessonQuestion, error) {
	return s.questions.Insert(ctx, lessonID, userID, userName, question)
}

// LookupUserName returns a user's display name, or fallback when unknown.
func (s *Service) LookupUserName(ctx context.Context, id int64, fallback string) string {
	users, err := s.users.FindByIDs(ctx, []int64{id})
	if err != nil || len(users) == 0 {
		return fallback
	}
	return users[0].Name
}

// --- instructors ---

// Instructors returns all instructors (users with at least one published
// course) with their aggregate stats.
func (s *Service) Instructors(ctx context.Context) ([]contract.Instructor, error) {
	stats, err := s.courses.InstructorStats(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(stats))
	for _, st := range stats {
		ids = append(ids, st.InstructorID)
	}
	names, err := s.userNames(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]contract.Instructor, 0, len(stats))
	for _, st := range stats {
		out = append(out, instructorFrom(st, names[st.InstructorID]))
	}
	return out, nil
}

// Instructor returns a single instructor by user id, or nil if the user does
// not exist.
func (s *Service) Instructor(ctx context.Context, id int64) (*contract.Instructor, error) {
	stats, err := s.courses.InstructorStats(ctx)
	if err != nil {
		return nil, err
	}
	var stat domain.InstructorStat
	stat.InstructorID = id
	for _, st := range stats {
		if st.InstructorID == id {
			stat = st
			break
		}
	}
	names, err := s.userNames(ctx, []int64{id})
	if err != nil {
		return nil, err
	}
	name, ok := names[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	inst := instructorFrom(stat, name)
	return &inst, nil
}

func (s *Service) userNames(ctx context.Context, ids []int64) (map[int64]string, error) {
	names := map[int64]string{}
	if len(ids) == 0 {
		return names, nil
	}
	users, err := s.users.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		names[u.ID] = u.Name
	}
	return names, nil
}

func instructorFrom(st domain.InstructorStat, name string) contract.Instructor {
	if name == "" {
		name = "Instructor"
	}
	return contract.Instructor{
		ID:          st.InstructorID,
		Name:        name,
		Headline:    "Instructor",
		AvatarColor: domain.AvatarColor(st.InstructorID),
		Students:    st.Students,
		Courses:     st.CourseCount,
		Rating:      st.Rating,
	}
}

// --- contract.CourseCatalog ---

// CoursesByIDs returns decorated courses for the given ids (unpublished included).
func (s *Service) CoursesByIDs(ctx context.Context, ids []int64) ([]contract.CourseView, error) {
	if len(ids) == 0 {
		return []contract.CourseView{}, nil
	}
	list, _, err := s.courses.List(ctx, CourseFilters{Page: 1, PageSize: len(ids), IDs: ids, IncludeUnpublished: true})
	if err != nil {
		return nil, err
	}
	if err := s.Decorate(ctx, list); err != nil {
		return nil, err
	}
	return list, nil
}

// CoursesByInstructor returns an instructor's courses (drafts included).
func (s *Service) CoursesByInstructor(ctx context.Context, instructorID int64) ([]contract.CourseView, error) {
	list, _, err := s.courses.List(ctx, CourseFilters{Page: 1, PageSize: 1000, InstructorID: instructorID, IncludeUnpublished: true})
	if err != nil {
		return nil, err
	}
	if err := s.Decorate(ctx, list); err != nil {
		return nil, err
	}
	return list, nil
}

// LessonsForCourse returns a course's lessons in curriculum order.
func (s *Service) LessonsForCourse(ctx context.Context, courseID int64) ([]contract.LessonInfo, error) {
	return s.courses.LessonsForCourses(ctx, []int64{courseID})
}

// LessonsForCourses returns lessons for several courses.
func (s *Service) LessonsForCourses(ctx context.Context, courseIDs []int64) ([]contract.LessonInfo, error) {
	return s.courses.LessonsForCourses(ctx, courseIDs)
}

// LessonsByIDs returns lessons by their ids (checkout).
func (s *Service) LessonsByIDs(ctx context.Context, ids []int64) ([]contract.LessonInfo, error) {
	return s.courses.LessonsByIDs(ctx, ids)
}

// IncrementStudentCount bumps a course's student counter.
func (s *Service) IncrementStudentCount(ctx context.Context, courseID int64) error {
	return s.courses.IncrementStudentCount(ctx, courseID)
}

// AvgQuizScore returns the mean attempt score across the given courses.
func (s *Service) AvgQuizScore(ctx context.Context, courseIDs []int64) (float64, error) {
	return s.quizzes.AvgScoreForCourses(ctx, courseIDs)
}

// Stats returns published-course/active-instructor counts.
func (s *Service) Stats(ctx context.Context) (contract.CourseStats, error) {
	return s.courses.Stats(ctx)
}

// CourseCountsByInstructor maps user id -> number of (non-deleted) courses.
func (s *Service) CourseCountsByInstructor(ctx context.Context, ids []int64) (map[int64]int, error) {
	return s.courses.CourseCountsByInstructor(ctx, ids)
}
