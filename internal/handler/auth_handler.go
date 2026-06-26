package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/fuad71/job-circular-api/internal/middleware"
	"github.com/fuad71/job-circular-api/internal/service"
	"github.com/fuad71/job-circular-api/pkg/response"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input service.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(input.Email)

	if input.Name == "" || input.Email == "" || input.Password == "" {
		response.Error(w, http.StatusBadRequest, "name, email and password are required")
		return
	}
	if len(input.Password) < 6 {
		response.Error(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}

	user, err := h.authSvc.Register(r.Context(), input)
	if err != nil {
		if strings.Contains(err.Error(), "already registered") {
			response.Error(w, http.StatusConflict, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "registration failed")
		return
	}

	response.JSON(w, http.StatusCreated, map[string]interface{}{
		"id":      user.ID,
		"message": "registration successful. check your email for verification",
	})
}

// POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input service.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input.Email = strings.TrimSpace(input.Email)
	if input.Email == "" || input.Password == "" {
		response.Error(w, http.StatusBadRequest, "email and password are required")
		return
	}

	output, err := h.authSvc.Login(r.Context(), input)
	if err != nil {
		if strings.Contains(err.Error(), "invalid email or password") {
			response.Error(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		response.Error(w, http.StatusInternalServerError, "login failed")
		return
	}

	// Set refresh token as httpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    output.RefreshToken,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   false, // set true in production
		SameSite: http.SameSiteStrictMode,
		MaxAge:   7 * 24 * 60 * 60, // 7 days
	})

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"access_token": output.AccessToken,
		"user":         output.User,
	})
}

// POST /auth/logout — JWT required
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.authSvc.Logout(r.Context(), claims.UserID); err != nil {
		response.Error(w, http.StatusInternalServerError, "logout failed")
		return
	}

	// Clear refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth",
		HttpOnly: true,
		MaxAge:   -1,
	})

	response.JSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// GET /auth/verify-email?token=xxx — Public
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		response.Error(w, http.StatusBadRequest, "missing token")
		return
	}

	if err := h.authSvc.VerifyEmail(r.Context(), token); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid or expired token")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "email verified successfully"})
}

// POST /auth/forgot-password — Public
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input.Email = strings.TrimSpace(input.Email)
	if input.Email == "" {
		response.Error(w, http.StatusBadRequest, "email is required")
		return
	}

	if err := h.authSvc.ForgotPassword(r.Context(), input.Email); err != nil {
		response.Error(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "if an account with that email exists, a reset link has been sent",
	})
}

// POST /auth/reset-password — Public
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var input service.ResetPasswordInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if input.Token == "" || input.NewPassword == "" {
		response.Error(w, http.StatusBadRequest, "token and new_password are required")
		return
	}
	if len(input.NewPassword) < 6 {
		response.Error(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}

	if err := h.authSvc.ResetPassword(r.Context(), input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid or expired token")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "password reset successfully"})
}

// POST /auth/refresh — reads refresh_token cookie
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		response.Error(w, http.StatusUnauthorized, "missing refresh token")
		return
	}

	output, err := h.authSvc.RefreshToken(r.Context(), cookie.Value)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	// Update the cookie with new refresh token
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    output.RefreshToken,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   7 * 24 * 60 * 60,
	})

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"access_token": output.AccessToken,
		"user":         output.User,
	})
}

// GET /auth/me — JWT required
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	profile, err := h.authSvc.GetProfile(r.Context(), claims.UserID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch profile")
		return
	}

	response.JSON(w, http.StatusOK, profile)
}
