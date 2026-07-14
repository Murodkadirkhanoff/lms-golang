package main

import (
	"errors"
	"net/http"

	"lms.chashma.uz/auth-service/internal/data"
	"lms.chashma.uz/pkg/auth"
	"lms.chashma.uz/pkg/jsonutil"
	"lms.chashma.uz/pkg/uidefaults"
	"lms.chashma.uz/pkg/validator"
)

// registerUserHandler — POST /v1/users. Kontrakt bo'yicha darhol token
// qaytariladi (email-activation keyingi bosqichda).
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	user := &data.User{
		Name:  input.Name,
		Email: input.Email,
		Role:  auth.RoleStudent,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateUser(v, user); !v.Valid() {
		app.FailedValidation(w, r, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.FailedValidation(w, r, v.Errors)
		default:
			app.ServerError(w, r, err)
		}
		return
	}

	token, err := auth.NewToken(app.config.jwtSecret, user.ID, user.Role, app.config.jwtTTL)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusCreated, jsonutil.Envelope{"user": user, "token": token}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

// resetPasswordHandler — PUT /v1/users/password. Token forgot-password
// oqimida yaratilgan bo'lishi kerak.
func (app *application) resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Password string `json:"password"`
		Token    string `json:"token"`
	}

	err := jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	v := validator.New()
	data.ValidatePasswordPlaintext(v, input.Password)
	v.Check(input.Token != "", "token", "must be provided")
	if !v.Valid() {
		app.FailedValidation(w, r, v.Errors)
		return
	}

	userID, err := app.models.Tokens.UserIDForToken(input.Token)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrInvalidToken):
			v.AddError("token", "invalid or expired password reset token")
			app.FailedValidation(w, r, v.Errors)
		default:
			app.ServerError(w, r, err)
		}
		return
	}

	user, err := app.models.Users.Get(userID)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	err = app.models.Users.UpdatePassword(user)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	err = app.models.Tokens.DeleteAllForUser(userID)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	message := "your password has been reset successfully"
	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"message": message}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

// adminListUsersHandler — GET /v1/admin/users (faqat admin).
func (app *application) adminListUsersHandler(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	page := jsonutil.ReadInt(qs, "page", 1)
	pageSize := jsonutil.ReadInt(qs, "pageSize", 20)

	v := validator.New()
	v.Check(page > 0, "page", "must be greater than zero")
	v.Check(pageSize > 0 && pageSize <= 100, "pageSize", "must be between 1 and 100")
	if !v.Valid() {
		app.FailedValidation(w, r, v.Errors)
		return
	}

	users, total, err := app.models.Users.List(page, pageSize)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	// coursesCreated/coursesEnrolled boshqa servislardan batch olinadi.
	ids := make([]int64, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}
	createdCounts := app.fetchCounts(r.Context(), app.courseClient, "/internal/course-counts", ids)
	enrolledCounts := app.fetchCounts(r.Context(), app.enrollmentClient, "/internal/enrollment-counts", ids)

	// Frontend AdminUser shakli (services/mock/users.ts).
	type adminUser struct {
		ID              int64  `json:"id"`
		Name            string `json:"name"`
		Email           string `json:"email"`
		Role            string `json:"role"`
		AvatarColor     string `json:"avatarColor"`
		JoinedAt        string `json:"joinedAt"`
		Status          string `json:"status"`
		CoursesCreated  int    `json:"coursesCreated"`
		CoursesEnrolled int    `json:"coursesEnrolled"`
	}

	items := make([]adminUser, 0, len(users))
	for _, u := range users {
		items = append(items, adminUser{
			ID:              u.ID,
			Name:            u.Name,
			Email:           u.Email,
			Role:            u.Role,
			AvatarColor:     uidefaults.AvatarColor(u.ID),
			JoinedAt:        u.CreatedAt.Format("2006-01-02"),
			Status:          "active",
			CoursesCreated:  createdCounts[u.ID],
			CoursesEnrolled: enrolledCounts[u.ID],
		})
	}

	result := jsonutil.Paginated[adminUser]{Items: items, Page: page, PageSize: pageSize, Total: total}

	err = jsonutil.WriteJSON(w, http.StatusOK, result, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

// internalUsersHandler — GET /internal/users?ids=1,2 (servislararo).
func (app *application) internalUsersHandler(w http.ResponseWriter, r *http.Request) {
	ids := jsonutil.ReadIDList(r.URL.Query(), "ids")
	if len(ids) == 0 {
		app.BadRequest(w, r, errors.New("ids query parameter must be provided"))
		return
	}

	users, err := app.models.Users.GetByIDs(ids)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"users": users}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}
