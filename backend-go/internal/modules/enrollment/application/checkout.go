package application

import (
	"context"
	"fmt"

	coursescontract "github.com/chashma/lms/internal/modules/courses/contract"
	"github.com/chashma/lms/internal/modules/enrollment/domain"
)

// CheckoutItem is one requested purchase (exactly one field must be set).
type CheckoutItem struct {
	CourseID *int64
	LessonID *int64
}

// Checkout prices a cart from the course catalog (never the client), records a
// paid order, and grants access. It returns field validation errors instead of
// an error when the cart is rejected. Payment is mocked: orders are paid
// immediately.
func (s *Service) Checkout(ctx context.Context, userID int64, items []CheckoutItem, paymentMethod string) (*domain.Order, map[string]string, error) {
	if paymentMethod == "" {
		paymentMethod = "card"
	}
	errs := map[string]string{}

	if len(items) == 0 {
		errs["items"] = "must contain at least one item"
		return nil, errs, nil
	}

	var courseIDs, lessonIDs []int64
	seenCourses := map[int64]bool{}
	seenLessons := map[int64]bool{}
	for i, item := range items {
		hasCourse := item.CourseID != nil && *item.CourseID > 0
		hasLesson := item.LessonID != nil && *item.LessonID > 0
		if hasCourse == hasLesson {
			addErr(errs, fmt.Sprintf("items[%d]", i), "must contain exactly one of course_id or lesson_id")
			continue
		}
		if hasCourse && !seenCourses[*item.CourseID] {
			seenCourses[*item.CourseID] = true
			courseIDs = append(courseIDs, *item.CourseID)
		}
		if hasLesson && !seenLessons[*item.LessonID] {
			seenLessons[*item.LessonID] = true
			lessonIDs = append(lessonIDs, *item.LessonID)
		}
	}
	if len(errs) > 0 {
		return nil, errs, nil
	}

	courses, err := s.coursesByID(ctx, courseIDs)
	if err != nil {
		return nil, nil, err
	}

	lessons := map[int64]coursescontract.LessonInfo{}
	lessonList, err := s.catalog.LessonsByIDs(ctx, lessonIDs)
	if err != nil {
		return nil, nil, err
	}
	for _, l := range lessonList {
		lessons[l.ID] = l
	}

	// Resolve the owning course of each purchased lesson (to block buying a
	// lesson from your own course).
	var lessonCourseIDs []int64
	for _, l := range lessons {
		if !seenCourses[l.CourseID] {
			lessonCourseIDs = append(lessonCourseIDs, l.CourseID)
		}
	}
	lessonCourses, err := s.coursesByID(ctx, lessonCourseIDs)
	if err != nil {
		return nil, nil, err
	}
	for id, c := range courses {
		lessonCourses[id] = c
	}

	ownedCourses, err := s.enrollments.OwnedCourses(ctx, userID, courseIDs)
	if err != nil {
		return nil, nil, err
	}
	ownedLessons, err := s.enrollments.OwnedLessons(ctx, userID, lessonIDs)
	if err != nil {
		return nil, nil, err
	}

	order := &domain.Order{UserID: userID, Status: "paid", PaymentMethod: paymentMethod}

	for i, id := range courseIDs {
		course, ok := courses[id]
		if !ok || !course.IsPublished {
			addErr(errs, fmt.Sprintf("items[%d].course_id", i), "course does not exist")
			continue
		}
		if course.Instructor != nil && course.Instructor.ID == userID {
			addErr(errs, fmt.Sprintf("items[%d].course_id", i), "you cannot purchase your own course")
			continue
		}
		if ownedCourses[id] {
			addErr(errs, fmt.Sprintf("items[%d].course_id", i), "you already own this course")
			continue
		}
		cid := course.ID
		order.Items = append(order.Items, domain.OrderItem{
			CourseID: &cid, CourseTitle: course.Title, Instructor: instructorName(course),
			ThumbnailColor: domain.ThumbnailColor(course.ID), Price: course.Price,
		})
	}

	for i, id := range lessonIDs {
		lesson, ok := lessons[id]
		if !ok {
			addErr(errs, fmt.Sprintf("items[%d].lesson_id", i), "lesson does not exist")
			continue
		}
		if lc, ok := lessonCourses[lesson.CourseID]; ok && lc.Instructor != nil && lc.Instructor.ID == userID {
			addErr(errs, fmt.Sprintf("items[%d].lesson_id", i), "you cannot purchase a lesson from your own course")
			continue
		}
		if ownedLessons[id] {
			addErr(errs, fmt.Sprintf("items[%d].lesson_id", i), "you already own this lesson")
			continue
		}
		lid := lesson.ID
		order.Items = append(order.Items, domain.OrderItem{
			LessonID: &lid, CourseTitle: lesson.CourseTitle,
			ThumbnailColor: domain.ThumbnailColor(lesson.CourseID), Price: lesson.Price,
		})
	}

	if len(errs) > 0 {
		return nil, errs, nil
	}

	for _, it := range order.Items {
		order.Total += it.Price
	}
	if err := s.orders.Insert(ctx, order); err != nil {
		return nil, nil, err
	}

	// Payment succeeded — grant access.
	for _, id := range courseIDs {
		if _, isNew, err := s.enrollments.Insert(ctx, userID, id); err != nil {
			return nil, nil, err
		} else if isNew {
			s.markEnrolled(ctx, id)
		}
		if err := s.grantCourseAccess(ctx, userID, id); err != nil {
			return nil, nil, err
		}
	}
	for _, id := range lessonIDs {
		lesson := lessons[id]
		if err := s.enrollments.GrantLessonAccess(ctx, userID, lesson.CourseID, []int64{lesson.ID}); err != nil {
			return nil, nil, err
		}
	}

	s.notify(ctx, userID, "system", "Purchase successful",
		fmt.Sprintf("Your order #%d has been paid. %d item(s) are now available in your library.",
			order.DBID, len(order.Items)))

	return order, nil, nil
}

func addErr(errs map[string]string, key, msg string) {
	if _, ok := errs[key]; !ok {
		errs[key] = msg
	}
}

func instructorName(c coursescontract.CourseView) string {
	if c.Instructor != nil {
		return c.Instructor.Name
	}
	return ""
}
