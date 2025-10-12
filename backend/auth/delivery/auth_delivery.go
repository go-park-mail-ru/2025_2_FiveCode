package authDelivery

import (
	"backend/apiutils"
	"backend/models"
	"backend/validation"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

type AuthDelivery struct {
	SessionDuration time.Duration
	Usecase         AuthUsecase
}

type AuthUsecase interface {
	Login(email string, password string) (*models.User, string, error)
	Logout(sessionID string) error
}

func NewAuthDelivery(uc AuthUsecase, sessionDuration time.Duration) *AuthDelivery {
	return &AuthDelivery{
		SessionDuration: sessionDuration,
		Usecase:         uc,
	}
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func (d *AuthDelivery) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		apiutils.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if err := validation.Validate().Struct(req); err != nil {
		apiutils.WriteValidationError(w, http.StatusBadRequest, err)
		return
	}

	user, sessionID, err := d.Usecase.Login(req.Email, req.Password)
	if err != nil {
		apiutils.WriteError(w, http.StatusBadRequest, fmt.Sprintf("login failed: %v", err))
		return
	}

	expiration := time.Now().Add(d.SessionDuration)
	session := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  expiration,
		HttpOnly: true,
	}
	http.SetCookie(w, session)

	apiutils.WriteJSON(w, http.StatusOK, user)
}

func (d *AuthDelivery) Logout(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session_id")
	if errors.Is(err, http.ErrNoCookie) {
		log.Info().Msg("no session cookie found")
		apiutils.WriteError(w, http.StatusBadRequest, "no session cookie")
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("error getting session cookie")
		apiutils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	err = d.Usecase.Logout(session.Value)
	if err != nil {
		apiutils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	session.Expires = time.Now().Add(-1 * time.Hour)
	http.SetCookie(w, session)

	apiutils.WriteJSON(w, http.StatusOK, map[string]string{"status": "logged out"})

}
