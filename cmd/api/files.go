package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"api.hloud.fprzg.net/internal/utils"
	"github.com/julienschmidt/httprouter"
)

var (
	ErrPathIsEmpty    = errors.New("path must not be an empty string")
	ErrPathIsNotValid = errors.New("path is not valid")
)

func (app *application) extractAndSanitizePath(r *http.Request) (*string, *string, error) {
	params := httprouter.ParamsFromContext(r.Context())

	pathParam := params.ByName("path")
	if pathParam == "" {
		return nil, nil, ErrPathIsEmpty
	}

	// sanitize
	pathSanitized := pathParam

	pathFull := filepath.Join(app.config.storageDir, pathSanitized)

	return &pathSanitized, &pathFull, nil
}

func (app *application) downloadHandler(w http.ResponseWriter, r *http.Request) {
}

func (app *application) mkdirHandler(w http.ResponseWriter, r *http.Request) {
	_, pathAbs, err := app.extractAndSanitizePath(r)
	if err != nil {
		switch {
		case errors.Is(err, ErrPathIsEmpty):
			app.errorResponse(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrPathIsNotValid):
			app.errorResponse(w, http.StatusBadRequest, err.Error())
		default:
			app.internalErrorResponse(w, err)
		}
		return
	}

	pathExist, err := utils.DirExist(*pathAbs)
	if pathExist {
		app.errorResponse(w, http.StatusBadRequest, "path already exist")
		return
	}
	if err != nil {
		app.internalErrorResponse(w, err)
	}

	err = os.MkdirAll(*pathAbs, os.ModePerm)
	if err != nil {
		app.internalErrorResponse(w, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"status": "directory created sucessfully"}, nil)
	if err != nil {
		app.internalErrorResponse(w, err)
	}
}

func (app *application) singleUploadHandler(w http.ResponseWriter, r *http.Request) {
	pathRel, pathAbs, err := app.extractAndSanitizePath(r)
	if err != nil {
		switch {
		case errors.Is(err, ErrPathIsEmpty):
			fallthrough
		case errors.Is(err, ErrPathIsEmpty):
			app.errorResponse(w, http.StatusBadRequest, err.Error())
		default:
			app.internalErrorResponse(w, err)
		}

		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		app.errorResponse(w, http.StatusBadRequest, "received malformed form")
		return
	}
	// TODO(Farid): Do we have to .Close() the file before returning despite it having an error?
	defer file.Close()

	fileExist, err := utils.FileExist(*pathAbs)
	if fileExist {
		app.errorResponse(w, http.StatusBadRequest, "file already exist")
		return
	}
	if err != nil {
		app.internalErrorResponse(w, err)
		return
	}

	outFile, err := os.Create(*pathAbs)
	if err != nil {
		switch {
		case errors.Is(err, os.ErrNotExist):
			msg := fmt.Sprintf("directory %q doesn't exist", filepath.Dir(*pathRel))
			app.errorResponse(w, http.StatusBadRequest, msg)
		default:
			app.internalErrorResponse(w, err)
		}

		return
	}
	defer outFile.Close()

	// TODO(Farid): io.Copy will give problems when EOF happens to be part of the file (as is the case in binary files).
	_, err = io.Copy(outFile, file)
	if err != nil {
		app.internalErrorResponse(w, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"response": "file saved sucessfully"}, nil)
	if err != nil {
		app.internalErrorResponse(w, err)
	}
}

func (app *application) chunkedUploadStartHandler(w http.ResponseWriter, r *http.Request) {

}

func (app *application) chunkedUploadPartHandler(w http.ResponseWriter, r *http.Request) {

}
