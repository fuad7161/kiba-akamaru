package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/fuad71/job-circular-api/internal/config"
	"github.com/fuad71/job-circular-api/internal/database"
	"github.com/fuad71/job-circular-api/internal/handler"
	authmw "github.com/fuad71/job-circular-api/internal/middleware"
	"github.com/fuad71/job-circular-api/internal/repository"
	"github.com/fuad71/job-circular-api/internal/service"
	"github.com/fuad71/job-circular-api/pkg/response"
)

func main() {
	log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	if config.AppEnv() == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	pg, err := database.ConnectPostgres(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	defer pg.Close()

	// ── Auth wiring ─────────────────────────────────────────────
	userRepo := repository.NewUserRepo(pg)
	authSvc := service.NewAuthService(userRepo, pg, cfg)
	authHandler := handler.NewAuthHandler(authSvc)
	authRequired := authmw.AuthRequired(authSvc)

	// ── Router ──────────────────────────────────────────────────
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Heartbeat("/healthz"))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		dbStatus := "ok"
		if err := pg.Ping(ctx); err != nil {
			dbStatus = "unavailable"
		}

		response.JSON(w, http.StatusOK, map[string]string{
			"status": "ok",
			"db":     dbStatus,
		})
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			response.JSON(w, http.StatusOK, map[string]string{
				"message": "BD Govt Job Circular API v1.0.0",
			})
		})

		// ── Auth routes ──────────────────────────────────────
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Get("/verify-email", authHandler.VerifyEmail)
			r.Post("/forgot-password", authHandler.ForgotPassword)
			r.Post("/reset-password", authHandler.ResetPassword)

			// JWTAuth group
			r.Group(func(r chi.Router) {
				r.Use(authRequired)
				r.Post("/logout", authHandler.Logout)
				r.Get("/me", authHandler.Me)
			})

			// Refresh (reads cookie, not auth header)
			r.Post("/refresh", authHandler.Refresh)
		})
	})

	// ── Serve Frontend ──────────────────────────────────────────
	fs := http.FileServer(http.Dir("frontend"))
	r.Handle("/*", fs)

	addr := fmt.Sprintf(":%s", cfg.AppPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("addr", addr).Msg("server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server stopped")
}
