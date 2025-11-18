package delivery

import (
	"backend/apiutils"
	authPB "backend/gen/go/auth"
	userPB "backend/gen/go/user"
	"backend/logger"
	"backend/utils"
	"backend/validation"
	"context"
	"encoding/json"
	"net/http"
	"time"

	pkgErrors "github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

//go:generate mockgen -destination=../mock/mock_auth_client.go -package=mock . AuthClient
type AuthClient interface {
	Login(ctx context.Context, in *authPB.LoginRequest, opts ...grpc.CallOption) (*authPB.LoginResponse, error)
	Register(ctx context.Context, in *authPB.RegisterRequest, opts ...grpc.CallOption) (*authPB.RegisterResponse, error)
	Logout(ctx context.Context, in *authPB.LogoutRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	GetCSRFToken(ctx context.Context, in *authPB.GetCSRFTokenRequest, opts ...grpc.CallOption) (*authPB.GetCSRFTokenResponse, error)
}

//go:generate mockgen -destination=../mock/mock_user_client.go -package=mock . UserServiceClient
type UserServiceClient interface {
	GetUser(ctx context.Context, in *userPB.GetUserRequest, opts ...grpc.CallOption) (*userPB.User, error)
}

type AuthDelivery struct {
	SessionDuration time.Duration
	AuthClient      AuthClient
	UserClient      UserServiceClient
}

func NewAuthDelivery(authClient AuthClient, userClient UserServiceClient, sessionDuration time.Duration) *AuthDelivery {
	return &AuthDelivery{
		SessionDuration: sessionDuration,
		AuthClient:      authClient,
		UserClient:      userClient,
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

	authResp, err := d.AuthClient.Login(r.Context(), &authPB.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		log.Warn().Err(err).Str("email", req.Email).Msg("gRPC login failed")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	userResp, err := d.UserClient.GetUser(r.Context(), &userPB.GetUserRequest{UserId: authResp.GetUserId()})
	if err != nil {
		log.Error().Err(err).Msg("failed to retrieve user profile after login")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	expiration := time.Now().Add(d.SessionDuration)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    authResp.GetSessionId(),
		Path:     "/",
		Expires:  expiration,
		HttpOnly: true,
	})

	log.Info().Uint64("user_id", authResp.GetUserId()).Msg("user logged in successfully")

	apiutils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"user": utils.ProtoUserToModel(userResp),
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

	authResp, err := d.AuthClient.Register(r.Context(), &authPB.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		log.Warn().Err(err).Str("email", req.Email).Msg("gRPC registration failed")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	userResp, err := d.UserClient.GetUser(r.Context(), &userPB.GetUserRequest{UserId: authResp.GetUserId()})
	if err != nil {
		log.Error().Err(err).Msg("failed to retrieve user profile after registration")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	expiration := time.Now().Add(d.SessionDuration)
	session := &http.Cookie{
		Name:     "session_id",
		Value:    authResp.GetSessionId(),
		Path:     "/",
		Expires:  expiration,
		HttpOnly: true,
	}
	http.SetCookie(w, session)

	log.Info().Uint64("user_id", authResp.GetUserId()).Msg("user registered successfully")

	apiutils.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"user": utils.ProtoUserToModel(userResp),
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

	_, err = d.AuthClient.Logout(r.Context(), &authPB.LogoutRequest{
		SessionId: session.Value,
	})
	if err != nil {
		log.Error().Err(err).Msg("gRPC logout failed")
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

	grpcResp, err := d.AuthClient.GetCSRFToken(r.Context(), &authPB.GetCSRFTokenRequest{
		SessionId: cookie.Value,
	})
	if err != nil {
		log.Error().Err(err).Msg("gRPC call to GetCSRFToken failed")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, map[string]string{"token": grpcResp.Token})
}
