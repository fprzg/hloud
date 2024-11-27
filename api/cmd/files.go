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

type FileMetadata struct {
	//CreationDate
	Name string `json:"name"`
	Size int64  `json:"size"`
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
	pathRel, pathAbs, err := app.handlePathExtraction(w, r)
	if err != nil {
		app.internalErrorResponse(w, err)
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

func (app *application) chunkedUploadStartHandler(w http.ResponseWriter, r *http.Request) {}

func (app *application) chunkedUploadPartHandler(w http.ResponseWriter, r *http.Request) {}

func (app *application) fileMetadataHandler(w http.ResponseWriter, r *http.Request) {
	_, pathAbs, err := app.handlePathExtraction(w, r)
	if err != nil {
		app.internalErrorResponse(w, err)
	}

	fileMetadata, err := app.getFileMetadata(*pathAbs)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			app.errorResponse(w, http.StatusBadRequest, "directory does not exist")
		default:
			app.internalErrorResponse(w, err)
		}

		return
	}

	app.writeJSON(w, http.StatusOK, fileMetadata, nil)
}

func (app *application) directoryMetadataHandler(w http.ResponseWriter, r *http.Request) {
	_, dirPathAbs, err := app.handlePathExtraction(w, r)
	if err != nil {
		app.internalErrorResponse(w, err)
	}

	filesInDir, err := os.ReadDir(*dirPathAbs)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			app.errorResponse(w, http.StatusBadRequest, "directory does not exist")
		default:
			app.internalErrorResponse(w, err)
		}

		return
	}
	var filesMetadata []FileMetadata

	for _, individualFilePath := range filesInDir {
		individualFilePathAbs := filepath.Join(*dirPathAbs, individualFilePath.Name())
		individualFileMetadata, _ := app.getFileMetadata(individualFilePathAbs)
		filesMetadata = append(filesMetadata, individualFileMetadata)
	}

	app.writeJSON(w, http.StatusOK, filesMetadata, nil)
}

func (app *application) rmdirHandler(w http.ResponseWriter, r *http.Request) {
	_, pathAbs, err := app.handlePathExtraction(w, r)
	if err != nil {
		app.internalErrorResponse(w, err)
		return
	}

	pathExist, _ := utils.DirExist(*pathAbs)
	if !pathExist {
		app.errorResponse(w, http.StatusBadRequest, "directory does not exist")
		return
	}

	err = os.RemoveAll(*pathAbs)
	if err != nil {
		app.internalErrorResponse(w, err)
	}

	app.writeJSON(w, http.StatusOK, envelope{"status": "directory deleted successfully"}, nil)
}

func (app *application) deleteFilehandler(w http.ResponseWriter, r *http.Request) {
	_, pathAbs, err := app.handlePathExtraction(w, r)
	if err != nil {
		app.internalErrorResponse(w, err)
		return
	}

	fileExist, _ := utils.FileExist(*pathAbs)
	if !fileExist {
		app.errorResponse(w, http.StatusBadRequest, "file does not exist")
		return
	}

	err = os.Remove(*pathAbs)
	if err != nil {
		app.internalErrorResponse(w, err)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"status": "file deleted successfully"}, nil)
}

func (app *application) renameHandler(w http.ResponseWriter, r *http.Request) {
	_, pathAbs, err := app.handlePathExtraction(w, r)
	if err != nil {
		app.internalErrorResponse(w, err)
		return
	}

	pathExist, err := utils.DirExist(*pathAbs)
	if !pathExist {
		app.errorResponse(w, http.StatusBadRequest, "directory does not exist")
		return
	}
	if err != nil {
		switch {
		case os.IsNotExist(err):
			app.errorResponse(w, http.StatusBadRequest, "directory does not exist")
		default:
			app.internalErrorResponse(w, err)
		}
		return
	}

	var jsonValue struct {
		NewName string `json:"new_name"`
	}

	err = app.readJSON(w, r, &jsonValue)
	if err != nil {
		app.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	newPathAbs := filepath.Join(filepath.Dir(*pathAbs), jsonValue.NewName)

	err = os.Rename(*pathAbs, newPathAbs)
	if err != nil {
		app.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"status": "directory renamed successfully"}, nil)
}

func (app *application) moveHandler(w http.ResponseWriter, r *http.Request) {
	_, oldPathAbs, err := app.handlePathExtraction(w, r)
	if err != nil {
		app.internalErrorResponse(w, err)
		return
	}

	pathExist, err := utils.DirExist(*oldPathAbs)
	if !pathExist {
		app.errorResponse(w, http.StatusBadRequest, "directory does not exist")
		return
	}
	if err != nil {
		app.internalErrorResponse(w, err)
		return
	}

	var jsonValue struct {
		NewPath string `json:"new_path"`
	}

	err = app.readJSON(w, r, &jsonValue)
	if err != nil {
		app.errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	newPathAbs := filepath.Join(app.config.storageDir, jsonValue.NewPath)
	pathExist, err = utils.DirExist(newPathAbs)
	if pathExist {
		app.errorResponse(w, http.StatusBadRequest, "there is already a file with the same name in the target directory")
		return
	}
	if err != nil {
		app.internalErrorResponse(w, err)
		return
	}

	pathExist, err = utils.DirExist(filepath.Dir(newPathAbs))
	if !pathExist {
		app.errorResponse(w, http.StatusBadRequest, "target directory does not exist")
		return
	}
	if err != nil {
		app.internalErrorResponse(w, err)
		return
	}

	err = os.Rename(*oldPathAbs, newPathAbs)
	if err != nil {
		app.internalErrorResponse(w, err)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"status": "directory renamed sucessfully"}, nil)
}

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

func (app *application) handlePathExtraction(w http.ResponseWriter, r *http.Request) (*string, *string, error) {
	pathRel, pathAbs, err := app.extractAndSanitizePath(r)
	if err != nil {
		switch {
		case errors.Is(err, ErrPathIsEmpty):
			fallthrough
		case errors.Is(err, ErrPathIsEmpty):
			app.errorResponse(w, http.StatusBadRequest, err.Error())
		}

		return nil, nil, err
	}

	return pathRel, pathAbs, nil
}

func (app *application) getFileMetadata(path string) (FileMetadata, error) {
	file, err := os.Stat(path)
	if err != nil {
		return FileMetadata{}, err
	}

	fileMetadata := FileMetadata{
		Name: file.Name(),
		Size: file.Size(),
	}

	return fileMetadata, nil
}
