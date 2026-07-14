package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// courseInfo course-service /internal/courses javobidan kerakli maydonlar.
// Raw — frontendga o'zgarishsiz uzatiladigan to'liq Course JSON.
type courseInfo struct {
	ID           int64   `json:"id"`
	Title        string  `json:"title"`
	Price        float64 `json:"price"`
	IsPublished  bool    `json:"isPublished"`
	TotalLessons int     `json:"totalLessons"`
	Instructor   struct {
		Name string `json:"name"`
	} `json:"instructor"`

	Raw json.RawMessage `json:"-"`
}

// lessonInfo course-service /internal/lessons va /internal/courses/{id}/lessons
// javoblaridagi elementlar.
type lessonInfo struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Price       float64 `json:"price"`
	IsFree      bool    `json:"isFree"`
	CourseID    int64   `json:"courseId"`
	CourseTitle string  `json:"courseTitle"`
}

func joinIDs(ids []int64) string {
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, strconv.FormatInt(id, 10))
	}
	return strings.Join(parts, ",")
}

// fetchCourses kurslarni batch oladi, id bo'yicha map qaytaradi.
func (app *application) fetchCourses(ctx context.Context, ids []int64) (map[int64]*courseInfo, error) {
	courses := map[int64]*courseInfo{}
	if len(ids) == 0 {
		return courses, nil
	}

	var response struct {
		Courses []json.RawMessage `json:"courses"`
	}

	err := app.courseClient.Get(ctx, "/internal/courses?ids="+joinIDs(ids), &response)
	if err != nil {
		return nil, fmt.Errorf("course-service: %w", err)
	}

	for _, raw := range response.Courses {
		var info courseInfo
		if err := json.Unmarshal(raw, &info); err != nil {
			return nil, err
		}
		info.Raw = raw
		courses[info.ID] = &info
	}

	return courses, nil
}

func (app *application) fetchCourseLessons(ctx context.Context, courseID int64) ([]*lessonInfo, error) {
	var response struct {
		Lessons []*lessonInfo `json:"lessons"`
	}

	err := app.courseClient.Get(ctx, fmt.Sprintf("/internal/courses/%d/lessons", courseID), &response)
	if err != nil {
		return nil, fmt.Errorf("course-service: %w", err)
	}

	return response.Lessons, nil
}

func (app *application) fetchLessons(ctx context.Context, ids []int64) (map[int64]*lessonInfo, error) {
	lessons := map[int64]*lessonInfo{}
	if len(ids) == 0 {
		return lessons, nil
	}

	var response struct {
		Lessons []*lessonInfo `json:"lessons"`
	}

	err := app.courseClient.Get(ctx, "/internal/lessons?ids="+joinIDs(ids), &response)
	if err != nil {
		return nil, fmt.Errorf("course-service: %w", err)
	}

	for _, l := range response.Lessons {
		lessons[l.ID] = l
	}

	return lessons, nil
}

// markEnrolled course-service'dagi student_count'ni oshiradi. Statistik
// hisoblagich bo'lgani uchun xato asosiy oqimni yiqitmaydi.
func (app *application) markEnrolled(ctx context.Context, courseID int64) {
	err := app.courseClient.Post(ctx, fmt.Sprintf("/internal/courses/%d/enrolled", courseID), nil)
	if err != nil {
		app.logger.Warn("failed to increment student count", "courseId", courseID, "error", err.Error())
	}
}

// fetchLessonsByCourses bir nechta kursning darslarini tartib bilan oladi
// (currentLesson hisoblash uchun).
func (app *application) fetchLessonsByCourses(ctx context.Context, courseIDs []int64) (map[int64][]*lessonInfo, error) {
	byCourse := map[int64][]*lessonInfo{}
	if len(courseIDs) == 0 {
		return byCourse, nil
	}

	var response struct {
		Lessons []*lessonInfo `json:"lessons"`
	}

	err := app.courseClient.Get(ctx, "/internal/courses/lessons?ids="+joinIDs(courseIDs), &response)
	if err != nil {
		return nil, fmt.Errorf("course-service: %w", err)
	}

	for _, l := range response.Lessons {
		byCourse[l.CourseID] = append(byCourse[l.CourseID], l)
	}

	return byCourse, nil
}

// grantCourseAccess kursning barcha darslariga kirish beradi
// (eski sync_lesson_access_on_enrollment triggerining o'rnini bosadi).
func (app *application) grantCourseAccess(ctx context.Context, userID, courseID int64) error {
	lessons, err := app.fetchCourseLessons(ctx, courseID)
	if err != nil {
		return err
	}

	lessonIDs := make([]int64, 0, len(lessons))
	for _, l := range lessons {
		lessonIDs = append(lessonIDs, l.ID)
	}

	return app.models.Enrollments.GrantLessonAccess(userID, courseID, lessonIDs)
}
