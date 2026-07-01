package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/fuad71/job-circular-api/internal/model"
	"github.com/fuad71/job-circular-api/internal/repository"
	"github.com/fuad71/job-circular-api/pkg/response"
)

type CircularHandler struct {
	circularRepo *repository.CircularRepo
}

func NewCircularHandler(cr *repository.CircularRepo) *CircularHandler {
	return &CircularHandler{circularRepo: cr}
}

// GET /circulars
func (h *CircularHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	f := model.CircularFilter{
		Page:         page,
		Limit:        limit,
		CategorySlug: q.Get("category"),
		Status:       q.Get("status"),
		Search:       q.Get("search"),
		DeadlineFrom: q.Get("deadline_from"),
		DeadlineTo:   q.Get("deadline_to"),
		Education:    q.Get("education"),
		Gender:       q.Get("gender"),
		Sort:         q.Get("sort"),
	}
	if f.Status == "" {
		f.Status = "active"
	}

	items, total, err := h.circularRepo.List(r.Context(), f)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch circulars")
		return
	}
	if items == nil {
		items = []model.CircularListItem{}
	}

	response.Paginated(w, http.StatusOK, items, page, limit, total)
}

// GET /circulars/featured
func (h *CircularHandler) Featured(w http.ResponseWriter, r *http.Request) {
	items, err := h.circularRepo.GetFeatured(r.Context(), 10)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch featured")
		return
	}
	if items == nil {
		items = []model.CircularListItem{}
	}
	response.JSON(w, http.StatusOK, items)
}

// GET /circulars/:id
func (h *CircularHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	c, err := h.circularRepo.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch circular")
		return
	}
	if c == nil {
		response.Error(w, http.StatusNotFound, "circular not found")
		return
	}

	if c.CategoryID != nil {
		cat, _ := h.circularRepo.GetCategoryByID(r.Context(), *c.CategoryID)
		c.Category = cat
	}
	if c.OrganizationID != nil {
		org, _ := h.circularRepo.GetOrganizationByID(r.Context(), *c.OrganizationID)
		c.Organization = org
	}

	response.JSON(w, http.StatusOK, c)
}

