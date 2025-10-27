package userDelivery

import (
	"backend/apiutils"
	"backend/constants"
	"backend/middleware"
	"backend/models"
	namederrors "backend/named_errors"
	"backend/validation"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type UserDelivery struct {
	Usecase UserUsecase
}

type UserUsecase interface {
	RegisterUser(email string, password string) (*models.User, error)
	GetUserBySession(session string) (*models.User, error)
	UpdateProfile(userID uint64, username *string, password *string) (*models.User, error)
	GetProfile(userID uint64) (*models.User, error)
	UploadAvatar(file io.Reader, filename, contentType string, size int64) (*models.File, error)
}

func NewUserDelivery(u UserUsecase) *UserDelivery {
	return &UserDelivery{
		Usecase: u,
	}
}

type registerRequest struct {
	Email           string `json:"email" valid:"required,email"`
	Password        string `json:"password" valid:"required,password"`
	ConfirmPassword string `json:"confirm_password" valid:"required,password"`
}

func (d *UserDelivery) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		apiutils.WriteError(w, http.StatusBadRequest, "invalid json")
	}

	if err = validation.ValidateStruct(req); err != nil {
		apiutils.WriteValidationError(w, http.StatusBadRequest, err)
		return
	}
	if req.Password != req.ConfirmPassword {
		apiutils.WriteError(w, http.StatusBadRequest, "passwords do not match")
		return
	}

	user, err := d.Usecase.RegisterUser(req.Email, req.Password)
	if errors.Is(err, namederrors.ErrUserExists) {
		apiutils.WriteError(w, http.StatusBadRequest, "user already exists")
		return
	}
	if err != nil {
		apiutils.WriteError(w, http.StatusInternalServerError, fmt.Sprint("error registering user:", err))
		return
	}

	apiutils.WriteJSON(w, http.StatusCreated, user)
}

func (d *UserDelivery) GetProfileBySession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if errors.Is(err, http.ErrNoCookie) {
		apiutils.WriteJSON(w, http.StatusOK, nil)
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("error reading session cookie")
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to get session cookie")
		return
	}

	sessionID := cookie.Value

	user, err := d.Usecase.GetUserBySession(sessionID)
	if errors.Is(err, namederrors.ErrInvalidSession) {
		apiutils.WriteJSON(w, http.StatusUnauthorized, nil)
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("error getting user by session")
		apiutils.WriteJSON(w, http.StatusInternalServerError, nil)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, user)
}

func (d *UserDelivery) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		apiutils.WriteError(w, http.StatusBadRequest, "content type must be multipart/form-data")
		return
	}

	username, password, avatarFileID, err := d.parseMultipartForm(r)
	if err != nil {
		apiutils.WriteError(w, http.StatusBadRequest, fmt.Sprintf("error parsing multipart form: %v", err))
		return
	}

	if username == nil && password == nil && avatarFileID == nil {
		apiutils.WriteError(w, http.StatusBadRequest, "at least one field must be provided")
		return
	}

	user, err := d.Usecase.UpdateProfile(userID, username, password)
	if err != nil {
		log.Error().Err(err).Msg("error updating profile")
		apiutils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("error updating profile: %v", err))
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, user)
}

func (d *UserDelivery) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	user, err := d.Usecase.GetProfile(userID)
	if err != nil {
		log.Error().Err(err).Msg("error getting profile")
		apiutils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("error getting profile: %v", err))
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, user)
}

func (d *UserDelivery) parseMultipartForm(r *http.Request) (*string, *string, *uint64, error) {
	err := r.ParseMultipartForm(constants.MaxAvatarFileSize)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error parsing multipart form: %w", err)
	}

	var username *string
	var password *string
	var avatarFileID *uint64

	if usernameStr := r.FormValue("username"); usernameStr != "" {
		username = &usernameStr
	}

	if passwordStr := r.FormValue("password"); passwordStr != "" {
		password = &passwordStr
	}

	file, header, err := r.FormFile("avatar")
	if err == nil {
		defer file.Close()

		fileModel, err := d.Usecase.UploadAvatar(file, header.Filename, header.Header.Get("Content-Type"), header.Size)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to upload file: %w", err)
		}

		avatarFileID = &fileModel.ID
	} else if err != http.ErrMissingFile {
		return nil, nil, nil, fmt.Errorf("error reading avatar file: %w", err)
	}

	return username, password, avatarFileID, nil
}
