package main

import (
	"net/http"

	"lms.chashma.uz/course-service/internal/data"
	"lms.chashma.uz/pkg/jsonutil"
	"lms.chashma.uz/pkg/uidefaults"
)

// Instruktorlar alohida jadval emas — kamida bitta published kursi bor
// userlar. Statistika courses jadvalidan, ismlar auth-service'dan.

func (app *application) listInstructorsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := app.models.Courses.InstructorStats()
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	ids := make([]int64, 0, len(stats))
	for _, s := range stats {
		ids = append(ids, s.InstructorID)
	}

	users := app.fetchUsers(r.Context(), ids)

	items := make([]*data.Instructor, 0, len(stats))
	for _, s := range stats {
		name := "Instructor"
		if u, ok := users[s.InstructorID]; ok {
			name = u.Name
		}
		items = append(items, &data.Instructor{
			ID:          s.InstructorID,
			Name:        name,
			Headline:    "Instructor",
			AvatarColor: uidefaults.AvatarColor(s.InstructorID),
			Students:    s.Students,
			Courses:     s.CourseCount,
			Rating:      s.Rating,
		})
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"items": items}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

func (app *application) showInstructorHandler(w http.ResponseWriter, r *http.Request) {
	id, err := jsonutil.ReadIDParam(r)
	if err != nil {
		app.NotFound(w, r)
		return
	}

	stats, err := app.models.Courses.InstructorStats()
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	var stat data.InstructorStat
	for _, s := range stats {
		if s.InstructorID == id {
			stat = *s
			break
		}
	}

	users := app.fetchUsers(r.Context(), []int64{id})
	u, ok := users[id]
	if !ok {
		app.NotFound(w, r)
		return
	}

	instructor := &data.Instructor{
		ID:          id,
		Name:        u.Name,
		Headline:    "Instructor",
		AvatarColor: uidefaults.AvatarColor(id),
		Students:    stat.Students,
		Courses:     stat.CourseCount,
		Rating:      stat.Rating,
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"instructor": instructor}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}
