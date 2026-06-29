package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/fuad71/job-circular-api/internal/config"
	"github.com/fuad71/job-circular-api/internal/model"
	"github.com/fuad71/job-circular-api/internal/repository"
)

type AuthService struct {
	userRepo *repository.UserRepo
	db       *pgxpool.Pool
	cfg      *config.Config
}

func NewAuthService(repo *repository.UserRepo, db *pgxpool.Pool, cfg *config.Config) *AuthService {
	return &AuthService{userRepo: repo, db: db, cfg: cfg}
}

// ── Password helpers ──────────────────────────────────────────────

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// ── JWT helpers ───────────────────────────────────────────────────

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func (s *AuthService) generateAccessToken(userID, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.JWTAccessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *AuthService) generateRefreshToken(userID string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.JWTRefreshTTL)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTRefreshSecret))
}

func (s *AuthService) parseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{},
		func(t *jwt.Token) (interface{}, error) {
			return []byte(s.cfg.JWTSecret), nil
		})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// ── Token helpers (PostgreSQL) ────────────────────────────────────

// storeRefreshToken upserts one refresh token row per user.
func (s *AuthService) storeRefreshToken(ctx context.Context, userID, token string) error {
	expiresAt := time.Now().Add(s.cfg.JWTRefreshTTL)
	_, err := s.db.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE
		  SET token = EXCLUDED.token,
		      expires_at = EXCLUDED.expires_at,
		      created_at = NOW()
	`, userID, token, expiresAt)
	return err
}

// getRefreshToken fetches the stored token for a user if it hasn't expired.
func (s *AuthService) getRefreshToken(ctx context.Context, userID string) (string, error) {
	var stored string
	err := s.db.QueryRow(ctx, `
		SELECT token FROM refresh_tokens
		WHERE user_id = $1 AND expires_at > NOW()
	`, userID).Scan(&stored)
	if err == pgx.ErrNoRows {
		return "", fmt.Errorf("refresh token not found or expired")
	}
	return stored, err
}

// deleteRefreshToken removes the refresh token row for a user (logout).
func (s *AuthService) deleteRefreshToken(ctx context.Context, userID string) error {
	_, err := s.db.Exec(ctx,
		`DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	return err
}

// ── Random token generation ───────────────────────────────────────

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// ── Email sending (log-based placeholder) ─────────────────────────
// Replace this with real SMTP later.

func (s *AuthService) sendEmail(to, subject, body string) {
	// TODO: implement SMTP sending using s.cfg.SMTPHost, s.cfg.SMTPUser, etc.
	fmt.Printf("[EMAIL] To: %s | Subject: %s\n%s\n", to, subject, body)
}

func (s *AuthService) sendVerificationEmail(email, name, token string) {
	link := fmt.Sprintf("%s/verify-email?token=%s", s.cfg.FrontendURL, token)
	body := fmt.Sprintf("Hi %s,\n\nClick the link to verify your email:\n%s\n", name, link)
	s.sendEmail(email, "Verify your email", body)
}

func (s *AuthService) sendPasswordResetEmail(email, name, token string) {
	link := fmt.Sprintf("%s/reset-password?token=%s", s.cfg.FrontendURL, token)
	body := fmt.Sprintf("Hi %s,\n\nClick the link to reset your password:\n%s\n\nThis link expires in 15 minutes.\n", name, link)
	s.sendEmail(email, "Reset your password", body)
}

// ── Auth operations ───────────────────────────────────────────────

type RegisterInput struct {
	Name           string  `json:"name"`
	Email          string  `json:"email"`
	Password       string  `json:"password"`
	Phone          *string `json:"phone,omitempty"`
	District       *string `json:"district,omitempty"`
	EducationLevel *string `json:"education_level,omitempty"`
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*model.User, error) {
	// Check if email already exists
	taken, err := s.userRepo.IsEmailTaken(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("check email: %w", err)
	}
	if taken {
		return nil, fmt.Errorf("email already registered")
	}

	hash, err := HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		Name:           input.Name,
		Email:          input.Email,
		PasswordHash:   hash,
		Role:           "user",
		IsVerified:     true,
		VerifyToken:    nil,
		Phone:          input.Phone,
		District:       input.District,
		EducationLevel: input.EducationLevel,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// Send verification email (async-friendly, non-blocking)
	// go s.sendVerificationEmail(user.Email, user.Name, token)

	return user, nil
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginOutput struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         *model.UserProfile
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	if !CheckPassword(input.Password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid email or password")
	}

	accessToken, err := s.generateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	if err := s.storeRefreshToken(ctx, user.ID, refreshToken); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	profile := userToProfile(user)
	return &LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         profile,
	}, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	user, err := s.userRepo.GetByVerifyToken(ctx, token)
	if err != nil {
		return fmt.Errorf("get user by token: %w", err)
	}
	if user == nil {
		return fmt.Errorf("invalid or expired verification token")
	}

	return s.userRepo.MarkVerified(ctx, user.ID)
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		// Don't reveal whether the email exists
		return nil
	}

	token := generateToken()
	exp := time.Now().Add(15 * time.Minute)

	if err := s.userRepo.SetResetToken(ctx, user.ID, token, exp); err != nil {
		return fmt.Errorf("set reset token: %w", err)
	}

	s.sendPasswordResetEmail(user.Email, user.Name, token)
	return nil
}

type ResetPasswordInput struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func (s *AuthService) ResetPassword(ctx context.Context, input ResetPasswordInput) error {
	user, err := s.userRepo.GetByResetToken(ctx, input.Token)
	if err != nil {
		return fmt.Errorf("get user by reset token: %w", err)
	}
	if user == nil {
		return fmt.Errorf("invalid or expired reset token")
	}

	hash, err := HashPassword(input.NewPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return s.userRepo.UpdatePassword(ctx, user.ID, hash)
}

func (s *AuthService) RefreshToken(ctx context.Context, oldRefreshToken string) (*LoginOutput, error) {
	// Parse the refresh token
	token, err := jwt.Parse(oldRefreshToken, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWTRefreshSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid refresh token claims")
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid refresh token subject")
	}

	// Check against stored token in PostgreSQL
	stored, err := s.getRefreshToken(ctx, userID)
	if err != nil || stored != oldRefreshToken {
		return nil, fmt.Errorf("refresh token expired or revoked")
	}

	// Get user to fetch role
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Issue new pair
	accessToken, err := s.generateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	newRefreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	if err := s.storeRefreshToken(ctx, user.ID, newRefreshToken); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	profile := userToProfile(user)
	return &LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         profile,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, userID string) error {
	return s.deleteRefreshToken(ctx, userID)
}

func (s *AuthService) GetProfile(ctx context.Context, userID string) (*model.UserProfile, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return userToProfile(user), nil
}

// GetJWTSecret exposes the JWT secret for the middleware.
func (s *AuthService) GetJWTSecret() string {
	return s.cfg.JWTSecret
}

// ── Helpers ───────────────────────────────────────────────────────

func userToProfile(u *model.User) *model.UserProfile {
	return &model.UserProfile{
		ID:             u.ID,
		Name:           u.Name,
		Email:          u.Email,
		Role:           u.Role,
		IsVerified:     u.IsVerified,
		Phone:          u.Phone,
		District:       u.District,
		EducationLevel: u.EducationLevel,
		CreatedAt:      u.CreatedAt,
	}
}
