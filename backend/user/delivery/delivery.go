package delivery

import (
	"backend/apiutils"
	"backend/logger"
	"backend/models"
	namederrors "backend/named_errors"
	"backend/validation"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

const (
	MinUsernameLength = 3
	MaxUsernameLength = 50
)

type UserDelivery struct {
	Usecase UserUsecase
}

//go:generate mockgen -source=delivery.go -destination=../mock/mock_delivery.go -package=mock
type UserUsecase interface {
	GetUserBySession(ctx context.Context, session string) (*models.User, error)
	UpdateProfile(ctx context.Context, username *string, password *string, avatarFileID *uint64) (*models.User, error)
	GetProfile(ctx context.Context) (*models.User, error)
}

func NewUserDelivery(u UserUsecase) *UserDelivery {
	return &UserDelivery{
		Usecase: u,
	}
}

type updateProfileRequest struct {
	Username     *string `json:"username"`
	Password     *string `json:"password"`
	AvatarFileID *uint64 `json:"avatar_file_id"`
}

type updatePasswordRequest struct {
	Password string `valid:"password"`
}

func (d *UserDelivery) GetProfileBySession(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	cookie, err := r.Cookie("session_id")
	if errors.Is(err, http.ErrNoCookie) {
		log.Info().Msg("no session cookie found, responding with null user")
		apiutils.WriteJSON(w, http.StatusOK, nil)
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("error reading session cookie")
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to get session cookie")
		return
	}

	sessionID := cookie.Value

	user, err := d.Usecase.GetUserBySession(r.Context(), sessionID)
	if err != nil {
		if errors.Is(err, namederrors.ErrInvalidSession) || errors.Is(err, namederrors.ErrNotFound) {
			log.Warn().Err(err).Msg("failed to get user by session, responding with null user")
			apiutils.WriteJSON(w, http.StatusOK, nil)
			return
		}
		log.Error().Err(err).Msg("error getting user by session")
		apiutils.WriteJSON(w, http.StatusInternalServerError, nil)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, user)
}

func (d *UserDelivery) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close request body")
		}
	}()

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("invalid json body for profile update")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if req.Username == nil && req.Password == nil && req.AvatarFileID == nil {
		log.Warn().Msg("attempted to update profile with no fields provided")
		apiutils.WriteError(w, http.StatusBadRequest, "at least one field must be provided")
		return
	}

	if req.Username != nil {
		if len(*req.Username) < MinUsernameLength || len(*req.Username) > MaxUsernameLength {
			log.Warn().Str("username", *req.Username).Msg("invalid username length")
			apiutils.WriteError(w, http.StatusBadRequest, fmt.Sprintf("username must be between %d and %d characters", MinUsernameLength, MaxUsernameLength))
			return
		}
	}

	if req.Password != nil {
		passwordReq := updatePasswordRequest{Password: *req.Password}
		if err := validation.ValidateStruct(passwordReq); err != nil {
			log.Warn().Err(err).Msg("password validation failed")
			apiutils.WriteValidationError(w, http.StatusBadRequest, err)
			return
		}
	}

	user, err := d.Usecase.UpdateProfile(r.Context(), req.Username, req.Password, req.AvatarFileID)
	if err != nil {
		if errors.Is(err, namederrors.ErrNotFound) {
			log.Warn().Err(err).Msg("user not found for profile update")
			apiutils.WriteError(w, http.StatusNotFound, "user not found")
			return
		}
		log.Error().Err(err).Msg("error updating profile")
		apiutils.WriteError(w, http.StatusInternalServerError, "error updating profile")
		return
	}

	log.Info().Uint64("user_id", user.ID).Msg("profile updated successfully")
	apiutils.WriteJSON(w, http.StatusOK, user)
}

func (d *UserDelivery) GetProfile(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	user, err := d.Usecase.GetProfile(r.Context())
	if errors.Is(err, namederrors.ErrNotFound) {
		log.Warn().Err(err).Msg("user not found when getting profile")
		apiutils.WriteError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("error getting profile")
		apiutils.WriteError(w, http.StatusInternalServerError, "error getting profile")
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, user)
}
