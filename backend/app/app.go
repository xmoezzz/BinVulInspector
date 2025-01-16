package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	"bin-vul-inspector/app/kit"
)

type App struct {
	*kit.Kit
	*http.Server
}

func New(kit *kit.Kit, handler http.Handler) *App {
	server := &http.Server{
		Handler:      handler,
		Addr:         kit.Config.Http.Address,
		ReadTimeout:  kit.Config.Http.ReadTimeout,
		WriteTimeout: kit.Config.Http.WriteTimeout,
		IdleTimeout:  kit.Config.Http.IdleTimeout,
	}

	return &App{
		Kit:    kit,
		Server: server,
	}
}

func (app *App) Serve() {
	app.Logger.Infof("server is running: %s", app.Config.Http.Address)
	if err := app.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		app.Logger.Errorf("server listen and serve error, %w", err)
	}
}

func (app *App) Close(ctx context.Context) {
	app.Logger.Info("shutting down the server")

	// server graceful shutdown
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := app.Shutdown(ctxWithTimeout); err != nil {
		app.Logger.Errorf("shutting down the server error, %w", err)
	} else {
		app.Logger.Info("server has been shutdown")
	}
}
