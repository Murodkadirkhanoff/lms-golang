package transport

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/chashma/lms/internal/modules/enrollment/application"
	"github.com/chashma/lms/internal/modules/enrollment/domain"
	"github.com/chashma/lms/internal/platform/web"
	"github.com/go-chi/chi/v5"
)

// --- order wire mapping ---

type orderItemJSON struct {
	CourseTitle    string  `json:"courseTitle"`
	Instructor     string  `json:"instructor"`
	ThumbnailColor string  `json:"thumbnailColor"`
	Price          float64 `json:"price"`
}

type orderJSON struct {
	ID            string          `json:"id"`
	Date          string          `json:"date"`
	Status        string          `json:"status"`
	PaymentMethod string          `json:"paymentMethod"`
	Items         []orderItemJSON `json:"items"`
	Total         float64         `json:"total"`
}

func toOrderJSON(o domain.Order) orderJSON {
	items := make([]orderItemJSON, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, orderItemJSON{
			CourseTitle: it.CourseTitle, Instructor: it.Instructor,
			ThumbnailColor: it.ThumbnailColor, Price: it.Price,
		})
	}
	return orderJSON{
		ID:            strconv.FormatInt(o.DBID, 10),
		Date:          o.CreatedAt.UTC().Format("2006-01-02"),
		Status:        o.PublicStatus(),
		PaymentMethod: o.PaymentMethod,
		Items:         items,
		Total:         o.Total,
	}
}

func validatePage(page, pageSize int) (map[string]string, bool) {
	v := web.NewValidator()
	v.Check(page > 0, "page", "must be greater than zero")
	v.Check(pageSize > 0 && pageSize <= 100, "pageSize", "must be between 1 and 100")
	return v.Errors, v.Valid()
}

// --- handlers ---

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	identity, _ := web.IdentityFrom(r.Context())
	stats, err := h.svc.Dashboard(r.Context(), identity.UserID)
	if err != nil {
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, stats, nil)
}

func (h *Handler) myCourses(w http.ResponseWriter, r *http.Request) {
	identity, _ := web.IdentityFrom(r.Context())
	page := web.ParamInt(r.URL.Query().Get("page"), 1)
	pageSize := web.ParamInt(r.URL.Query().Get("pageSize"), 20)
	if errs, ok := validatePage(page, pageSize); !ok {
		web.FailedValidation(w, errs)
		return
	}
	items, total, err := h.svc.MyCourses(r.Context(), identity.UserID, page, pageSize)
	if err != nil {
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{
		"items": items, "page": page, "pageSize": pageSize, "total": total,
	}, nil)
}

func (h *Handler) myCertificates(w http.ResponseWriter, r *http.Request) {
	identity, _ := web.IdentityFrom(r.Context())
	items, err := h.svc.Certificates(r.Context(), identity.UserID)
	if err != nil {
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{"items": items}, nil)
}

func (h *Handler) downloadCertificate(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	identity, _ := web.IdentityFrom(r.Context())

	pdf, filename, err := h.svc.RenderCertificate(r.Context(), id, identity.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			web.NotFound(w)
			return
		}
		web.ServerError(w, r, err)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Content-Type", "application/pdf")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdf)
}

func (h *Handler) myNotifications(w http.ResponseWriter, r *http.Request) {
	identity, _ := web.IdentityFrom(r.Context())
	items, err := h.svc.Notifications(r.Context(), identity.UserID)
	if err != nil {
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{"items": items}, nil)
}

func (h *Handler) readNotification(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	identity, _ := web.IdentityFrom(r.Context())
	if err := h.svc.MarkNotificationRead(r.Context(), id, identity.UserID); err != nil {
		h.writeErr(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{"message": "notification marked as read"}, nil)
}

func (h *Handler) readAllNotifications(w http.ResponseWriter, r *http.Request) {
	identity, _ := web.IdentityFrom(r.Context())
	if err := h.svc.MarkAllNotificationsRead(r.Context(), identity.UserID); err != nil {
		web.ServerError(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{"message": "all notifications marked as read"}, nil)
}

func (h *Handler) myOrders(w http.ResponseWriter, r *http.Request) {
	identity, _ := web.IdentityFrom(r.Context())
	page := web.ParamInt(r.URL.Query().Get("page"), 1)
	pageSize := web.ParamInt(r.URL.Query().Get("pageSize"), 20)
	if errs, ok := validatePage(page, pageSize); !ok {
		web.FailedValidation(w, errs)
		return
	}
	orders, total, err := h.svc.Orders(r.Context(), identity.UserID, page, pageSize)
	if err != nil {
		web.ServerError(w, r, err)
		return
	}
	items := make([]orderJSON, 0, len(orders))
	for _, o := range orders {
		items = append(items, toOrderJSON(o))
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{
		"items": items, "page": page, "pageSize": pageSize, "total": total,
	}, nil)
}

func (h *Handler) myOrder(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	identity, _ := web.IdentityFrom(r.Context())
	order, err := h.svc.Order(r.Context(), id, identity.UserID)
	if err != nil {
		h.writeErr(w, r, err)
		return
	}
	web.WriteJSON(w, http.StatusOK, web.Envelope{"order": toOrderJSON(*order)}, nil)
}

type checkoutItemReq struct {
	CourseID *int64 `json:"course_id"`
	LessonID *int64 `json:"lesson_id"`
}

type checkoutReq struct {
	Items         []checkoutItemReq `json:"items"`
	PaymentMethod string            `json:"payment_method"`
}

func (h *Handler) checkout(w http.ResponseWriter, r *http.Request) {
	identity, _ := web.IdentityFrom(r.Context())

	var in checkoutReq
	if err := web.ReadJSON(w, r, &in); err != nil {
		web.WriteBadRequest(w, err)
		return
	}
	items := make([]application.CheckoutItem, 0, len(in.Items))
	for _, it := range in.Items {
		items = append(items, application.CheckoutItem{CourseID: it.CourseID, LessonID: it.LessonID})
	}

	order, validationErrs, err := h.svc.Checkout(r.Context(), identity.UserID, items, in.PaymentMethod)
	if err != nil {
		web.ServerError(w, r, err)
		return
	}
	if len(validationErrs) > 0 {
		web.FailedValidation(w, validationErrs)
		return
	}
	web.WriteJSON(w, http.StatusCreated, web.Envelope{"order": toOrderJSON(*order)}, nil)
}