// POST /circulars — Admin
func (h *CircularHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title            string   `json:"title"`
		OrganizationName string   `json:"organization_name"`
		OrganizationID   *int     `json:"organization_id,omitempty"`
		CategoryID       *int     `json:"category_id,omitempty"`
		Vacancy          *int     `json:"vacancy,omitempty"`
		JobType          string   `json:"job_type,omitempty"`
		Gender           *string  `json:"gender,omitempty"`
		AgeMin           *int     `json:"age_min,omitempty"`
		AgeMax           *int     `json:"age_max,omitempty"`
		AgeNote          *string  `json:"age_note,omitempty"`
		EducationLevel   *string  `json:"education_level,omitempty"`
		EducationDetail  *string  `json:"education_detail,omitempty"`
		ExperienceYears  *int     `json:"experience_years,omitempty"`
		ExperienceNote   *string  `json:"experience_note,omitempty"`
		SalaryMin        *float64 `json:"salary_min,omitempty"`
		SalaryMax        *float64 `json:"salary_max,omitempty"`
		SalaryGrade      *string  `json:"salary_grade,omitempty"`
		SalaryDisplay    *string  `json:"salary_display,omitempty"`
		Location         string   `json:"location,omitempty"`
		District         *string  `json:"district,omitempty"`
		Division         *string  `json:"division,omitempty"`
		PublishedDate    string   `json:"published_date"`
		Deadline         *string  `json:"application_deadline,omitempty"`
		ExamDate         *string  `json:"exam_date,omitempty"`
		ApplyURL         *string  `json:"apply_url,omitempty"`
		ApplyVia         *string  `json:"apply_via,omitempty"`
		TeletalkCode     *string  `json:"teletalk_code,omitempty"`
		Description      *string  `json:"description,omitempty"`
		Requirements     *string  `json:"requirements,omitempty"`
		CircularImageURL *string  `json:"circular_image_url,omitempty"`
		CircularPDFURL   *string  `json:"circular_pdf_url,omitempty"`
		Status           string   `json:"status,omitempty"`
		IsFeatured       bool     `json:"is_featured,omitempty"`
		SourceURL        *string  `json:"source_url,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.OrganizationName) == "" || input.PublishedDate == "" {
		response.Error(w, http.StatusBadRequest, "title, organization_name, and published_date are required")
		return
	}
	if input.Status == "" {
		input.Status = "active"
	}
	if input.JobType == "" {
		input.JobType = "permanent"
	}
	if input.Location == "" {
		input.Location = "Bangladesh"
	}

	pubDate, err := parseDateStr(input.PublishedDate)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid published_date format (use YYYY-MM-DD)")
		return
	}

	c := &model.Circular{
		Source:              "manual",
		Title:               strings.TrimSpace(input.Title),
		OrganizationName:    strings.TrimSpace(input.OrganizationName),
		OrganizationID:      input.OrganizationID,
		CategoryID:          input.CategoryID,
		Vacancy:             input.Vacancy,
		JobType:             input.JobType,
		Gender:              input.Gender,
		AgeMin:              input.AgeMin,
		AgeMax:              input.AgeMax,
		AgeNote:             input.AgeNote,
		EducationLevel:      input.EducationLevel,
		EducationDetail:     input.EducationDetail,
		ExperienceYears:     input.ExperienceYears,
		ExperienceNote:      input.ExperienceNote,
		SalaryMin:           input.SalaryMin,
		SalaryMax:           input.SalaryMax,
		SalaryGrade:         input.SalaryGrade,
		SalaryDisplay:       input.SalaryDisplay,
		Location:            input.Location,
		District:            input.District,
		Division:            input.Division,
		PublishedDate:       pubDate,
		ApplicationDeadline: parseOptionalDate(input.Deadline),
		ExamDate:            parseOptionalDate(input.ExamDate),
		ApplyURL:            input.ApplyURL,
		ApplyVia:            input.ApplyVia,
		TeletalkCode:        input.TeletalkCode,
		Description:         input.Description,
		Requirements:        input.Requirements,
		CircularImageURL:    input.CircularImageURL,
		CircularPDFURL:      input.CircularPDFURL,
		Status:              input.Status,
		IsFeatured:          input.IsFeatured,
		SourceURL:           input.SourceURL,
	}

	if err := h.circularRepo.Create(r.Context(), c); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create circular")
		return
	}

	response.JSON(w, http.StatusCreated, c)
}

// PUT /circulars/:id — Admin
func (h *CircularHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	existing, err := h.circularRepo.GetByID(r.Context(), id)
	if err != nil || existing == nil {
		response.Error(w, http.StatusNotFound, "circular not found")
		return
	}

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	applyField(input, "title", &existing.Title)
	applyField(input, "organization_name", &existing.OrganizationName)
	applyField(input, "status", &existing.Status)
	applyBoolField(input, "is_featured", &existing.IsFeatured)
	applyPtrStrField(input, "description", &existing.Description)
	applyPtrStrField(input, "requirements", &existing.Requirements)
	applyPtrStrField(input, "salary_display", &existing.SalaryDisplay)
	applyPtrStrField(input, "apply_url", &existing.ApplyURL)
	applyPtrStrField(input, "apply_via", &existing.ApplyVia)
	applyField(input, "location", &existing.Location)
	applyPtrStrField(input, "circular_image_url", &existing.CircularImageURL)
	applyPtrStrField(input, "circular_pdf_url", &existing.CircularPDFURL)

	if err := h.circularRepo.Update(r.Context(), existing); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update circular")
		return
	}

	response.JSON(w, http.StatusOK, existing)
}

// DELETE /circulars/:id — Admin
func (h *CircularHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.circularRepo.Delete(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to delete circular")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// PATCH /circulars/:id/feature — Admin
func (h *CircularHandler) ToggleFeature(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	featured, err := h.circularRepo.ToggleFeatured(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "circular not found")
		return
	}
	response.JSON(w, http.StatusOK, map[string]bool{"is_featured": featured})
}

// ── Categories & Organizations ──────────────────────────────────────────────────

func (h *CircularHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.circularRepo.ListCategories(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch categories")
		return
	}
	if cats == nil {
		cats = []model.Category{}
	}
	response.JSON(w, http.StatusOK, cats)
}

func (h *CircularHandler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	orgs, err := h.circularRepo.ListOrganizations(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch organizations")
		return
	}
	if orgs == nil {
		orgs = []model.Organization{}
	}
	response.JSON(w, http.StatusOK, orgs)
}

// ── Helpers ─────────────────────────────────────────────────────────────────────

func parseDateStr(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

func parseOptionalDate(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil
	}
	return &t
}

func applyField(input map[string]interface{}, key string, target *string) {
	if v, ok := input[key]; ok {
		if s, ok := v.(string); ok {
			*target = s
		}
	}
}

func applyBoolField(input map[string]interface{}, key string, target *bool) {
	if v, ok := input[key]; ok {
		if b, ok := v.(bool); ok {
			*target = b
		}
	}
}

func applyPtrStrField(input map[string]interface{}, key string, target **string) {
	if v, ok := input[key]; ok {
		if s, ok := v.(string); ok {
			*target = &s
		}
	}
}
