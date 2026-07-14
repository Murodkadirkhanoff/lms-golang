package main

import (
	"net/http"

	"lms.chashma.uz/pkg/jsonutil"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	env := jsonutil.Envelope{
		"status": "available",
		"system_info": map[string]string{
			"service":     "auth",
			"environment": app.config.envName,
			"version":     version,
		},
	}

	err := jsonutil.WriteJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.ServerError(w, r, err)
	}
}
