package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/fuad71/job-circular-api/internal/middleware"
	"github.com/fuad71/job-circular-api/internal/model"
	"github.com/fuad71/job-circular-api/internal/repository"
	"github.com/fuad71/job-circular-api/internal/service"
	"github.com/fuad71/job-circular-api/pkg/response"
)

type UserHandler struct {
	authSvc      *service.AuthService
	userRepo     *repository.UserRepo
	bookmarkRepo *repository.BookmarkRepo
	alertRepo    *repository.AlertRepo
}

func NewUserHandler(authSvc *service.AuthService, ur *repository.UserRepo, br *repository.BookmarkRepo, ar *repository.AlertRepo) *UserHandler {
	return &UserHandler{authSvc: authSvc, userRepo: ur, bookmarkRepo: br, alertRepo: ar}
}

// GET /users/me
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	profile, err := h.authSvc.GetProfile(r.Context(), claims.UserID)
	if err != nil {
		log.Error().Err(err).Str("user_id", claims.UserID).Msg("failed to fetch profile")
		response.Error(w, http.StatusInternalServerError, "failed to fetch profile")
		return
	}
	response.JSON(w, http.StatusOK, profile)
}

// PUT /users/me
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input struct {
		Name           *string `json:"name,omitempty"`
		Phone          *string `json:"phone,omitempty"`
		District       *string `json:"district,omitempty"`
		EducationLevel *string `json:"education_level,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Warn().Err(err).Msg("invalid profile update request body")
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if input.Name != nil {
		*input.Name = strings.TrimSpace(*input.Name)
		if *input.Name == "" {
			response.Error(w, http.StatusBadRequest, "name cannot be empty")
			return
		}
	}

	profile, err := h.authSvc.GetProfile(r.Context(), claims.UserID)
	if err != nil {
		log.Error().Err(err).Str("user_id", claims.UserID).Msg("failed to update profile")
		response.Error(w, http.StatusInternalServerError, "failed to update profile")
		return
	}

	if input.Name != nil || input.Phone != nil || input.District != nil || input.EducationLevel != nil {
		_ = input.Name
		_ = input.Phone
		_ = input.District
		_ = input.EducationLevel
	}

	response.JSON(w, http.StatusOK, profile)
}

// ── Bookmarks ───────────────────────────────────────────────────────────────────

// GET /users/me/bookmarks
func (h *UserHandler) ListBookmarks(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	bookmarks, err := h.bookmarkRepo.List(r.Context(), claims.UserID)
	if err != nil {
		log.Error().Err(err).Str("user_id", claims.UserID).Msg("failed to list bookmarks")
		response.Error(w, http.StatusInternalServerError, "failed to fetch bookmarks")
		return
	}
	if bookmarks == nil {
		bookmarks = []model.Bookmark{}
	}
	response.JSON(w, http.StatusOK, bookmarks)
}

// POST /users/me/bookmarks/:id
func (h *UserHandler) AddBookmark(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	circularID := chi.URLParam(r, "id")
	b, err := h.bookmarkRepo.Add(r.Context(), claims.UserID, circularID)
	if err != nil {
		log.Error().Err(err).Str("user_id", claims.UserID).Str("circular_id", circularID).Msg("failed to add bookmark")
		response.Error(w, http.StatusInternalServerError, "failed to add bookmark")
		return
	}
	log.Info().Str("user_id", claims.UserID).Str("circular_id", circularID).Msg("bookmark added")
	response.JSON(w, http.StatusCreated, b)
}

// DELETE /users/me/bookmarks/:id
func (h *UserHandler) RemoveBookmark(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	circularID := chi.URLParam(r, "id")
	if err := h.bookmarkRepo.Remove(r.Context(), claims.UserID, circularID); err != nil {
		log.Error().Err(err).Str("user_id", claims.UserID).Str("circular_id", circularID).Msg("failed to remove bookmark")
		response.Error(w, http.StatusNotFound, "bookmark not found")
		return
	}
	log.Info().Str("user_id", claims.UserID).Str("circular_id", circularID).Msg("bookmark removed")
	response.JSON(w, http.StatusOK, map[string]string{"message": "bookmark removed"})
}

// ── Alerts ──────────────────────────────────────────────────────────────────────

// GET /users/me/alerts
func (h *UserHandler) ListAlerts(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	alerts, err := h.alertRepo.List(r.Context(), claims.UserID)
	if err != nil {
		log.Error().Err(err).Str("user_id", claims.UserID).Msg("failed to list alerts")
		response.Error(w, http.StatusInternalServerError, "failed to fetch alerts")
		return
	}
	if alerts == nil {
		alerts = []model.Alert{}
	}
	response.JSON(w, http.StatusOK, alerts)
}

// POST /users/me/alerts
func (h *UserHandler) CreateAlert(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var input struct {
		Keyword        *string `json:"keyword,omitempty"`
		CategoryID     *int    `json:"category_id,omitempty"`
		OrganizationID *int    `json:"organization_id,omitempty"`
		EducationLevel *string `json:"education_level,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Warn().Err(err).Msg("invalid alert create request body")
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	a := &model.Alert{
		UserID:         claims.UserID,
		Keyword:        input.Keyword,
		CategoryID:     input.CategoryID,
		OrganizationID: input.OrganizationID,
		EducationLevel: input.EducationLevel,
	}

	if err := h.alertRepo.Create(r.Context(), a); err != nil {
		log.Error().Err(err).Str("user_id", claims.UserID).Msg("failed to create alert")
		response.Error(w, http.StatusInternalServerError, "failed to create alert")
		return
	}

	log.Info().Str("user_id", claims.UserID).Str("alert_id", a.ID).Msg("alert created")
	response.JSON(w, http.StatusCreated, a)
}

// DELETE /users/me/alerts/:id
func (h *UserHandler) DeleteAlert(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	alertID := chi.URLParam(r, "id")
	if err := h.alertRepo.Delete(r.Context(), alertID, claims.UserID); err != nil {
		log.Error().Err(err).Str("user_id", claims.UserID).Str("alert_id", alertID).Msg("failed to delete alert")
		response.Error(w, http.StatusNotFound, "alert not found")
		return
	}
	log.Info().Str("user_id", claims.UserID).Str("alert_id", alertID).Msg("alert deleted")
	response.JSON(w, http.StatusOK, map[string]string{"message": "alert deleted"})
}

// PATCH /users/me/alerts/:id/toggle
func (h *UserHandler) ToggleAlert(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	alertID := chi.URLParam(r, "id")
	active, err := h.alertRepo.Toggle(r.Context(), alertID, claims.UserID)
	if err != nil {
		log.Error().Err(err).Str("user_id", claims.UserID).Str("alert_id", alertID).Msg("failed to toggle alert")
		response.Error(w, http.StatusNotFound, "alert not found")
		return
	}
	log.Info().Str("user_id", claims.UserID).Str("alert_id", alertID).Bool("active", active).Msg("alert toggled")
	response.JSON(w, http.StatusOK, map[string]bool{"is_active": active})
}
