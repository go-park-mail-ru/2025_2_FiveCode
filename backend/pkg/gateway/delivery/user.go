package delivery

import (
	"backend/apiutils"
	authPB "backend/gen/go/auth"
	userPB "backend/gen/go/user"
	"backend/logger"
	"backend/middleware"
	"backend/models"
	"backend/validation"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	MinUsernameLength = 3
	MaxUsernameLength = 50
)

type UserDelivery struct {
	UserClient userPB.UserServiceClient
	AuthClient authPB.AuthClient
}

//go:generate mockgen -source=delivery.go -destination=../mock/mock_delivery.go -package=mock
type UserUsecase interface {
	GetUserBySession(ctx context.Context, session string) (*models.User, error)
	UpdateProfile(ctx context.Context, username *string, password *string, avatarFileID *uint64) (*models.User, error)
	GetProfile(ctx context.Context) (*models.User, error)
	DeleteProfile(ctx context.Context, sessionID string) error
}

func NewUserDelivery(u userPB.UserServiceClient, a authPB.AuthClient) *UserDelivery {
	return &UserDelivery{
		UserClient: u,
		AuthClient: a,
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

	authResp, err := d.AuthClient.GetUserIDBySession(r.Context(), &authPB.GetUserIDBySessionRequest{
		SessionId: cookie.Value,
	})
	if err != nil {
		log.Error().Err(err).Msg("gRPC call to Auth service failed")
		apiutils.WriteError(w, http.StatusInternalServerError, "session validation failed")
		return
	}

	if !authResp.GetIsValid() {
		apiutils.WriteJSON(w, http.StatusOK, nil)
		return
	}

	userResp, err := d.UserClient.GetUser(r.Context(), &userPB.GetUserRequest{
		UserId: authResp.GetUserId(),
	})
	if err != nil {
		log.Error().Err(err).Msg("gRPC call to User service failed")
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to get user profile")
		return
	}

	user := protoUserToModel(userResp)

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

	grpcReq := &userPB.UpdateUserRequest{UserId: userID}
	if req.Username != nil {
		grpcReq.Username = *req.Username
	}
	if req.Password != nil {
		grpcReq.Password = *req.Password
	}
	if req.AvatarFileID != nil {
		grpcReq.AvatarFileId = *req.AvatarFileID
	}

	updatedUser, err := d.UserClient.UpdateUser(r.Context(), grpcReq)
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.NotFound {
			apiutils.WriteError(w, http.StatusNotFound, "user not found")
			return
		}
		log.Error().Err(err).Msg("error updating profile via user service")
		apiutils.WriteError(w, http.StatusInternalServerError, "error updating profile")
		return
	}

	user := protoUserToModel(updatedUser)

	log.Info().Uint64("user_id", user.ID).Msg("profile updated successfully")
	apiutils.WriteJSON(w, http.StatusOK, user)
}

func (d *UserDelivery) GetProfile(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	grpcResp, err := d.UserClient.GetUser(r.Context(), &userPB.GetUserRequest{UserId: userID})
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.NotFound {
			apiutils.WriteError(w, http.StatusNotFound, "user not found")
			return
		}
		log.Error().Err(err).Msg("error getting profile from user service")
		apiutils.WriteError(w, http.StatusInternalServerError, "error getting profile")
		return
	}

	user := protoUserToModel(grpcResp)
	log.Info().Uint64("user_id", user.ID).Msg("profile retrieved successfully")

	apiutils.WriteJSON(w, http.StatusOK, user)
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

	_, err = d.UserClient.DeleteUser(r.Context(), &userPB.DeleteUserRequest{UserId: userID})
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() != codes.NotFound {
			log.Error().Err(err).Msg("error deleting profile via user service")
			apiutils.WriteError(w, http.StatusInternalServerError, "error deleting profile")
			return
		}
	}

	_, err = d.AuthClient.Logout(r.Context(), &authPB.LogoutRequest{SessionId: cookie.Value})
	if err != nil {
		log.Error().Err(err).Msg("failed to delete session after user deletion")
	}

	cookie.Expires = time.Now().AddDate(0, 0, -1)
	http.SetCookie(w, cookie)

	log.Info().Msg("profile deleted successfully")
	w.WriteHeader(http.StatusNoContent)
}

func protoUserToModel(p *userPB.User) *models.User {
	if p == nil {
		return nil
	}
	m := &models.User{
		ID:        p.Id,
		Email:     p.Email,
		Username:  p.Username,
		CreatedAt: p.CreatedAt.AsTime(),
	}
	if p.UpdatedAt != nil && p.UpdatedAt.IsValid() {
		updatedTime := p.UpdatedAt.AsTime()
		m.UpdatedAt = &updatedTime
	}
	if p.AvatarFileId != nil {
		m.AvatarFileID = p.AvatarFileId
	}
	return m
}
