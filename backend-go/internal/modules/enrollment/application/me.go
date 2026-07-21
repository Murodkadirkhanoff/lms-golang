package application

import (
	"context"
	"fmt"

	coursescontract "github.com/chashma/lms/internal/modules/courses/contract"
	"github.com/chashma/lms/internal/modules/enrollment/domain"
)

func (s *Service) coursesByID(ctx context.Context, ids []int64) (map[int64]coursescontract.CourseView, error) {
	m := map[int64]coursescontract.CourseView{}
	if len(ids) == 0 {
		return m, nil
	}
	courses, err := s.catalog.CoursesByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, c := range courses {
		m[c.ID] = c
	}
	return m, nil
}

// Dashboard returns the learner dashboard summary.
func (s *Service) Dashboard(ctx context.Context, userID int64) (DashboardStats, error) {
	enrolls, err := s.enrollments.ListByUser(ctx, userID)
	if err != nil {
		return DashboardStats{}, err
	}
	courseIDs := courseIDsOf(enrolls)
	courses, err := s.coursesByID(ctx, courseIDs)
	if err != nil {
		return DashboardStats{}, err
	}
	completed, err := s.enrollments.CompletedCounts(ctx, userID)
	if err != nil {
		return DashboardStats{}, err
	}

	inProgress, completedCourses := 0, 0
	for _, e := range enrolls {
		course, ok := courses[e.CourseID]
		if !ok || course.TotalLessons == 0 {
			continue
		}
		done := completed[e.CourseID]
		if done >= course.TotalLessons {
			completedCourses++
		} else if done > 0 {
			inProgress++
		}
	}

	certs, err := s.certificates.CountByUser(ctx, userID)
	if err != nil {
		return DashboardStats{}, err
	}
	return DashboardStats{
		Enrolled: len(enrolls), InProgress: inProgress, Completed: completedCourses, Certificates: certs,
	}, nil
}

// MyCourses returns a page of the learner's course library with progress.
func (s *Service) MyCourses(ctx context.Context, userID int64, page, pageSize int) ([]EnrolledCourseView, int, error) {
	enrolls, total, err := s.enrollments.PageByUser(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	courseIDs := courseIDsOf(enrolls)
	courses, err := s.coursesByID(ctx, courseIDs)
	if err != nil {
		return nil, 0, err
	}
	completed, err := s.enrollments.CompletedCounts(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	lessonsByCourse := map[int64][]coursescontract.LessonInfo{}
	lessons, err := s.catalog.LessonsForCourses(ctx, courseIDs)
	if err != nil {
		return nil, 0, err
	}
	for _, l := range lessons {
		lessonsByCourse[l.CourseID] = append(lessonsByCourse[l.CourseID], l)
	}
	doneLessons, err := s.enrollments.CompletedLessonIDs(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	items := make([]EnrolledCourseView, 0, len(enrolls))
	for _, e := range enrolls {
		course, ok := courses[e.CourseID]
		if !ok {
			continue // course may have been deleted
		}
		done := completed[e.CourseID]
		progress := 0
		if course.TotalLessons > 0 {
			progress = done * 100 / course.TotalLessons
		}
		currentLesson := ""
		completedIDs := []int64{}
		for _, l := range lessonsByCourse[e.CourseID] {
			if doneLessons[l.ID] {
				completedIDs = append(completedIDs, l.ID)
			} else if currentLesson == "" {
				currentLesson = l.Title
			}
		}
		items = append(items, EnrolledCourseView{
			EnrollmentID: e.ID, Course: course, Progress: progress,
			CurrentLesson: currentLesson, LessonsCompleted: done, CompletedLessonIDs: completedIDs,
		})
	}
	return items, total, nil
}

// Certificates returns a user's certificates.
func (s *Service) Certificates(ctx context.Context, userID int64) ([]domain.Certificate, error) {
	return s.certificates.ListByUser(ctx, userID)
}

// CertificateForUser returns a single certificate owned by the user.
func (s *Service) CertificateForUser(ctx context.Context, id, userID int64) (*domain.Certificate, error) {
	return s.certificates.FindForUser(ctx, id, userID)
}

// RenderCertificate returns a downloadable PDF for a user's certificate.
func (s *Service) RenderCertificate(ctx context.Context, id, userID int64) ([]byte, string, error) {
	cert, err := s.certificates.FindForUser(ctx, id, userID)
	if err != nil {
		return nil, "", err
	}
	studentName := s.LookupUserName(ctx, userID, "Student")
	pdf, err := s.certRenderer.Render(studentName, *cert)
	if err != nil {
		return nil, "", err
	}
	return pdf, fmt.Sprintf("certificate-LH-%06d.pdf", cert.ID), nil
}

// Notifications returns a user's notifications.
func (s *Service) Notifications(ctx context.Context, userID int64) ([]domain.Notification, error) {
	return s.notifications.ListByUser(ctx, userID)
}

// MarkNotificationRead marks one notification read.
func (s *Service) MarkNotificationRead(ctx context.Context, id, userID int64) error {
	return s.notifications.MarkRead(ctx, id, userID)
}

// MarkAllNotificationsRead marks all a user's notifications read.
func (s *Service) MarkAllNotificationsRead(ctx context.Context, userID int64) error {
	return s.notifications.MarkAllRead(ctx, userID)
}

// Orders returns a page of a user's orders.
func (s *Service) Orders(ctx context.Context, userID int64, page, pageSize int) ([]domain.Order, int, error) {
	return s.orders.ListByUser(ctx, userID, page, pageSize)
}

// Order returns a single order owned by the user.
func (s *Service) Order(ctx context.Context, id, userID int64) (*domain.Order, error) {
	return s.orders.FindForUser(ctx, id, userID)
}

// LookupUserName returns a user's display name, or fallback when unknown.
func (s *Service) LookupUserName(ctx context.Context, id int64, fallback string) string {
	users, err := s.users.FindByIDs(ctx, []int64{id})
	if err != nil || len(users) == 0 {
		return fallback
	}
	return users[0].Name
}

func courseIDsOf(enrolls []domain.Enrollment) []int64 {
	ids := make([]int64, 0, len(enrolls))
	for _, e := range enrolls {
		ids = append(ids, e.CourseID)
	}
	return ids
}
