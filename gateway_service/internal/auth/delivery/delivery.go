package delivery

import (
	"backend/gateway_service/internal/apiutils"
	"backend/gateway_service/internal/user/models"
	"backend/gateway_service/internal/validation"
	"backend/pkg/logger"
	"context"
	"encoding/json"
	"net/http"
	"time"

	pkgErrors "github.com/pkg/errors"
)

//go:generate mockgen -source=delivery.go -destination=mock/mock_delivery.go -package=mock
type AuthUsecase interface {
	Login(ctx context.Context, email, password string) (string, *models.User, error)
	Register(ctx context.Context, email, password string) (string, *models.User, error)
	Logout(ctx context.Context, sessionID string) error
	GetCSRFToken(ctx context.Context, sessionID string) (string, error)
}

type AuthDelivery struct {
	usecase         AuthUsecase
	SessionDuration time.Duration
}

func NewAuthDelivery(usecase AuthUsecase, sessionDuration time.Duration) *AuthDelivery {
	return &AuthDelivery{
		usecase:         usecase,
		SessionDuration: sessionDuration,
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

	sessionID, user, err := d.usecase.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		log.Warn().Err(err).Str("email", req.Email).Msg("login failed")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	expiration := time.Now().Add(d.SessionDuration)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		Expires:  expiration,
		HttpOnly: true,
	})

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

	sessionID, user, err := d.usecase.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		log.Warn().Err(err).Str("email", req.Email).Msg("registration failed")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	expiration := time.Now().Add(d.SessionDuration)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		Expires:  expiration,
		HttpOnly: true,
	})

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

	err = d.usecase.Logout(r.Context(), session.Value)
	if err != nil {
		log.Error().Err(err).Msg("logout failed")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	session.Expires = time.Now().Add(-1 * time.Hour)
	http.SetCookie(w, session)

	log.Info().Msg("user logged out successfully")
	apiutils.WriteJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func (d *AuthDelivery) GetCSRFToken(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	cookie, err := r.Cookie("session_id")
	if err != nil {
		log.Warn().Msg("csrf token request: no session cookie")
		apiutils.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	token, err := d.usecase.GetCSRFToken(r.Context(), cookie.Value)
	if err != nil {
		log.Error().Err(err).Msg("GetCSRFToken failed")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, map[string]string{"token": token})
}
