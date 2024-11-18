package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/julienschmidt/httprouter"
)

func (app *application) singleDownloadHandler(w http.ResponseWriter, r *http.Request) {
}

func (app *application) singleUploadHandler(w http.ResponseWriter, r *http.Request) {
}

func (app *application) multiUploadHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	path := params.ByName("path")
	if path == "" {
		// path must be a non empty string
		return
	}

	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		// couldn't create the "path" directory
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		// couln't read the file
		return
	}
	defer file.Close()

	filepath := filepath.Join(path, header.Filename)

	outFile, err := os.Create(filepath)
	if err != nil {
		// couldn't create the file
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		// error saving the file
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Println(w, "File '%s' saved in '%s'", header.Filename, filepath)
}
