package application

import (
	"context"
	"log/slog"
)

// TeachingStats computes the instructor studio dashboard for a user's courses.
func (s *Service) TeachingStats(ctx context.Context, userID int64) (TeachingStats, error) {
	courses, err := s.catalog.CoursesByInstructor(ctx, userID)
	if err != nil {
		return TeachingStats{}, err
	}

	var courseIDs []int64
	published, drafts := 0, 0
	ratingSum, ratingCount := 0.0, 0
	for _, c := range courses {
		courseIDs = append(courseIDs, c.ID)
		if c.IsPublished {
			published++
		} else {
			drafts++
		}
		if c.RatingCount > 0 {
			ratingSum += c.Rating
			ratingCount++
		}
	}

	var lessonIDs []int64
	if lessons, err := s.catalog.LessonsForCourses(ctx, courseIDs); err != nil {
		slog.Warn("teaching stats: failed to fetch lessons", "err", err)
	} else {
		for _, l := range lessons {
			lessonIDs = append(lessonIDs, l.ID)
		}
	}

	revenue, err := s.orders.RevenueForItems(ctx, courseIDs, lessonIDs)
	if err != nil {
		return TeachingStats{}, err
	}
	studentCounts, err := s.enrollments.CountsByCourses(ctx, courseIDs)
	if err != nil {
		return TeachingStats{}, err
	}
	totalStudents, err := s.enrollments.DistinctStudentsForCourses(ctx, courseIDs)
	if err != nil {
		return TeachingStats{}, err
	}
	completedStats, err := s.enrollments.CompletedStatsByCourses(ctx, courseIDs)
	if err != nil {
		return TeachingStats{}, err
	}

	avgQuizScore := 0.0
	if q, err := s.catalog.AvgQuizScore(ctx, courseIDs); err != nil {
		slog.Warn("teaching stats: failed to fetch quiz stats", "err", err)
	} else {
		avgQuizScore = q
	}

	engagement := make([]CourseEngagement, 0, len(courses))
	completionSum, completionCourses := 0, 0
	for _, c := range courses {
		students := studentCounts[c.ID]
		completion := 0
		if students > 0 && c.TotalLessons > 0 {
			completion = completedStats.Counts[c.ID] * 100 / (students * c.TotalLessons)
		}
		if students > 0 {
			completionSum += completion
			completionCourses++
		}
		engagement = append(engagement, CourseEngagement{
			CourseID: c.ID, Title: c.Title, Students: students, Completion: completion,
		})
	}

	avgCompletion := 0
	if completionCourses > 0 {
		avgCompletion = completionSum / completionCourses
	}
	avgRating := 0.0
	if ratingCount > 0 {
		avgRating = ratingSum / float64(ratingCount)
	}

	return TeachingStats{
		TotalRevenue:     revenue.Total,
		MonthlyRevenue:   revenue.Monthly,
		TotalStudents:    totalStudents,
		ActiveStudents:   completedStats.ActiveStudents,
		PublishedCourses: published,
		DraftCourses:     drafts,
		AvgRating:        avgRating,
		AvgCompletion:    avgCompletion,
		AvgQuizScore:     avgQuizScore,
		Engagement:       engagement,
	}, nil
}
