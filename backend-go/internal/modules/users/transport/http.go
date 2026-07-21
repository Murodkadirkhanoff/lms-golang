// Package transport is the users context's HTTP adapter. It maps requests to
// use cases and domain errors to the shared error envelope.
package transport

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	coursescontract "github.com/chashma/lms/internal/modules/courses/contract"
	enrollmentcontract "github.com/chashma/lms/internal/modules/enrollment/contract"
	"github.com/chashma/lms/internal/modules/users/application"
	"github.com/chashma/lms/internal/modules/users/domain"
	"github.com/chashma/lms/internal/platform/web"
	"github.com/go-chi/chi/v5"
)

// Handler serves the users context endpoints.
type Handler struct {
	svc     *application.Service
	catalog coursescontract.CourseCatalog
	gate    enrollmentcontract.EnrollmentGate
}

// NewHandler builds the users HTTP handler. catalog/gate are peer contracts
// used only by the admin dashboard.
func NewHandler(svc *application.Service, catalog coursescontract.CourseCatalog, gate enrollmentcontract.EnrollmentGate) *Handler {
	return &Handler{svc: svc, catalog: catalog, gate: gate}
}

type userJSON struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
}

func toUserJSON(u *domain.User) userJSON {
	return userJSON{ID: u.ID, CreatedAt: u.CreatedAt, Name: u.Name, Email: u.Email, Role: u.Role}
}

// --- auth ---

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var in registerRequest
	if err := web.ReadJSON(w, r, &in); err != nil {
		web.WriteBadRequest(w, err)
		return
	}

	v := web.NewValidator()
	domain.ValidateName(v, in.Name)
	domain.ValidateEmail(v, in.Email)
	domain.ValidatePassword(v, in.Password)
	if !v.Valid() {
		web.FailedValidation(w, v.Errors)
		return
	}

	user, token, err := h.svc.Register(r.Context(), in.Name, in.Email, in.Password)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicateEmail) {
			v.AddError("email", "a user with this email address already exists")
			web.FailedValidation(w, v.Errors)
			return
		}
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusCreated, web.Envelope{"user": toUserJSON(user), "token": token}, nil)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var in loginRequest
	if err := web.ReadJSON(w, r, &in); err != nil {
		web.WriteBadRequest(w, err)
		return
	}

	v := web.NewValidator()
	domain.ValidateEmail(v, in.Email)
	domain.ValidatePassword(v, in.Password)
	if !v.Valid() {
		web.FailedValidation(w, v.Errors)
		return
	}

	user, token, err := h.svc.Authenticate(r.Context(), in.Email, in.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			web.ErrorResponse(w, http.StatusUnauthorized, "invalid authentication credentials")
			return
		}
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{"user": toUserJSON(user), "token": token}, nil)
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

func (h *Handler) forgotPassword(w http.ResponseWriter, r *http.Request) {
	var in forgotPasswordRequest
	if err := web.ReadJSON(w, r, &in); err != nil {
		web.WriteBadRequest(w, err)
		return
	}

	v := web.NewValidator()
	domain.ValidateEmail(v, in.Email)
	if !v.Valid() {
		web.FailedValidation(w, v.Errors)
		return
	}

	if err := h.svc.ForgotPassword(r.Context(), in.Email); err != nil {
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{
		"message": "if the email address exists, a password reset link will be sent",
	}, nil)
}

type resetPasswordRequest struct {
	Password string `json:"password"`
	Token    string `json:"token"`
}

func (h *Handler) resetPassword(w http.ResponseWriter, r *http.Request) {
	var in resetPasswordRequest
	if err := web.ReadJSON(w, r, &in); err != nil {
		web.WriteBadRequest(w, err)
		return
	}

	v := web.NewValidator()
	domain.ValidatePassword(v, in.Password)
	v.Check(in.Token != "", "token", "must be provided")
	if !v.Valid() {
		web.FailedValidation(w, v.Errors)
		return
	}

	err := h.svc.ResetPassword(r.Context(), in.Password, in.Token)
	if err != nil {
		if errors.Is(err, application.ErrInvalidResetToken) {
			v.AddError("token", "invalid or expired password reset token")
			web.FailedValidation(w, v.Errors)
			return
		}
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{
		"message": "your password has been reset successfully",
	}, nil)
}

// --- profile ---

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	id, _ := web.IdentityFrom(r.Context())
	user, err := h.svc.Me(r.Context(), id.UserID)
	if err != nil {
		h.writeUserErr(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{"user": toUserJSON(user)}, nil)
}

type updateProfileRequest struct {
	Name string `json:"name"`
}

