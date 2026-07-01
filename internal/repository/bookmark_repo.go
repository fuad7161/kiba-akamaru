package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fuad71/job-circular-api/internal/model"
)

type BookmarkRepo struct {
	pool *pgxpool.Pool
}

func NewBookmarkRepo(pool *pgxpool.Pool) *BookmarkRepo {
	return &BookmarkRepo{pool: pool}
}

func (r *BookmarkRepo) List(ctx context.Context, userID string) ([]model.Bookmark, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, circular_id, note, created_at FROM bookmarks WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list bookmarks: %w", err)
	}
	defer rows.Close()

	var bookmarks []model.Bookmark
	for rows.Next() {
		var b model.Bookmark
		if err := rows.Scan(&b.ID, &b.UserID, &b.CircularID, &b.Note, &b.CreatedAt); err != nil {
			return nil, err
		}
		bookmarks = append(bookmarks, b)
	}
	return bookmarks, nil
}

func (r *BookmarkRepo) Add(ctx context.Context, userID, circularID string) (*model.Bookmark, error) {
	b := &model.Bookmark{UserID: userID, CircularID: circularID}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO bookmarks (user_id, circular_id) VALUES ($1, $2)
		ON CONFLICT (user_id, circular_id) DO UPDATE SET created_at = NOW()
		RETURNING id, created_at`,
		userID, circularID,
	).Scan(&b.ID, &b.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("add bookmark: %w", err)
	}
	return b, nil
}

func (r *BookmarkRepo) Remove(ctx context.Context, userID, circularID string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM bookmarks WHERE user_id = $1 AND circular_id = $2`, userID, circularID)
	if err != nil {
		return fmt.Errorf("remove bookmark: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("bookmark not found")
	}
	return nil
}

// ── Alerts ──────────────────────────────────────────────────────────────────────

type AlertRepo struct {
	pool *pgxpool.Pool
}

func NewAlertRepo(pool *pgxpool.Pool) *AlertRepo {
	return &AlertRepo{pool: pool}
}

func (r *AlertRepo) List(ctx context.Context, userID string) ([]model.Alert, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, keyword, category_id, organization_id, education_level, is_active, created_at
		FROM alerts WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list alerts: %w", err)
	}
	defer rows.Close()

	var alerts []model.Alert
	for rows.Next() {
		var a model.Alert
		if err := rows.Scan(&a.ID, &a.UserID, &a.Keyword, &a.CategoryID, &a.OrganizationID,
			&a.EducationLevel, &a.IsActive, &a.CreatedAt); err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}
	return alerts, nil
}

func (r *AlertRepo) Create(ctx context.Context, a *model.Alert) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO alerts (user_id, keyword, category_id, organization_id, education_level)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`,
		a.UserID, a.Keyword, a.CategoryID, a.OrganizationID, a.EducationLevel,
	).Scan(&a.ID, &a.CreatedAt)
}

func (r *AlertRepo) Delete(ctx context.Context, id, userID string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM alerts WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete alert: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("alert not found")
	}
	return nil
}

func (r *AlertRepo) GetByID(ctx context.Context, id, userID string) (*model.Alert, error) {
	a := &model.Alert{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, keyword, category_id, organization_id, education_level, is_active, created_at
		FROM alerts WHERE id = $1 AND user_id = $2`, id, userID,
	).Scan(&a.ID, &a.UserID, &a.Keyword, &a.CategoryID, &a.OrganizationID,
		&a.EducationLevel, &a.IsActive, &a.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return a, err
}

func (r *AlertRepo) Toggle(ctx context.Context, id, userID string) (bool, error) {
	var active bool
	err := r.pool.QueryRow(ctx,
		`UPDATE alerts SET is_active = NOT is_active WHERE id = $1 AND user_id = $2 RETURNING is_active`, id, userID,
	).Scan(&active)
	if err == pgx.ErrNoRows {
		return false, fmt.Errorf("alert not found")
	}
	return active, err
}
