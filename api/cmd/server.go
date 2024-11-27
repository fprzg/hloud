package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

/*
func (app *application) serve() error {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.config.port),
		Handler: app.routes(),
	}

	shutdownError := make(chan error)

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

	app.LoggerPrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.config.env,
	})

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.Logger.PrintInfo("stopped server", map[string]string{
		"addr", srv.Addr,
	})

	return nil
}
*/

func (app *application) serve() error {
	srv := &http.Server{
		Addr:    app.cfg.GetPort(),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	return err
}

func (app *application) routes() http.Handler {
	router := httprouter.New()

	//router.NotFound =  http.HandlerFunc(app.notFound)
	//router.ErrorResponse = http.HandlerFunc(app.errorResponse)

	app.handlers.HealthCheck.Routes(router)
	app.handlers.Files.Routes(router)

	// TODO(Farid): Add middleware
	return router
}
