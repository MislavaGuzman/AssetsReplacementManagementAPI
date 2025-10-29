package main

import (
	"context"
	"database/sql"
	"errors"
	"expvar"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"

	"github.com/MislavaGuzman/AssetsReplacementManagementAPI/internal/ratelimiter"
	"github.com/MislavaGuzman/AssetsReplacementManagementAPI/internal/services"
	"github.com/MislavaGuzman/AssetsReplacementManagementAPI/internal/store"
)

type application struct {
	config        appConfig
	store         store.Storage
	logger        *zap.SugaredLogger
	db            *sql.DB
	rateLimiter   ratelimiter.Limiter
	ticketService *services.TicketService
}

// ROUTER
func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)
		r.Handle("/debug/vars", expvar.Handler())
	})
	r.Route("/v1/asset-replacement-tickets", func(r chi.Router) {
		r.Get("/", app.getAllAssetReplacementTicketsHandler)
		r.Post("/", app.createAssetReplacementTicketHandler)
		r.Get("/by-ticket", app.getAssetReplacementTicketHandler)
		r.Patch("/by-ticket", app.updateAssetReplacementTicketHandler)
		r.Delete("/", app.deleteAssetReplacementTicketHandler)
		r.Get("/basic", app.getBasicTicketsHandler)
		r.Post("/upsert-batch", app.upsertBatchHandler)
		r.Post("/upsert-csv", app.upsertBatchCSVHandler)
	})

	return r
}

func (app *application) run(mux http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  1 * time.Minute,
	}

	shutdown := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit
		app.logger.Infow("signal caught", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		shutdown <- srv.Shutdown(ctx)
	}()

	app.logger.Infow("server has started", "addr", app.config.addr, "env", app.config.env)
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdown
	if err != nil {
		return err
	}

	app.logger.Infow("Server stopped", "addr", app.config.addr, "env", app.config.env)
	return nil
}
