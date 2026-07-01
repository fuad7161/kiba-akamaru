package model

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Circular struct {
	ID                  string     `json:"id"`
	ExternalID          *string    `json:"external_id,omitempty"`
	Source              string     `json:"source"`
	SourceURL           *string    `json:"source_url,omitempty"`
	Title               string     `json:"title"`
	TitleBn             *string    `json:"title_bn,omitempty"`
	OrganizationID      *int       `json:"organization_id,omitempty"`
	OrganizationName    string     `json:"organization_name"`
	CategoryID          *int       `json:"category_id,omitempty"`
	Vacancy             *int       `json:"vacancy,omitempty"`
	JobType             string     `json:"job_type"`
	Gender              *string    `json:"gender,omitempty"`
	AgeMin              *int       `json:"age_min,omitempty"`
	AgeMax              *int       `json:"age_max,omitempty"`
	AgeNote             *string    `json:"age_note,omitempty"`
	EducationLevel      *string    `json:"education_level,omitempty"`
	EducationDetail     *string    `json:"education_detail,omitempty"`
	ExperienceYears     *int       `json:"experience_years,omitempty"`
	ExperienceNote      *string    `json:"experience_note,omitempty"`
	SalaryMin           *float64   `json:"salary_min,omitempty"`
	SalaryMax           *float64   `json:"salary_max,omitempty"`
	SalaryGrade         *string    `json:"salary_grade,omitempty"`
	SalaryDisplay       *string    `json:"salary_display,omitempty"`
	Location            string     `json:"location"`
	District            *string    `json:"district,omitempty"`
	Division            *string    `json:"division,omitempty"`
	PublishedDate       time.Time  `json:"published_date"`
	ApplicationDeadline *time.Time `json:"application_deadline,omitempty"`
	ExamDate            *time.Time `json:"exam_date,omitempty"`
	ApplyURL            *string    `json:"apply_url,omitempty"`
	ApplyVia            *string    `json:"apply_via,omitempty"`
	TeletalkCode        *string    `json:"teletalk_code,omitempty"`
	Description         *string    `json:"description,omitempty"`
	Requirements        *string    `json:"requirements,omitempty"`
	CircularImageURL    *string    `json:"circular_image_url,omitempty"`
	CircularPDFURL      *string    `json:"circular_pdf_url,omitempty"`
	Status              string     `json:"status"`
	IsFeatured          bool       `json:"is_featured"`
	IsVerified          bool       `json:"is_verified"`
	ViewCount           int        `json:"view_count"`
	ContentHash         *string    `json:"-"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`

	// Joined fields (for list/detail response)
	Category     *Category     `json:"category,omitempty"`
	Organization *Organization `json:"organization,omitempty"`
}

// CircularListItem is a lighter version for list views
type CircularListItem struct {
	ID                  string     `json:"id"`
	Title               string     `json:"title"`
	OrganizationName    string     `json:"organization_name"`
	Category            *Category  `json:"category,omitempty"`
	Vacancy             *int       `json:"vacancy,omitempty"`
	SalaryDisplay       *string    `json:"salary_display,omitempty"`
	PublishedDate       time.Time  `json:"published_date"`
	ApplicationDeadline *time.Time `json:"application_deadline,omitempty"`
	ApplyVia            *string    `json:"apply_via,omitempty"`
	Location            string     `json:"location"`
	District            *string    `json:"district,omitempty"`
	JobType             string     `json:"job_type"`
	Status              string     `json:"status"`
	IsFeatured          bool       `json:"is_featured"`
}

// CircularFilter holds all possible query parameters for listing
type CircularFilter struct {
	Page         int
	Limit        int
	CategorySlug string
	Status       string
	Search       string
	DeadlineFrom string // YYYY-MM-DD
	DeadlineTo   string // YYYY-MM-DD
	Education    string
	Gender       string
	Sort         string
}

// Nullable time scanner helper
func scanTime(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time
	return &t
}

func scanDate(d pgtype.Date) *time.Time {
	if !d.Valid {
		return nil
	}
	t := d.Time
	return &t
}
