package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fuad71/job-circular-api/internal/model"
)

type CircularRepo struct {
	pool *pgxpool.Pool
}

func NewCircularRepo(pool *pgxpool.Pool) *CircularRepo {
	return &CircularRepo{pool: pool}
}

func (r *CircularRepo) List(ctx context.Context, f model.CircularFilter) ([]model.CircularListItem, int, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	argIdx := 1

	if f.Status != "" && f.Status != "all" {
		where = append(where, fmt.Sprintf("c.status = $%d", argIdx))
		args = append(args, f.Status)
		argIdx++
	}

	if f.CategorySlug != "" && f.CategorySlug != "all" {
		where = append(where, fmt.Sprintf("cat.slug = $%d", argIdx))
		args = append(args, f.CategorySlug)
		argIdx++
	}

	if f.Search != "" {
		where = append(where, fmt.Sprintf("(c.title ILIKE $%d OR c.organization_name ILIKE $%d)", argIdx, argIdx+1))
		args = append(args, "%"+f.Search+"%", "%"+f.Search+"%")
		argIdx += 2
	}

	if f.DeadlineFrom != "" {
		where = append(where, fmt.Sprintf("c.application_deadline >= $%d", argIdx))
		args = append(args, f.DeadlineFrom)
		argIdx++
	}

	if f.DeadlineTo != "" {
		where = append(where, fmt.Sprintf("c.application_deadline <= $%d", argIdx))
		args = append(args, f.DeadlineTo)
		argIdx++
	}

	if f.Education != "" {
		where = append(where, fmt.Sprintf("c.education_level ILIKE $%d", argIdx))
		args = append(args, "%"+f.Education+"%")
		argIdx++
	}

	if f.Gender != "" {
		where = append(where, fmt.Sprintf("c.gender = $%d", argIdx))
		args = append(args, f.Gender)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	// Count
	var total int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM circulars c
		LEFT JOIN categories cat ON cat.id = c.category_id WHERE %s`, whereClause)
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count circulars: %w", err)
	}

	// Sort
	sortSQL := "c.published_date DESC"
	switch f.Sort {
	case "deadline_asc":
		sortSQL = "c.application_deadline ASC NULLS LAST"
	case "views_desc":
		sortSQL = "c.view_count DESC"
	}

	offset := (f.Page - 1) * f.Limit
	listArgs := append(args, f.Limit, offset)
	limitIdx := argIdx
	offsetIdx := argIdx + 1

	query := fmt.Sprintf(`SELECT c.id, c.title, c.organization_name,
		cat.id, cat.name, cat.slug,
		c.vacancy, c.salary_display, c.published_date, c.application_deadline,
		c.apply_via, c.location, c.district, c.job_type, c.status, c.is_featured
		FROM circulars c
		LEFT JOIN categories cat ON cat.id = c.category_id
		WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		whereClause, sortSQL, limitIdx, offsetIdx)

	rows, err := r.pool.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list circulars: %w", err)
	}
	defer rows.Close()

	var items []model.CircularListItem
	for rows.Next() {
		var item model.CircularListItem
		var catID *int
		var catName, catSlug *string

		if err := rows.Scan(
			&item.ID, &item.Title, &item.OrganizationName,
			&catID, &catName, &catSlug,
			&item.Vacancy, &item.SalaryDisplay, &item.PublishedDate, &item.ApplicationDeadline,
			&item.ApplyVia, &item.Location, &item.District, &item.JobType, &item.Status, &item.IsFeatured,
		); err != nil {
			return nil, 0, fmt.Errorf("scan circular: %w", err)
		}

		if catID != nil {
			item.Category = &model.Category{ID: *catID}
			if catName != nil {
				item.Category.Name = *catName
			}
			if catSlug != nil {
				item.Category.Slug = *catSlug
			}
		}

		items = append(items, item)
	}

	return items, total, nil
}

