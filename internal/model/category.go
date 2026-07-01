package model

import "time"

type Category struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	NameBn    *string   `json:"name_bn,omitempty"`
	Slug      string    `json:"slug"`
	Icon      *string   `json:"icon,omitempty"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

type Organization struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	NameBn       *string   `json:"name_bn,omitempty"`
	Type         *string   `json:"type,omitempty"`
	Website      *string   `json:"website,omitempty"`
	LogoURL      *string   `json:"logo_url,omitempty"`
	ApplyBaseURL *string   `json:"apply_base_url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
