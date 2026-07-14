package main

import (
	"errors"
	"net/http"
	"time"

	"lms.chashma.uz/auth-service/internal/data"
	"lms.chashma.uz/pkg/auth"
	"lms.chashma.uz/pkg/jsonutil"
	"lms.chashma.uz/pkg/validator"
)

// createAuthenticationTokenHandler — POST /v1/tokens/authentication (login).
// Frontend javobni butunligicha AuthResult deb oladi: {user, token}.
func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	v := validator.New()

	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.Valid() {
		app.FailedValidation(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.InvalidCredentials(w, r)
		default:
			app.ServerError(w, r, err)
		}
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	if !match {
		app.InvalidCredentials(w, r)
		return
	}

	token, err := auth.NewToken(app.config.jwtSecret, user.ID, user.Role, app.config.jwtTTL)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"user": user, "token": token}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}

// forgotPasswordHandler — POST /v1/tokens/password-reset. Token yaratiladi
// va saqlanadi; SMTP hali ulanmagani uchun token dev-logga yoziladi (email
// integratsiyasi qo'shilganda shu joydan yuboriladi). Email mavjudligini
// javobda oshkor qilmaymiz.
func (app *application) forgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}

	err := jsonutil.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	v := validator.New()
	if data.ValidateEmail(v, input.Email); !v.Valid() {
		app.FailedValidation(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil && !errors.Is(err, data.ErrRecordNotFound) {
		app.ServerError(w, r, err)
		return
	}

	if user != nil {
		token, err := app.models.Tokens.New(user.ID, 45*time.Minute)
		if err != nil {
			app.ServerError(w, r, err)
			return
		}

		// TODO: SMTP ulangach shu yerdan email yuboriladi.
		app.logger.Info("password reset token created (email hali ulanmagan)",
			"email", user.Email, "token", token.Plaintext)
	}

	message := "if the email address exists, a password reset link will be sent"
	err = jsonutil.WriteJSON(w, http.StatusOK, jsonutil.Envelope{"message": message}, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}
