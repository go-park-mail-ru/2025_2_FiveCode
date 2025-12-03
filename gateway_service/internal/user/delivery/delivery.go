package delivery

import (
	"backend/gateway_service/internal/apiutils"
	"backend/gateway_service/internal/middleware"
	"backend/gateway_service/internal/user/models"
	"backend/gateway_service/internal/validation"
	"backend/pkg/logger"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	MinUsernameLength = 3
	MaxUsernameLength = 50
)

//go:generate mockgen -source=delivery.go -destination=mock/mock_delivery.go -package=mock
type UserUsecase interface {
	GetProfile(ctx context.Context, userID uint64) (*models.User, error)
	UpdateProfile(ctx context.Context, input *models.UpdateUserInput) (*models.User, error)
	DeleteProfile(ctx context.Context, userID uint64, sessionID string) error
	GetProfileBySession(ctx context.Context, sessionID string) (*models.User, error)
}

type UserDelivery struct {
	usecase UserUsecase
}

func NewUserDelivery(usecase UserUsecase) *UserDelivery {
	return &UserDelivery{
		usecase: usecase,
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

	user, err := d.usecase.GetProfileBySession(r.Context(), cookie.Value)
	if err != nil {
		log.Error().Err(err).Msg("failed to get profile by session")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	if user == nil {
		apiutils.WriteJSON(w, http.StatusOK, nil)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, user)
}

func (d *UserDelivery) GetProfile(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	user, err := d.usecase.GetProfile(r.Context(), userID)
	if err != nil {
		log.Warn().Err(err).Msg("error getting profile")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	log.Info().Uint64("user_id", user.ID).Msg("profile retrieved successfully")
	apiutils.WriteJSON(w, http.StatusOK, user)
}

func (d *UserDelivery) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

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

	input := &models.UpdateUserInput{
		ID: userID,
	}
	if req.Username != nil {
		input.Username = req.Username
	}
	if req.Password != nil {
		input.Password = req.Password
	}
	if req.AvatarFileID != nil {
		input.AvatarFileID = req.AvatarFileID
	}

	updatedUser, err := d.usecase.UpdateProfile(r.Context(), input)
	if err != nil {
		log.Warn().Err(err).Msg("error updating profile")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	log.Info().Uint64("user_id", updatedUser.ID).Msg("profile updated successfully")
	apiutils.WriteJSON(w, http.StatusOK, updatedUser)
}

func (d *UserDelivery) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	cookie, err := r.Cookie("session_id")
	if err != nil {
		log.Error().Err(err).Msg("error getting session cookie for deletion")
		apiutils.WriteError(w, http.StatusBadRequest, "no session cookie")
		return
	}

	err = d.usecase.DeleteProfile(r.Context(), userID, cookie.Value)
	if err != nil {
		log.Error().Err(err).Msg("error deleting profile")
		apiutils.HandleGrpcError(w, err, log)
	}

	cookie.Expires = time.Now().AddDate(0, 0, -1)
	http.SetCookie(w, cookie)

	log.Info().Msg("profile deleted successfully")
	w.WriteHeader(http.StatusNoContent)
}
