package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"api.hloud.fprzg.net/internal/info"
	"api.hloud.fprzg.net/internal/utils"
	"github.com/julienschmidt/httprouter"
)

var (
	ErrPathIsEmpty    = errors.New("path must not be an empty string")
	ErrPathIsNotValid = errors.New("path is not valid")
)

type FileHandlers struct {
	cfg   *info.Config
	build *info.Build
}
type FileMetadata struct {
	//CreationDate
	Name string `json:"name"`
	Size int64  `json:"size"`
}

func (h FileHandlers) Routes(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, "/v1/get/*path", h.DownloadHandler)

	router.HandlerFunc(http.MethodPost, "/v1/mkdir/*path", h.MkdirHandler)
	router.HandlerFunc(http.MethodPost, "/v1/upload/*path", h.SingleUploadHandler)
	router.HandlerFunc(http.MethodPost, "/v1/chunked-upload-start/*path", h.ChunkedUploadStartHandler)
	router.HandlerFunc(http.MethodPost, "/v1/chunked-upload-part/*path", h.ChunkedUploadPartHandler)

	router.HandlerFunc(http.MethodGet, "/v1/fileMetadata/*path", h.FileMetadataHandler)
	router.HandlerFunc(http.MethodGet, "/v1/dirMetadata/*path", h.DirectoryMetadataHandler)

	router.HandlerFunc(http.MethodDelete, "/v1/rmdir/*path", h.rmdirHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/delete/*path", h.deleteFilehandler)

	router.HandlerFunc(http.MethodPost, "/v1/rename/*path", h.renameHandler)
	router.HandlerFunc(http.MethodPost, "/v1/move/*path", h.moveHandler)
}

func (h *FileHandlers) DownloadHandler(w http.ResponseWriter, r *http.Request) {
}

func (h *FileHandlers) MkdirHandler(w http.ResponseWriter, r *http.Request) {
	_, pathAbs, err := h.extractAndSanitizePath(r)
	if err != nil {
		switch {
		case errors.Is(err, ErrPathIsEmpty):
			errorResponse(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrPathIsNotValid):
			errorResponse(w, http.StatusBadRequest, err.Error())
		default:
			internalErrorResponse(w, err)
		}
		return
	}

	pathExist, err := utils.DirExist(*pathAbs)
	if pathExist {
		errorResponse(w, http.StatusBadRequest, "path already exist")
		return
	}
	if err != nil {
		internalErrorResponse(w, err)
	}

	err = os.MkdirAll(*pathAbs, os.ModePerm)
	if err != nil {
		internalErrorResponse(w, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Envelope{"status": "directory created sucessfully"}, nil)
	if err != nil {
		internalErrorResponse(w, err)
	}
}

func (h *FileHandlers) SingleUploadHandler(w http.ResponseWriter, r *http.Request) {
	pathRel, pathAbs, err := h.handlePathExtraction(w, r)
	if err != nil {
		internalErrorResponse(w, err)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "received malformed form")
		return
	}
	// TODO(Farid): Do we have to .Close() the file before returning despite it having an error?
	defer file.Close()

	fileExist, err := utils.FileExist(*pathAbs)
	if fileExist {
		errorResponse(w, http.StatusBadRequest, "file already exist")
		return
	}
	if err != nil {
		internalErrorResponse(w, err)
		return
	}

	outFile, err := os.Create(*pathAbs)
	if err != nil {
		switch {
		case errors.Is(err, os.ErrNotExist):
			msg := fmt.Sprintf("directory %q doesn't exist", filepath.Dir(*pathRel))
			errorResponse(w, http.StatusBadRequest, msg)
		default:
			internalErrorResponse(w, err)
		}

		return
	}
	defer outFile.Close()

	// TODO(Farid): io.Copy will give problems when EOF happens to be part of the file (as is the case in binary files).
	_, err = io.Copy(outFile, file)
	if err != nil {
		internalErrorResponse(w, err)
		return
	}

	err = utils.WriteJSON(w, http.StatusOK, utils.Envelope{"response": "file saved sucessfully"}, nil)
	if err != nil {
		internalErrorResponse(w, err)
	}
}

func (h *FileHandlers) ChunkedUploadStartHandler(w http.ResponseWriter, r *http.Request) {}

func (h *FileHandlers) ChunkedUploadPartHandler(w http.ResponseWriter, r *http.Request) {}

func (h *FileHandlers) FileMetadataHandler(w http.ResponseWriter, r *http.Request) {
	_, pathAbs, err := h.handlePathExtraction(w, r)
	if err != nil {
		internalErrorResponse(w, err)
	}

	fileMetadata, err := getFileMetadata(*pathAbs)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			errorResponse(w, http.StatusBadRequest, "directory does not exist")
		default:
			internalErrorResponse(w, err)
		}

		return
	}

	utils.WriteJSON(w, http.StatusOK, fileMetadata, nil)
}

func (h *FileHandlers) DirectoryMetadataHandler(w http.ResponseWriter, r *http.Request) {
	_, dirPathAbs, err := h.handlePathExtraction(w, r)
	if err != nil {
		internalErrorResponse(w, err)
	}

	filesInDir, err := os.ReadDir(*dirPathAbs)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			errorResponse(w, http.StatusBadRequest, "directory does not exist")
		default:
			internalErrorResponse(w, err)
		}

		return
	}
	var filesMetadata []FileMetadata

	for _, individualFilePath := range filesInDir {
		individualFilePathAbs := filepath.Join(*dirPathAbs, individualFilePath.Name())
		individualFileMetadata, _ := getFileMetadata(individualFilePathAbs)
		filesMetadata = append(filesMetadata, individualFileMetadata)
	}

	utils.WriteJSON(w, http.StatusOK, filesMetadata, nil)
}

func (h *FileHandlers) rmdirHandler(w http.ResponseWriter, r *http.Request) {
	_, pathAbs, err := h.handlePathExtraction(w, r)
	if err != nil {
		internalErrorResponse(w, err)
		return
	}

	pathExist, _ := utils.DirExist(*pathAbs)
	if !pathExist {
		errorResponse(w, http.StatusBadRequest, "directory does not exist")
		return
	}

	err = os.RemoveAll(*pathAbs)
	if err != nil {
		internalErrorResponse(w, err)
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"status": "directory deleted successfully"}, nil)
}

func (h *FileHandlers) deleteFilehandler(w http.ResponseWriter, r *http.Request) {
	_, pathAbs, err := h.handlePathExtraction(w, r)
	if err != nil {
		internalErrorResponse(w, err)
		return
	}

	fileExist, _ := utils.FileExist(*pathAbs)
	if !fileExist {
		errorResponse(w, http.StatusBadRequest, "file does not exist")
		return
	}

	err = os.Remove(*pathAbs)
	if err != nil {
		internalErrorResponse(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"status": "file deleted successfully"}, nil)
}

func (h *FileHandlers) renameHandler(w http.ResponseWriter, r *http.Request) {
	_, pathAbs, err := h.handlePathExtraction(w, r)
	if err != nil {
		internalErrorResponse(w, err)
		return
	}

	pathExist, err := utils.DirExist(*pathAbs)
	if !pathExist {
		errorResponse(w, http.StatusBadRequest, "directory does not exist")
		return
	}
	if err != nil {
		switch {
		case os.IsNotExist(err):
			errorResponse(w, http.StatusBadRequest, "directory does not exist")
		default:
			internalErrorResponse(w, err)
		}
		return
	}

	var jsonValue struct {
		NewName string `json:"new_name"`
	}

	err = utils.ReadJSON(w, r, &jsonValue)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	newPathAbs := filepath.Join(filepath.Dir(*pathAbs), jsonValue.NewName)

	err = os.Rename(*pathAbs, newPathAbs)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"status": "directory renamed successfully"}, nil)
}

func (h *FileHandlers) moveHandler(w http.ResponseWriter, r *http.Request) {
	_, oldPathAbs, err := h.handlePathExtraction(w, r)
	if err != nil {
		internalErrorResponse(w, err)
		return
	}

	pathExist, err := utils.DirExist(*oldPathAbs)
	if !pathExist {
		errorResponse(w, http.StatusBadRequest, "directory does not exist")
		return
	}
	if err != nil {
		internalErrorResponse(w, err)
		return
	}

	var jsonValue struct {
		NewPath string `json:"new_path"`
	}

	err = utils.ReadJSON(w, r, &jsonValue)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	newPathAbs := filepath.Join(h.cfg.StorageDir, jsonValue.NewPath)
	pathExist, err = utils.DirExist(newPathAbs)
	if pathExist {
		errorResponse(w, http.StatusBadRequest, "there is already a file with the same name in the target directory")
		return
	}
	if err != nil {
		internalErrorResponse(w, err)
		return
	}

	pathExist, err = utils.DirExist(filepath.Dir(newPathAbs))
	if !pathExist {
		errorResponse(w, http.StatusBadRequest, "target directory does not exist")
		return
	}
	if err != nil {
		internalErrorResponse(w, err)
		return
	}

	err = os.Rename(*oldPathAbs, newPathAbs)
	if err != nil {
		internalErrorResponse(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"status": "directory renamed sucessfully"}, nil)
}

func (h *FileHandlers) extractAndSanitizePath(r *http.Request) (*string, *string, error) {
	params := httprouter.ParamsFromContext(r.Context())

	pathParam := params.ByName("path")
	if pathParam == "" {
		return nil, nil, ErrPathIsEmpty
	}

	// sanitize
	pathSanitized := pathParam

	pathFull := filepath.Join(h.cfg.StorageDir, pathSanitized)

	return &pathSanitized, &pathFull, nil
}

func (h *FileHandlers) handlePathExtraction(w http.ResponseWriter, r *http.Request) (*string, *string, error) {
	pathRel, pathAbs, err := h.extractAndSanitizePath(r)
	if err != nil {
		switch {
		case errors.Is(err, ErrPathIsEmpty):
			fallthrough
		case errors.Is(err, ErrPathIsEmpty):
			errorResponse(w, http.StatusBadRequest, err.Error())
		}

		return nil, nil, err
	}

	return pathRel, pathAbs, nil
}

func getFileMetadata(path string) (FileMetadata, error) {
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
