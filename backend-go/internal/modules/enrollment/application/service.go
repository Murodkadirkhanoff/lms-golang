package application

import (
	"context"
	"log/slog"

	coursescontract "github.com/chashma/lms/internal/modules/courses/contract"
	"github.com/chashma/lms/internal/modules/enrollment/contract"
	"github.com/chashma/lms/internal/modules/enrollment/domain"
	userscontract "github.com/chashma/lms/internal/modules/users/contract"
)

// Service implements contract.EnrollmentGate and the enrollment use cases.
type Service struct {
	enrollments   EnrollmentRepository
	orders        OrderRepository
	certificates  CertificateRepository
	notifications NotificationRepository
	certRenderer  CertificateRenderer
	catalog       coursescontract.CourseCatalog
	users         userscontract.UserDirectory
}

// NewService wires the enrollment service. It depends on the courses catalog
// and users directory only through their contracts.
func NewService(
	enrollments EnrollmentRepository,
	orders OrderRepository,
	certificates CertificateRepository,
	notifications NotificationRepository,
	certRenderer CertificateRenderer,
	catalog coursescontract.CourseCatalog,
	users userscontract.UserDirectory,
) *Service {
	return &Service{
		enrollments: enrollments, orders: orders, certificates: certificates,
		notifications: notifications, certRenderer: certRenderer, catalog: catalog, users: users,
	}
}

var _ contract.EnrollmentGate = (*Service)(nil)

// --- contract.EnrollmentGate ---

// Revenue is the sum of all paid orders.
func (s *Service) Revenue(ctx context.Context) (float64, error) {
	return s.orders.Revenue(ctx)
}

// EnrollmentCountsByUser maps user id -> enrollment count.
func (s *Service) EnrollmentCountsByUser(ctx context.Context, userIDs []int64) (map[int64]int, error) {
	return s.enrollments.EnrollmentCountsByUser(ctx, userIDs)
}

// AccessibleLessonIDs lists lesson ids a user may view in a course.
func (s *Service) AccessibleLessonIDs(ctx context.Context, userID, courseID int64) ([]int64, error) {
	return s.enrollments.AccessibleLessonIDs(ctx, userID, courseID)
}

// IsEnrolled reports whether the user is enrolled in the course.
func (s *Service) IsEnrolled(ctx context.Context, userID, courseID int64) (bool, error) {
	owned, err := s.enrollments.OwnedCourses(ctx, userID, []int64{courseID})
	if err != nil {
		return false, err
	}
	return owned[courseID], nil
}

// --- orchestration helpers ---

// grantCourseAccess grants a user access to every lesson in a course.
func (s *Service) grantCourseAccess(ctx context.Context, userID, courseID int64) error {
	lessons, err := s.catalog.LessonsForCourse(ctx, courseID)
	if err != nil {
		return err
	}
	ids := make([]int64, 0, len(lessons))
	for _, l := range lessons {
		ids = append(ids, l.ID)
	}
	return s.enrollments.GrantLessonAccess(ctx, userID, courseID, ids)
}

// markEnrolled bumps the course's student counter; failures do not abort.
func (s *Service) markEnrolled(ctx context.Context, courseID int64) {
	if err := s.catalog.IncrementStudentCount(ctx, courseID); err != nil {
		slog.Warn("failed to increment student count", "courseId", courseID, "err", err)
	}
}

// notify records an in-app notification; failures do not abort the request.
func (s *Service) notify(ctx context.Context, userID int64, ntype, title, body string) {
	if err := s.notifications.Insert(ctx, userID, ntype, title, body); err != nil {
		slog.Warn("failed to insert notification", "err", err)
	}
}

// maybeIssueCertificate issues a certificate when a course is fully completed.
func (s *Service) maybeIssueCertificate(ctx context.Context, userID, courseID int64) {
	courses, err := s.catalog.CoursesByIDs(ctx, []int64{courseID})
	if err != nil || len(courses) == 0 {
		if err != nil {
			slog.Warn("certificate check: failed to fetch course", "err", err)
		}
		return
	}
	course := courses[0]
	if course.TotalLessons == 0 {
		return
	}
	counts, err := s.enrollments.CompletedCounts(ctx, userID)
	if err != nil {
		slog.Warn("certificate check: failed to count lessons", "err", err)
		return
	}
	if counts[courseID] < course.TotalLessons {
		return
	}
	issued, err := s.certificates.Issue(ctx, userID, courseID, course.Title)
	if err != nil {
		slog.Warn("failed to issue certificate", "err", err)
		return
	}
	if issued {
		s.notify(ctx, userID, "course", "Certificate earned",
			"Congratulations! You completed \""+course.Title+"\" and earned a certificate.")
	}
}

// --- enrollment use cases ---

// Enroll enrolls a user in a course and grants access to all its lessons.
func (s *Service) Enroll(ctx context.Context, userID int64, course coursescontract.CourseView) (domain.Enrollment, error) {
	enr, isNew, err := s.enrollments.Insert(ctx, userID, course.ID)
	if err != nil {
		return domain.Enrollment{}, err
	}
	if err := s.grantCourseAccess(ctx, userID, course.ID); err != nil {
		return domain.Enrollment{}, err
	}
	if isNew {
		s.markEnrolled(ctx, course.ID)
		s.notify(ctx, userID, "course", "Enrolled in a course",
			"You are now enrolled in \""+course.Title+"\". Happy learning!")
	}
	return enr, nil
}

// FindEnrollment returns an enrollment by id.
func (s *Service) FindEnrollment(ctx context.Context, id int64) (*domain.Enrollment, error) {
	return s.enrollments.FindByID(ctx, id)
}

// UpdateProgress marks a lesson complete/incomplete and, when completing,
// issues a certificate if the course is now finished.
func (s *Service) UpdateProgress(ctx context.Context, userID, lessonID int64, completed bool, courseID int64) error {
	if err := s.enrollments.SetLessonCompleted(ctx, userID, lessonID, completed); err != nil {
		return err
	}
	if completed {
		s.maybeIssueCertificate(ctx, userID, courseID)
	}
	return nil
}

// CourseByID returns a single decorated course, or nil if not found.
func (s *Service) CourseByID(ctx context.Context, id int64) (*coursescontract.CourseView, error) {
	courses, err := s.catalog.CoursesByIDs(ctx, []int64{id})
	if err != nil {
		return nil, err
	}
	if len(courses) == 0 {
		return nil, nil
	}
	return &courses[0], nil
}
