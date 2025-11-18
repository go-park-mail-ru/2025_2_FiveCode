package delivery

import (
	"backend/apiutils"
	authPB "backend/gen/go/auth"
	userPB "backend/gen/go/user"
	"backend/logger"
	"backend/middleware"
	"backend/utils"
	"backend/validation"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	MinUsernameLength = 3
	MaxUsernameLength = 50
)

//go:generate mockgen -destination=../mock/mock_user_client.go -package=mock . UserServiceClient
type UserServiceClient interface {
	GetUser(ctx context.Context, in *userPB.GetUserRequest, opts ...grpc.CallOption) (*userPB.User, error)
	UpdateUser(ctx context.Context, in *userPB.UpdateUserRequest, opts ...grpc.CallOption) (*userPB.User, error)
	DeleteUser(ctx context.Context, in *userPB.DeleteUserRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

//go:generate mockgen -destination=../mock/mock_auth_client.go -package=mock . AuthClient
type AuthClient interface {
	GetUserIDBySession(ctx context.Context, in *authPB.GetUserIDBySessionRequest, opts ...grpc.CallOption) (*authPB.GetUserIDBySessionResponse, error)
	Logout(ctx context.Context, in *authPB.LogoutRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type UserDelivery struct {
	UserClient UserServiceClient
	AuthClient AuthClient
}

func NewUserDelivery(u UserServiceClient, a AuthClient) *UserDelivery {
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
		apiutils.HandleGrpcError(w, err, log)
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
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	user := utils.ProtoUserToModel(userResp)

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
		log.Warn().Err(err).Msg("error updating profile")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	user := utils.ProtoUserToModel(updatedUser)

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
		log.Warn().Err(err).Msg("error getting profile from user service")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	user := utils.ProtoUserToModel(grpcResp)
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
		log.Error().Err(err).Msg("error deleting user via user service")
		apiutils.HandleGrpcError(w, err, log)
		return
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
