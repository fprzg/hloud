package main

import "net/http"

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	// log the error

	msg := "the server encountered an error and could not process your request"
	app.errorResponse(w, r, http.StatusNotFound, msg)
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, msg string) {
	env := envelope{
		"error": msg,
	}

	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		// log the error
		w.WriteHeader(500)
	}
}
