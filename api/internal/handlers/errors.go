package handlers

import (
	"fmt"
	"net/http"

	"api.hloud.fprzg.net/internal/utils"
)

func internalErrorResponse(w http.ResponseWriter, err error) {
	fmt.Printf("Encountered error: %s\n", err.Error())

	errorResponse(w, http.StatusInternalServerError, "the server encountered an error and could not process your request")
}

func errorResponse(w http.ResponseWriter, status int, msg string) {
	env := utils.Envelope{
		"error": msg,
	}

	err := utils.WriteJSON(w, status, env, nil)
	if err != nil {
		fmt.Printf("Encountered error parsing json: %v/n", err)

		w.WriteHeader(500)
	}
}
