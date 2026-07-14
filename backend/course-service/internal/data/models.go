package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Categories CategoryModel
	Courses    CourseModel
	Reviews    ReviewModel
	Quizzes    QuizModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Categories: CategoryModel{DB: db},
		Courses:    CourseModel{DB: db},
		Reviews:    ReviewModel{DB: db},
		Quizzes:    QuizModel{DB: db},
	}
}