func (r *CircularRepo) GetFeatured(ctx context.Context, limit int) ([]model.CircularListItem, error) {
	query := `SELECT c.id, c.title, c.organization_name,
		cat.id, cat.name, cat.slug,
		c.vacancy, c.salary_display, c.published_date, c.application_deadline,
		c.apply_via, c.location, c.district, c.job_type, c.status, c.is_featured
		FROM circulars c
		LEFT JOIN categories cat ON cat.id = c.category_id
		WHERE c.is_featured = true AND c.status = 'active'
		ORDER BY c.published_date DESC LIMIT $1`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get featured: %w", err)
	}
	defer rows.Close()

	var items []model.CircularListItem
	for rows.Next() {
		var item model.CircularListItem
		var catID *int
		var catName, catSlug *string

		if err := rows.Scan(
			&item.ID, &item.Title, &item.OrganizationName,
			&catID, &catName, &catSlug,
			&item.Vacancy, &item.SalaryDisplay, &item.PublishedDate, &item.ApplicationDeadline,
			&item.ApplyVia, &item.Location, &item.District, &item.JobType, &item.Status, &item.IsFeatured,
		); err != nil {
			return nil, fmt.Errorf("scan featured: %w", err)
		}

		if catID != nil {
			item.Category = &model.Category{ID: *catID}
			if catName != nil {
				item.Category.Name = *catName
			}
			if catSlug != nil {
				item.Category.Slug = *catSlug
			}
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *CircularRepo) GetByID(ctx context.Context, id string) (*model.Circular, error) {
	c := &model.Circular{}
	query := `SELECT id, external_id, source, source_url, title, title_bn,
		organization_id, organization_name, category_id,
		vacancy, job_type, gender, age_min, age_max, age_note,
		education_level, education_detail, experience_years, experience_note,
		salary_min, salary_max, salary_grade, salary_display,
		location, district, division,
		published_date, application_deadline, exam_date,
		apply_url, apply_via, teletalk_code,
		description, requirements, circular_image_url, circular_pdf_url,
		status, is_featured, is_verified, view_count, content_hash,
		created_at, updated_at
		FROM circulars WHERE id = $1`

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.ExternalID, &c.Source, &c.SourceURL, &c.Title, &c.TitleBn,
		&c.OrganizationID, &c.OrganizationName, &c.CategoryID,
		&c.Vacancy, &c.JobType, &c.Gender, &c.AgeMin, &c.AgeMax, &c.AgeNote,
		&c.EducationLevel, &c.EducationDetail, &c.ExperienceYears, &c.ExperienceNote,
		&c.SalaryMin, &c.SalaryMax, &c.SalaryGrade, &c.SalaryDisplay,
		&c.Location, &c.District, &c.Division,
		&c.PublishedDate, &c.ApplicationDeadline, &c.ExamDate,
		&c.ApplyURL, &c.ApplyVia, &c.TeletalkCode,
		&c.Description, &c.Requirements, &c.CircularImageURL, &c.CircularPDFURL,
		&c.Status, &c.IsFeatured, &c.IsVerified, &c.ViewCount, &c.ContentHash,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get circular by id: %w", err)
	}

	// Increment view count
	_, _ = r.pool.Exec(ctx, `UPDATE circulars SET view_count = view_count + 1 WHERE id = $1`, id)

	return c, nil
}

