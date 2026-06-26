package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fuad71/job-circular-api/internal/model"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Create(ctx context.Context, u *model.User) error {
	query := `
		INSERT INTO users (name, email, password_hash, role, verify_token, phone, district, education_level)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		u.Name, u.Email, u.PasswordHash, u.Role, u.VerifyToken,
		u.Phone, u.District, u.EducationLevel,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	u := &model.User{}
	query := `SELECT id, name, email, password_hash, role, is_verified,
		verify_token, reset_token, reset_token_exp, phone, district,
		education_level, last_login, created_at, updated_at
		FROM users WHERE email = $1`

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.IsVerified,
		&u.VerifyToken, &u.ResetToken, &u.ResetTokenExp, &u.Phone, &u.District,
		&u.EducationLevel, &u.LastLogin, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	u := &model.User{}
	query := `SELECT id, name, email, password_hash, role, is_verified,
		verify_token, reset_token, reset_token_exp, phone, district,
		education_level, last_login, created_at, updated_at
		FROM users WHERE id = $1`

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.IsVerified,
		&u.VerifyToken, &u.ResetToken, &u.ResetTokenExp, &u.Phone, &u.District,
		&u.EducationLevel, &u.LastLogin, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *UserRepo) GetByVerifyToken(ctx context.Context, token string) (*model.User, error) {
	u := &model.User{}
	query := `SELECT id, name, email, password_hash, role, is_verified,
		verify_token, reset_token, reset_token_exp, phone, district,
		education_level, last_login, created_at, updated_at
		FROM users WHERE verify_token = $1`

	err := r.pool.QueryRow(ctx, query, token).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.IsVerified,
		&u.VerifyToken, &u.ResetToken, &u.ResetTokenExp, &u.Phone, &u.District,
		&u.EducationLevel, &u.LastLogin, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *UserRepo) MarkVerified(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET is_verified = true, verify_token = NULL WHERE id = $1`, id)
	return err
}

func (r *UserRepo) SetResetToken(ctx context.Context, id string, token string, exp time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET reset_token = $1, reset_token_exp = $2 WHERE id = $3`,
		token, exp, id)
	return err
}

func (r *UserRepo) GetByResetToken(ctx context.Context, token string) (*model.User, error) {
	u := &model.User{}
	query := `SELECT id, name, email, password_hash, role, is_verified,
		verify_token, reset_token, reset_token_exp, phone, district,
		education_level, last_login, created_at, updated_at
		FROM users WHERE reset_token = $1 AND reset_token_exp > NOW()`

	err := r.pool.QueryRow(ctx, query, token).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.IsVerified,
		&u.VerifyToken, &u.ResetToken, &u.ResetTokenExp, &u.Phone, &u.District,
		&u.EducationLevel, &u.LastLogin, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *UserRepo) UpdatePassword(ctx context.Context, id string, hash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET password_hash = $1, reset_token = NULL, reset_token_exp = NULL WHERE id = $2`,
		hash, id)
	return err
}

func (r *UserRepo) UpdateLastLogin(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET last_login = NOW() WHERE id = $1`, id)
	return err
}

// IsEmailTaken returns true if the email is already registered
func (r *UserRepo) IsEmailTaken(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check email exists: %w", err)
	}
	return exists, nil
}
