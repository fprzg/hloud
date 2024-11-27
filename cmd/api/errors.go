package main

import (
	"fmt"
	"net/http"
)

func (app *application) internalErrorResponse(w http.ResponseWriter, err error) {
	fmt.Printf("Encountered error: %s\n", err.Error())

	app.errorResponse(w, http.StatusInternalServerError, "the server encountered an error and could not process your request")
}

func (app *application) errorResponse(w http.ResponseWriter, status int, msg string) {
	env := envelope{
		"error": msg,
	}

	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		fmt.Printf("Encountered error parsing json: %v/n", err)

		w.WriteHeader(500)
	}
}
