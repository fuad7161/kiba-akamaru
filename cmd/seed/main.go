package main

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/fuad71/job-circular-api/internal/config"
	"github.com/fuad71/job-circular-api/internal/database"
	"github.com/fuad71/job-circular-api/internal/service"
)

func main() {
	log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	if cfg.SeedAdminEmail == "" || cfg.SeedAdminPassword == "" {
		log.Fatal().Msg("SEED_ADMIN_EMAIL and SEED_ADMIN_PASSWORD must be set in .env")
	}

	pg, err := database.ConnectPostgres(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	defer pg.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if admin already exists
	var exists bool
	err = pg.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, cfg.SeedAdminEmail).Scan(&exists)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to check existing user")
	}
	if exists {
		log.Info().Str("email", cfg.SeedAdminEmail).Msg("admin user already exists, nothing to do")
		return
	}

	hash, err := service.HashPassword(cfg.SeedAdminPassword)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to hash password")
	}

	var id string
	err = pg.QueryRow(ctx,
		`INSERT INTO users (name, email, password_hash, role, is_verified) VALUES ($1, $2, $3, $4, true) RETURNING id`,
		cfg.SeedAdminName, cfg.SeedAdminEmail, hash, "admin",
	).Scan(&id)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create admin user")
	}

	log.Info().Str("email", cfg.SeedAdminEmail).Str("name", cfg.SeedAdminName).Str("id", id).Msg("admin user created successfully")
}