func (h *Handler) updateProfile(w http.ResponseWriter, r *http.Request) {
	var in updateProfileRequest
	if err := web.ReadJSON(w, r, &in); err != nil {
		web.WriteBadRequest(w, err)
		return
	}

	v := web.NewValidator()
	domain.ValidateName(v, in.Name)
	if !v.Valid() {
		web.FailedValidation(w, v.Errors)
		return
	}

	id, _ := web.IdentityFrom(r.Context())
	user, err := h.svc.UpdateProfile(r.Context(), id.UserID, in.Name)
	if err != nil {
		h.writeUserErr(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{"user": toUserJSON(user)}, nil)
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *Handler) changePassword(w http.ResponseWriter, r *http.Request) {
	var in changePasswordRequest
	if err := web.ReadJSON(w, r, &in); err != nil {
		web.WriteBadRequest(w, err)
		return
	}

	v := web.NewValidator()
	v.Check(in.CurrentPassword != "", "current_password", "must be provided")
	domain.ValidatePassword(v, in.NewPassword)
	if !v.Valid() {
		web.FailedValidation(w, v.Errors)
		return
	}

	id, _ := web.IdentityFrom(r.Context())
	err := h.svc.ChangePassword(r.Context(), id.UserID, in.CurrentPassword, in.NewPassword)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			web.ErrorResponse(w, http.StatusUnauthorized, "invalid authentication credentials")
			return
		}
		h.writeUserErr(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{
		"message": "your password has been updated successfully",
	}, nil)
}

func (h *Handler) writeUserErr(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		web.NotFound(w)
	case errors.Is(err, domain.ErrEditConflict):
		web.EditConflict(w)
	default:
		web.ServerError(w, r, err)
	}
}

// --- admin ---

type adminUserJSON struct {
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

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	page := web.ParamInt(r.URL.Query().Get("page"), 1)
	pageSize := web.ParamInt(r.URL.Query().Get("pageSize"), 20)

	v := web.NewValidator()
	v.Check(page > 0, "page", "must be greater than zero")
	v.Check(pageSize > 0 && pageSize <= 100, "pageSize", "must be between 1 and 100")
	if !v.Valid() {
		web.FailedValidation(w, v.Errors)
		return
	}

	users, total, err := h.svc.ListUsers(r.Context(), page, pageSize)
	if err != nil {
		web.ServerError(w, r, err)
		return
	}

	ids := make([]int64, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}

	createdCounts := map[int64]int{}
	enrolledCounts := map[int64]int{}
	if len(ids) > 0 {
		if createdCounts, err = h.catalog.CourseCountsByInstructor(r.Context(), ids); err != nil {
			web.ServerError(w, r, err)
			return
		}
		if enrolledCounts, err = h.gate.EnrollmentCountsByUser(r.Context(), ids); err != nil {
			web.ServerError(w, r, err)
			return
		}
	}

	items := make([]adminUserJSON, 0, len(users))
	for _, u := range users {
		items = append(items, adminUserJSON{
			ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role,
			AvatarColor:     domain.AvatarColor(u.ID),
			JoinedAt:        u.CreatedAt.UTC().Format("2006-01-02"),
			Status:          "active",
			CoursesCreated:  createdCounts[u.ID],
			CoursesEnrolled: enrolledCounts[u.ID],
		})
	}

	web.WriteJSON(w, http.StatusOK, web.Envelope{
		"items": items, "page": page, "pageSize": pageSize, "total": total,
	}, nil)
}

type updateRoleRequest struct {
	Role string `json:"role"`
}

func (h *Handler) updateRole(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		web.NotFound(w)
		return
	}

	var in updateRoleRequest
	if err := web.ReadJSON(w, r, &in); err != nil {
		web.WriteBadRequest(w, err)
		return
	}

	v := web.NewValidator()
	v.Check(web.Permitted(in.Role, domain.RoleStudent, domain.RoleInstructor, domain.RoleAdmin),
		"role", "must be one of student, instructor, admin")
	if !v.Valid() {
		web.FailedValidation(w, v.Errors)
		return
	}

	actor, _ := web.IdentityFrom(r.Context())
	if actor.UserID == id {
		web.ErrorResponse(w, http.StatusBadRequest, "you cannot change your own role")
		return
	}

	if err := h.svc.UpdateRole(r.Context(), id, in.Role); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			web.NotFound(w)
			return
		}
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{"message": "user role updated to " + in.Role}, nil)
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	totalUsers, err := h.svc.CountUsers(r.Context())
	if err != nil {
		web.ServerError(w, r, err)
		return
	}
	courseStats, err := h.catalog.Stats(r.Context())
	if err != nil {
		web.ServerError(w, r, err)
		return
	}
	revenue, err := h.gate.Revenue(r.Context())
	if err != nil {
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{
		"totalUsers":        totalUsers,
		"totalCourses":      courseStats.TotalCourses,
		"revenue":           revenue,
		"activeInstructors": courseStats.ActiveInstructors,
	}, nil)
}
