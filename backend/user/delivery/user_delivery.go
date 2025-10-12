package userDelivery

import (
	"backend/apiutils"
	"backend/models"
	"backend/validation"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"net/http"
)

type UserDelivery struct {
	Usecase UserUsecase
}

type UserUsecase interface {
	RegisterUser(email string, password string) (*models.User, error)
	GetUserBySession(session string) (*models.User, error)
}

func NewUserDelivery(u UserUsecase) *UserDelivery {
	return &UserDelivery{
		Usecase: u,
	}
}

type registerRequest struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=8"`
}

func (d *UserDelivery) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		apiutils.WriteError(w, http.StatusBadRequest, "invalid json")
	}

	if err = validation.Validate().Struct(req); err != nil {
		apiutils.WriteValidationError(w, http.StatusBadRequest, err)
		return
	}
	if req.Password != req.ConfirmPassword {
		apiutils.WriteError(w, http.StatusBadRequest, "passwords do not match")
		return
	}

	user, err := d.Usecase.RegisterUser(req.Email, req.Password)
	if err != nil {
		apiutils.WriteError(w, http.StatusInternalServerError, fmt.Sprint("error registering user:", err))
		return
	}

	apiutils.WriteJSON(w, http.StatusCreated, user)
}

func (d *UserDelivery) GetProfile(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if errors.Is(err, http.ErrNoCookie) {
		apiutils.WriteJSON(w, http.StatusOK, nil)
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("error reading session cookie")
		apiutils.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	sessionID := cookie.Value

	user, err := d.Usecase.GetUserBySession(sessionID)
	if err != nil {
		log.Error().Err(err).Msg("error getting user by session")
		apiutils.WriteJSON(w, http.StatusOK, nil)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, user)
}
