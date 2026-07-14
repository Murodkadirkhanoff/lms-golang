package main

import (
	"errors"
	"fmt"
	"net/http"

	"lms.chashma.uz/enrollment-service/internal/data"
	"lms.chashma.uz/pkg/jsonutil"
	"lms.chashma.uz/pkg/middleware"
	"lms.chashma.uz/pkg/uidefaults"
	"lms.chashma.uz/pkg/validator"
)

// checkoutHandler — POST /v1/me/orders. Narxlar clientdan emas, course-service
// DB'sidan olinadi. To'lov hozircha mock: buyurtma darhol "paid" bo'ladi.
func (app *application) checkoutHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ContextGetUser(r)

	var input struct {
		Items []struct {
			CourseID *int64 `json:"course_id"`
			LessonID *int64 `json:"lesson_id"`
		} `json:"items"`
		PaymentMethod string `json:"payment_method"`
	}

	err := jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	v := validator.New()
	v.Check(len(input.Items) > 0, "items", "must contain at least one item")
	if input.PaymentMethod == "" {
		input.PaymentMethod = "card"
	}

	// Takroriy elementlar savatda bo'lsa ham bir marta hisoblanadi.
	courseIDs := []int64{}
	lessonIDs := []int64{}
	seenCourses := map[int64]bool{}
	seenLessons := map[int64]bool{}
	for i, item := range input.Items {
		hasCourse := item.CourseID != nil && *item.CourseID > 0
		hasLesson := item.LessonID != nil && *item.LessonID > 0
		if hasCourse == hasLesson {
			v.AddError(fmt.Sprintf("items[%d]", i), "must contain exactly one of course_id or lesson_id")
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

	if !v.Valid() {
		app.FailedValidation(w, r, v.Errors)
		return
	}

	courses, err := app.fetchCourses(r.Context(), courseIDs)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	lessons, err := app.fetchLessons(r.Context(), lessonIDs)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	// Dars xaridida ham darsning kursi kimniki ekanini bilish kerak
	// (o'z kursini sotib olishni taqiqlash uchun).
	lessonCourseIDs := []int64{}
	for _, l := range lessons {
		if !seenCourses[l.CourseID] {
			lessonCourseIDs = append(lessonCourseIDs, l.CourseID)
		}
	}
	lessonCourses, err := app.fetchCourses(r.Context(), lessonCourseIDs)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}
	for id, c := range courses {
		lessonCourses[id] = c
	}

	// Allaqachon egalik qilinayotgan narsani qayta sotib bo'lmaydi.
	ownedCourses, err := app.models.Enrollments.OwnedCourses(claims.UserID, courseIDs)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}
	ownedLessons, err := app.models.Enrollments.OwnedLessons(claims.UserID, lessonIDs)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	order := &data.Order{
		UserID:        claims.UserID,
		Status:        "paid", // mock to'lov
		PaymentMethod: input.PaymentMethod,
	}

	for i, id := range courseIDs {
		course, ok := courses[id]
		switch {
		case !ok || !course.IsPublished:
			v.AddError(fmt.Sprintf("items[%d].course_id", i), "course does not exist")
			continue
		case course.Instructor.ID == claims.UserID:
			v.AddError(fmt.Sprintf("items[%d].course_id", i), "you cannot purchase your own course")
			continue
		case ownedCourses[id]:
			v.AddError(fmt.Sprintf("items[%d].course_id", i), "you already own this course")
			continue
		}
		order.Items = append(order.Items, &data.OrderItem{
			CourseID:       &course.ID,
			CourseTitle:    course.Title,
			Instructor:     course.Instructor.Name,
			ThumbnailColor: uidefaults.ThumbnailColor(course.ID),
			Price:          course.Price,
		})
	}

	for i, id := range lessonIDs {
		lesson, ok := lessons[id]
		switch {
		case !ok:
			v.AddError(fmt.Sprintf("items[%d].lesson_id", i), "lesson does not exist")
			continue
		case lessonCourses[lesson.CourseID] != nil && lessonCourses[lesson.CourseID].Instructor.ID == claims.UserID:
			v.AddError(fmt.Sprintf("items[%d].lesson_id", i), "you cannot purchase a lesson from your own course")
			continue
		case ownedLessons[id]:
			v.AddError(fmt.Sprintf("items[%d].lesson_id", i), "you already own this lesson")
			continue
		}
		order.Items = append(order.Items, &data.OrderItem{
			LessonID:       &lesson.ID,
			CourseTitle:    lesson.CourseTitle,
			ThumbnailColor: uidefaults.ThumbnailColor(lesson.CourseID),
			Price:          lesson.Price,
		})
	}

	if !v.Valid() {
		app.FailedValidation(w, r, v.Errors)
		return
	}

	err = app.models.Orders.Insert(order)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	order.Finalize()

	// To'lov o'tdi — endi kirish beriladi. Kurs: enrollment + barcha darslar
	// (eski DB-trigger o'rnida). Alohida dars: faqat o'sha dars.
	for _, id := range courseIDs {
		_, isNew, err := app.models.Enrollments.Insert(claims.UserID, id)
		if err != nil {
			app.ServerError(w, r, err)
			return
		}
		err = app.grantCourseAccess(r.Context(), claims.UserID, id)
		if err != nil {
			app.ServerError(w, r, err)
			return
		}
		if isNew {
			app.markEnrolled(r.Context(), id)
		}
	}

	for _, id := range lessonIDs {
		lesson := lessons[id]
		err = app.models.Enrollments.GrantLessonAccess(claims.UserID, lesson.CourseID, []int64{lesson.ID})
		if err != nil {
			app.ServerError(w, r, err)
			return
		}
	}

	app.notify(claims.UserID, "system", "Purchase successful",
		fmt.Sprintf("Your order #%s has been paid. %d item(s) are now available in your library.",
			order.PublicID, len(order.Items)))

	err = jsonutil.WriteJSON(w, http.StatusCreated, jsonutil.Envelope{"order": order}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

// meOrdersHandler — GET /v1/me/orders.
func (app *application) meOrdersHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ContextGetUser(r)

	orders, err := app.models.Orders.ListByUser(claims.UserID)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"items": orders}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

// meOrderHandler — GET /v1/me/orders/{id}.
func (app *application) meOrderHandler(w http.ResponseWriter, r *http.Request) {
	id, err := jsonutil.ReadIDParam(r)
	if err != nil {
		app.NotFound(w, r)
		return
	}

	claims := middleware.ContextGetUser(r)

	order, err := app.models.Orders.GetForUser(id, claims.UserID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.NotFound(w, r)
		default:
			app.ServerError(w, r, err)
		}
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"order": order}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}
