package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) serve() error {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.config.port),
		Handler: app.routes(),
	}

	shutdownError := make(chan error)

	/*
		go func() {
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			s := <-quit

			// log we are shutting down the server

			// shutdown the database

			// complete background tasks

			// app.wg.Wait()
			shutdownError <- nil
		}()
	*/

	/*
		app.LoggerPrintInfo("starting server", map[string]string{
			"addr": srv.Addr,
			"env": app.config.env,
		})
	*/

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	/*
		app.Logger.PrintInfo("stopped server", map[string]string{
			"addr", srv.Addr,
		})
	*/

	return nil
}

func (app *application) routes() http.Handler {
	router := httprouter.New()

	//router.NotFound =  http.HandlerFunc(app.notFound)
	//router.ErrorResponse = http.HandlerFunc(app.errorResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheck)

	router.HandlerFunc(http.MethodGet, "/v1/:path", app.singleDownloadHandler)

	router.HandlerFunc(http.MethodPut, "/v1/upload/:path", app.singleUploadHandler)
	router.HandlerFunc(http.MethodPost, "/v1/upload/:path", app.multiUploadHandler)

	//router.HandlerFunc(http.MethodGet"/v1/upload/:path", app.uploadHandler)

	// TODO(Farid): Add middleware
	return router
}
