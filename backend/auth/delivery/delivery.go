package delivery

import (
	"backend/apiutils"
	"backend/logger"
	"backend/models"
	namederrors "backend/named_errors"
	"backend/validation"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	pkgErrors "github.com/pkg/errors"
)

type AuthDelivery struct {
	SessionDuration time.Duration
	Usecase         AuthUsecase
}

//go:generate mockgen -source=delivery.go -destination=../mock/mock_delivery.go -package=mock
type AuthUsecase interface {
	Login(ctx context.Context, email string, password string) (*models.User, string, error)
	Register(ctx context.Context, email string, password string) (*models.User, string, error)
	Logout(ctx context.Context, sessionID string) error
	GenerateCSRFToken(ctx context.Context, sessionID string) (string, error) // НОВОЕ
}

func NewAuthDelivery(uc AuthUsecase, sessionDuration time.Duration) *AuthDelivery {
	return &AuthDelivery{
		SessionDuration: sessionDuration,
		Usecase:         uc,
	}
}

type loginRequest struct {
	Email    string `json:"email" valid:"required,email"`
	Password string `json:"password" valid:"required,password"`
}

type registerRequest struct {
	Email           string `json:"email" valid:"required,email"`
	Password        string `json:"password" valid:"required,password"`
	ConfirmPassword string `json:"confirm_password" valid:"required,password"`
}

func (d *AuthDelivery) Login(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close request body")
		}
	}()

	var req loginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Warn().Err(err).Msg("invalid json body")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if err := validation.ValidateStruct(req); err != nil {
		log.Warn().Err(err).Msg("validation failed")
		apiutils.WriteValidationError(w, http.StatusBadRequest, err)
		return
	}

	user, sessionID, err := d.Usecase.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		log.Warn().Err(err).Str("email", req.Email).Msg("login failed")
		apiutils.WriteError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	expiration := time.Now().Add(d.SessionDuration)
	session := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		Expires:  expiration,
		HttpOnly: true,
	}
	http.SetCookie(w, session)

	log.Info().Uint64("user_id", user.ID).Msg("user logged in successfully")
	apiutils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"user": user,
	})
}

func (d *AuthDelivery) Register(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close request body")
		}
	}()

	var req registerRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Warn().Err(err).Msg("invalid json body for registration")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if err := validation.ValidateStruct(req); err != nil {
		log.Warn().Err(err).Msg("validation failed")
		apiutils.WriteValidationError(w, http.StatusBadRequest, err)
		return
	}

	if req.Password != req.ConfirmPassword {
		log.Warn().Msg("passwords do not match")
		apiutils.WriteError(w, http.StatusBadRequest, "passwords do not match")
		return
	}

	user, sessionID, err := d.Usecase.Register(r.Context(), req.Email, req.Password)
	if pkgErrors.Is(err, namederrors.ErrUserExists) {
		log.Warn().Str("email", req.Email).Msg("user already exists")
		apiutils.WriteError(w, http.StatusBadRequest, "user already exists")
		return
	}
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("registration failed")
		apiutils.WriteError(w, http.StatusInternalServerError, "registration failed")
		return
	}

	expiration := time.Now().Add(d.SessionDuration)
	session := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		Expires:  expiration,
		HttpOnly: true,
	}
	http.SetCookie(w, session)

	log.Info().Uint64("user_id", user.ID).Msg("user registered successfully")
	apiutils.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"user": user,
	})
}

func (d *AuthDelivery) Logout(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	session, err := r.Cookie("session_id")
	if pkgErrors.Is(err, http.ErrNoCookie) {
		log.Info().Msg("no session cookie found for logout")
		apiutils.WriteError(w, http.StatusBadRequest, "no session cookie")
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("error getting session cookie")
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to get session cookie")
		return
	}

	err = d.Usecase.Logout(r.Context(), session.Value)
	if pkgErrors.Is(err, namederrors.ErrInvalidSession) {
		log.Warn().Msg("logout with invalid session")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid session")
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to logout")
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to logout")
		return
	}

	session.Expires = time.Now().Add(-1 * time.Hour)
	http.SetCookie(w, session)

	log.Info().Msg("user logged out successfully")
	apiutils.WriteJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func (d *AuthDelivery) GetCSRFToken(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	session, err := r.Cookie("session_id")
	if errors.Is(err, http.ErrNoCookie) {
		log.Warn().Msg("no session cookie for csrf token request")
		apiutils.WriteError(w, http.StatusUnauthorized, "no session cookie")
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("error getting session cookie")
		apiutils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	token, err := d.Usecase.GenerateCSRFToken(r.Context(), session.Value)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate csrf token")
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to generate csrf token")
		return
	}

	log.Info().Str("session_id", session.Value).Msg("csrf token generated successfully")
	apiutils.WriteJSON(w, http.StatusOK, map[string]string{
		"csrf_token": token,
	})
}
