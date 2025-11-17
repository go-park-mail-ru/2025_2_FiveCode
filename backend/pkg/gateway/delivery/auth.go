package delivery

import (
	"backend/apiutils"
	authPB "backend/gen/go/auth"
	"backend/logger"
	"backend/models"
	"backend/validation"
	"encoding/json"
	"net/http"
	"time"

	pkgErrors "github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthDelivery struct {
	SessionDuration time.Duration
	AuthClient      authPB.AuthClient
}

func NewAuthDelivery(authClient authPB.AuthClient, sessionDuration time.Duration) *AuthDelivery {
	return &AuthDelivery{
		SessionDuration: sessionDuration,
		AuthClient:      authClient,
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

	grpcResp, err := d.AuthClient.Login(r.Context(), &authPB.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		log.Warn().Err(err).Str("email", req.Email).Msg("gRPC login failed")
		st, ok := status.FromError(err)
		if !ok {
			apiutils.WriteError(w, http.StatusInternalServerError, "login failed")
			return
		}
		if st.Code() == codes.Unauthenticated {
			apiutils.WriteError(w, http.StatusUnauthorized, "invalid email or password")
		} else {
			apiutils.WriteError(w, http.StatusInternalServerError, "login failed")
		}
		return
	}

	expiration := time.Now().Add(d.SessionDuration)
	session := &http.Cookie{
		Name:     "session_id",
		Value:    grpcResp.GetSessionId(),
		Path:     "/",
		Expires:  expiration,
		HttpOnly: true,
	}
	http.SetCookie(w, session)

	log.Info().Str("session_id", grpcResp.GetSessionId()).Msg("user logged in successfully")
	apiutils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"user": models.User{
			ID:    grpcResp.GetUserId(),
			Email: req.Email,
		},
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

	grpcResp, err := d.AuthClient.Register(r.Context(), &authPB.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		log.Warn().Err(err).Str("email", req.Email).Msg("gRPC registration failed")
		st, ok := status.FromError(err)
		if !ok {
			apiutils.WriteError(w, http.StatusInternalServerError, "registration failed")
			return
		}
		if st.Code() == codes.AlreadyExists {
			apiutils.WriteError(w, http.StatusConflict, "email already registered")
		} else {
			apiutils.WriteError(w, http.StatusInternalServerError, "registration failed")
		}
		return
	}

	expiration := time.Now().Add(d.SessionDuration)
	session := &http.Cookie{
		Name:     "session_id",
		Value:    grpcResp.GetSessionId(),
		Path:     "/",
		Expires:  expiration,
		HttpOnly: true,
	}
	http.SetCookie(w, session)

	log.Info().Str("session_id", grpcResp.GetSessionId()).Msg("user registered successfully")
	apiutils.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"user": models.User{
			ID:    grpcResp.GetUserId(),
			Email: req.Email,
		},
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
		log.Error().Err(err).Str("session_id", session.Value).Msg("gRPC logout failed")
		apiutils.WriteError(w, http.StatusInternalServerError, "logout failed")
		return
	}

	session.Expires = time.Now().Add(-1 * time.Hour)
	http.SetCookie(w, session)

	log.Info().Msg("user logged out successfully")
	apiutils.WriteJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}
