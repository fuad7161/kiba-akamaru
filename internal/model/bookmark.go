package model

import "time"

type Bookmark struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	CircularID string    `json:"circular_id"`
	Note       *string   `json:"note,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type Alert struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Keyword        *string   `json:"keyword,omitempty"`
	CategoryID     *int      `json:"category_id,omitempty"`
	OrganizationID *int      `json:"organization_id,omitempty"`
	EducationLevel *string   `json:"education_level,omitempty"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
}

type ScrapeLog struct {
	ID           int        `json:"id"`
	Source       string     `json:"source"`
	StartedAt    time.Time  `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
	Status       string     `json:"status"`
	TotalFetched int        `json:"total_fetched"`
	NewInserted  int        `json:"new_inserted"`
	Updated      int        `json:"updated"`
	Skipped      int        `json:"skipped"`
	ErrorMessage *string    `json:"error_message,omitempty"`
	Meta         []byte     `json:"meta,omitempty"`
}
