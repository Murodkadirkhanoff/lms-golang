package main

import (
	"net/http"

	"lms.chashma.uz/pkg/jsonutil"
	"lms.chashma.uz/pkg/middleware"
)

// courseEngagement — studio analytics "completion by course" qatori.
type courseEngagement struct {
	CourseID   int64  `json:"courseId"`
	Title      string `json:"title"`
	Students   int    `json:"students"`
	Completion int    `json:"completion"` // 0-100 %
}

// meTeachingStatsHandler — GET /v1/me/teaching/stats. Instruktor (joriy user)
// kurslari bo'yicha studio dashboard/analytics ko'rsatkichlari. Kurslar
// ro'yxati course-service'dan, enrollments/orders shu servisning DB'sidan.
func (app *application) meTeachingStatsHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ContextGetUser(r)

	courses, err := app.fetchCoursesByInstructor(r.Context(), claims.UserID)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	courseIDs := make([]int64, 0, len(courses))
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

	// Alohida sotiladigan darslar ham daromadga kiradi.
	lessonsByCourse, err := app.fetchLessonsByCourses(r.Context(), courseIDs)
	if err != nil {
		app.logger.Warn("teaching stats: failed to fetch lessons", "error", err.Error())
		lessonsByCourse = map[int64][]*lessonInfo{}
	}
	lessonIDs := []int64{}
	for _, lessons := range lessonsByCourse {
		for _, l := range lessons {
			lessonIDs = append(lessonIDs, l.ID)
		}
	}

	totalRevenue, monthly, err := app.models.Orders.RevenueForItems(courseIDs, lessonIDs)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	studentCounts, err := app.models.Enrollments.CountsByCourses(courseIDs)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	totalStudents, err := app.models.Enrollments.DistinctStudentsForCourses(courseIDs)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	completedCounts, activeStudents, err := app.models.Enrollments.CompletedStatsByCourses(courseIDs)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	avgQuizScore, err := app.fetchQuizAvgScore(r.Context(), courseIDs)
	if err != nil {
		app.logger.Warn("teaching stats: failed to fetch quiz stats", "error", err.Error())
	}

	// Har kurs bo'yicha tugatilganlik: tugatilgan dars yozuvlari /
	// (studentlar * darslar soni).
	engagement := []courseEngagement{}
	completionSum, completionCourses := 0, 0
	for _, c := range courses {
		students := studentCounts[c.ID]
		completion := 0
		if students > 0 && c.TotalLessons > 0 {
			completion = completedCounts[c.ID] * 100 / (students * c.TotalLessons)
		}
		if students > 0 {
			completionSum += completion
			completionCourses++
		}
		engagement = append(engagement, courseEngagement{
			CourseID:   c.ID,
			Title:      c.Title,
			Students:   students,
			Completion: completion,
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

	stats := jsonutil.Envelope{
		"totalRevenue":     totalRevenue,
		"monthlyRevenue":   monthly,
		"totalStudents":    totalStudents,
		"activeStudents":   activeStudents,
		"publishedCourses": published,
		"draftCourses":     drafts,
		"avgRating":        avgRating,
		"avgCompletion":    avgCompletion,
		"avgQuizScore":     avgQuizScore,
		"engagement":       engagement,
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, stats, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}
