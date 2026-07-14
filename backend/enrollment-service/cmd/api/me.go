package main

import (
	"net/http"

	"lms.chashma.uz/pkg/jsonutil"
	"lms.chashma.uz/pkg/middleware"
)

// meCertificatesHandler — GET /v1/me/certificates.
func (app *application) meCertificatesHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ContextGetUser(r)

	certificates, err := app.models.Certificates.ListByUser(claims.UserID)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"items": certificates}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

// meNotificationsHandler — GET /v1/me/notifications.
func (app *application) meNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ContextGetUser(r)

	notifications, err := app.models.Notifications.ListByUser(claims.UserID)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"items": notifications}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

// readAllNotificationsHandler — POST /v1/me/notifications/read-all.
func (app *application) readAllNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ContextGetUser(r)

	err := app.models.Notifications.MarkAllRead(claims.UserID)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"message": "all notifications marked as read"}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

// notify xabarnoma yozadi; xato bo'lsa so'rovni yiqitmaydi, faqat log.
func (app *application) notify(userID int64, notifType, title, body string) {
	err := app.models.Notifications.Insert(userID, notifType, title, body)
	if err != nil {
		app.logger.Warn("failed to insert notification", "error", err.Error())
	}
}
