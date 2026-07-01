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
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
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

	// ── Seed admin user (if configured) ────────────────────────
	seedAdminUser(pg)

	// ── Repositories ───────────────────────────────────────────
	userRepo := repository.NewUserRepo(pg)
	circularRepo := repository.NewCircularRepo(pg)
	bookmarkRepo := repository.NewBookmarkRepo(pg)
	alertRepo := repository.NewAlertRepo(pg)

	// ── Services ───────────────────────────────────────────────
	authSvc := service.NewAuthService(userRepo, pg, cfg)

	// ── Handlers ───────────────────────────────────────────────
	authHandler := handler.NewAuthHandler(authSvc)
	circularHandler := handler.NewCircularHandler(circularRepo)
	userHandler := handler.NewUserHandler(authSvc, userRepo, bookmarkRepo, alertRepo)
	adminHandler := handler.NewAdminHandler(circularRepo)

	// ── Middleware ─────────────────────────────────────────────
	authRequired := authmw.AuthRequired(authSvc)
	adminRequired := authmw.AdminOnly

	// ── Router ─────────────────────────────────────────────────
	r := chi.NewRouter()

	// CORS
	r.Use(corsMiddleware(cfg.FrontendURL))
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Heartbeat("/healthz"))

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

			r.Group(func(r chi.Router) {
				r.Use(authRequired)
				r.Post("/logout", authHandler.Logout)
				r.Get("/me", authHandler.Me)
			})
			r.Post("/refresh", authHandler.Refresh)
		})

		// ── Circulars (public) ───────────────────────────────
		r.Get("/circulars", circularHandler.List)
		r.Get("/circulars/featured", circularHandler.Featured)
		r.Get("/circulars/{id}", circularHandler.Get)

		// ── Circulars (admin) ────────────────────────────────
		r.Group(func(r chi.Router) {
			r.Use(authRequired)
			r.Use(adminRequired)
			r.Post("/circulars", circularHandler.Create)
			r.Put("/circulars/{id}", circularHandler.Update)
			r.Delete("/circulars/{id}", circularHandler.Delete)
			r.Patch("/circulars/{id}/feature", circularHandler.ToggleFeature)
		})

		// ── Categories & Organizations (public) ──────────────
		r.Get("/categories", circularHandler.ListCategories)
		r.Get("/organizations", circularHandler.ListOrganizations)

		// ── User profile + bookmarks + alerts (JWT) ─────────
		r.Group(func(r chi.Router) {
			r.Use(authRequired)
			r.Get("/users/me", userHandler.GetProfile)
			r.Put("/users/me", userHandler.UpdateProfile)
			r.Get("/users/me/bookmarks", userHandler.ListBookmarks)
			r.Post("/users/me/bookmarks/{id}", userHandler.AddBookmark)
			r.Delete("/users/me/bookmarks/{id}", userHandler.RemoveBookmark)
			r.Get("/users/me/alerts", userHandler.ListAlerts)
			r.Post("/users/me/alerts", userHandler.CreateAlert)
			r.Delete("/users/me/alerts/{id}", userHandler.DeleteAlert)
			r.Patch("/users/me/alerts/{id}/toggle", userHandler.ToggleAlert)
		})

		// ── Admin (JWT + admin role) ────────────────────────
		r.Group(func(r chi.Router) {
			r.Use(authRequired)
			r.Use(adminRequired)
			r.Get("/admin/stats", adminHandler.Stats)
			r.Get("/admin/users", adminHandler.ListUsers)
			r.Post("/admin/scrape/run", adminHandler.TriggerScrape)
			r.Get("/admin/scrape/logs", adminHandler.ScrapeLogs)
		})
	})

	// ── Serve Frontend ──────────────────────────────────────────
	r.Handle("/css/*", http.StripPrefix("/css", http.FileServer(http.Dir("frontend/css"))))
	r.Handle("/js/*", http.StripPrefix("/js", http.FileServer(http.Dir("frontend/js"))))
	r.Handle("/pages/*", http.StripPrefix("/pages", http.FileServer(http.Dir("frontend/pages"))))
	r.Handle("/components/*", http.StripPrefix("/components", http.FileServer(http.Dir("frontend/components"))))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/index.html")
	})

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

// seedAdminUser creates an admin user if it doesn't exist.
// Configure via SEED_ADMIN_EMAIL and SEED_ADMIN_PASSWORD env vars.
func seedAdminUser(pg *pgxpool.Pool) {
	email := os.Getenv("SEED_ADMIN_EMAIL")
	password := os.Getenv("SEED_ADMIN_PASSWORD")
	if email == "" || password == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var exists bool
	err := pg.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)
	if err != nil {
		log.Warn().Err(err).Msg("seed admin: failed to check existing user")
		return
	}
	if exists {
		log.Info().Str("email", email).Msg("seed admin: user already exists, skipping")
		return
	}

	hash, err := service.HashPassword(password)
	if err != nil {
		log.Warn().Err(err).Msg("seed admin: failed to hash password")
		return
	}

	var id string
	err = pg.QueryRow(ctx,
		`INSERT INTO users (name, email, password_hash, role, is_verified) VALUES ($1, $2, $3, $4, true) RETURNING id`,
		"Admin", email, hash, "admin",
	).Scan(&id)
	if err != nil {
		log.Warn().Err(err).Msg("seed admin: failed to create admin user")
		return
	}

	log.Info().Str("email", email).Str("id", id).Msg("seed admin: admin user created")
}

// corsMiddleware adds CORS headers for the frontend
func corsMiddleware(frontendURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", frontendURL)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