func (r *CircularRepo) GetCategoryByID(ctx context.Context, id int) (*model.Category, error) {
	c := &model.Category{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, name_bn, slug, icon, sort_order, created_at FROM categories WHERE id = $1`, id,
	).Scan(&c.ID, &c.Name, &c.NameBn, &c.Slug, &c.Icon, &c.SortOrder, &c.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return c, err
}

func (r *CircularRepo) GetOrganizationByID(ctx context.Context, id int) (*model.Organization, error) {
	o := &model.Organization{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, name_bn, type, website, logo_url, apply_base_url, created_at FROM organizations WHERE id = $1`, id,
	).Scan(&o.ID, &o.Name, &o.NameBn, &o.Type, &o.Website, &o.LogoURL, &o.ApplyBaseURL, &o.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return o, err
}

func (r *CircularRepo) Create(ctx context.Context, c *model.Circular) error {
	query := `INSERT INTO circulars (
		source, title, organization_name, organization_id, category_id,
		vacancy, job_type, gender, age_min, age_max, age_note,
		education_level, education_detail, experience_years, experience_note,
		salary_min, salary_max, salary_grade, salary_display,
		location, district, division,
		published_date, application_deadline, exam_date,
		apply_url, apply_via, teletalk_code,
		description, requirements, circular_image_url, circular_pdf_url,
		status, is_featured, source_url
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,$36)
		RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, query,
		c.Source, c.Title, c.OrganizationName, c.OrganizationID, c.CategoryID,
		c.Vacancy, c.JobType, c.Gender, c.AgeMin, c.AgeMax, c.AgeNote,
		c.EducationLevel, c.EducationDetail, c.ExperienceYears, c.ExperienceNote,
		c.SalaryMin, c.SalaryMax, c.SalaryGrade, c.SalaryDisplay,
		c.Location, c.District, c.Division,
		c.PublishedDate, c.ApplicationDeadline, c.ExamDate,
		c.ApplyURL, c.ApplyVia, c.TeletalkCode,
		c.Description, c.Requirements, c.CircularImageURL, c.CircularPDFURL,
		c.Status, c.IsFeatured, c.SourceURL,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func (r *CircularRepo) Update(ctx context.Context, c *model.Circular) error {
	query := `UPDATE circulars SET
		source=$1, title=$2, organization_name=$3, organization_id=$4, category_id=$5,
		vacancy=$6, job_type=$7, gender=$8, age_min=$9, age_max=$10, age_note=$11,
		education_level=$12, education_detail=$13, experience_years=$14, experience_note=$15,
		salary_min=$16, salary_max=$17, salary_grade=$18, salary_display=$19,
		location=$20, district=$21, division=$22,
		published_date=$23, application_deadline=$24, exam_date=$25,
		apply_url=$26, apply_via=$27, teletalk_code=$28,
		description=$29, requirements=$30, circular_image_url=$31, circular_pdf_url=$32,
		status=$33, is_featured=$34, source_url=$35
		WHERE id=$36`
	_, err := r.pool.Exec(ctx, query,
		c.Source, c.Title, c.OrganizationName, c.OrganizationID, c.CategoryID,
		c.Vacancy, c.JobType, c.Gender, c.AgeMin, c.AgeMax, c.AgeNote,
		c.EducationLevel, c.EducationDetail, c.ExperienceYears, c.ExperienceNote,
		c.SalaryMin, c.SalaryMax, c.SalaryGrade, c.SalaryDisplay,
		c.Location, c.District, c.Division,
		c.PublishedDate, c.ApplicationDeadline, c.ExamDate,
		c.ApplyURL, c.ApplyVia, c.TeletalkCode,
		c.Description, c.Requirements, c.CircularImageURL, c.CircularPDFURL,
		c.Status, c.IsFeatured, c.SourceURL,
		c.ID,
	)
	return err
}

func (r *CircularRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM circulars WHERE id = $1`, id)
	return err
}

func (r *CircularRepo) ToggleFeatured(ctx context.Context, id string) (bool, error) {
	var featured bool
	err := r.pool.QueryRow(ctx,
		`UPDATE circulars SET is_featured = NOT is_featured WHERE id = $1 RETURNING is_featured`, id,
	).Scan(&featured)
	return featured, err
}

func (r *CircularRepo) ListCategories(ctx context.Context) ([]model.Category, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name, name_bn, slug, icon, sort_order, created_at FROM categories ORDER BY sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.NameBn, &c.Slug, &c.Icon, &c.SortOrder, &c.CreatedAt); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, nil
}

func (r *CircularRepo) ListOrganizations(ctx context.Context) ([]model.Organization, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name, name_bn, type, website, logo_url, apply_base_url, created_at FROM organizations ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []model.Organization
	for rows.Next() {
		var o model.Organization
		if err := rows.Scan(&o.ID, &o.Name, &o.NameBn, &o.Type, &o.Website, &o.LogoURL, &o.ApplyBaseURL, &o.CreatedAt); err != nil {
			return nil, err
		}
		orgs = append(orgs, o)
	}
	return orgs, nil
}

// Admin stats
func (r *CircularRepo) GetStats(ctx context.Context) (map[string]int, error) {
	stats := map[string]int{}
	var val int

	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM circulars WHERE status='active'`).Scan(&val); err != nil {
		return nil, err
	}
	stats["active_circulars"] = val

	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM circulars`).Scan(&val); err != nil {
		return nil, err
	}
	stats["total_circulars"] = val

	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&val); err != nil {
		return nil, err
	}
	stats["total_users"] = val

	return stats, nil
}

func (r *CircularRepo) ListUsers(ctx context.Context) ([]model.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, email, phone, district, role, is_verified, created_at FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.District, &u.Role, &u.IsVerified, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *CircularRepo) ListScrapeLogs(ctx context.Context, limit int) ([]model.ScrapeLog, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, source, started_at, finished_at, status, total_fetched, new_inserted, updated, skipped, error_message
		FROM scrape_logs ORDER BY started_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []model.ScrapeLog
	for rows.Next() {
		var l model.ScrapeLog
		if err := rows.Scan(&l.ID, &l.Source, &l.StartedAt, &l.FinishedAt, &l.Status,
			&l.TotalFetched, &l.NewInserted, &l.Updated, &l.Skipped, &l.ErrorMessage); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}
