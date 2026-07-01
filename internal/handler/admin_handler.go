package handler

import (
	"net/http"

	"github.com/fuad71/job-circular-api/internal/repository"
	"github.com/fuad71/job-circular-api/pkg/response"
)

type AdminHandler struct {
	circularRepo *repository.CircularRepo
}

func NewAdminHandler(cr *repository.CircularRepo) *AdminHandler {
	return &AdminHandler{circularRepo: cr}
}

// GET /admin/stats
func (h *AdminHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.circularRepo.GetStats(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch stats")
		return
	}
	response.JSON(w, http.StatusOK, stats)
}

// GET /admin/users
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.circularRepo.ListUsers(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch users")
		return
	}
	response.JSON(w, http.StatusOK, users)
}

// POST /admin/scrape/run
func (h *AdminHandler) TriggerScrape(w http.ResponseWriter, r *http.Request) {
	// Placeholder — scraper not yet implemented
	response.JSON(w, http.StatusOK, map[string]string{
		"message": "scrape triggered (not yet implemented)",
	})
}

// GET /admin/scrape/logs
func (h *AdminHandler) ScrapeLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := h.circularRepo.ListScrapeLogs(r.Context(), 50)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch scrape logs")
		return
	}
	response.JSON(w, http.StatusOK, logs)
}
