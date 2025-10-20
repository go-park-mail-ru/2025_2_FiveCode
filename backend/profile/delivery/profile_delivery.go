package profileDelivery

import (
	"backend/apiutils"
	"backend/constants"
	"backend/middleware"
	"backend/models"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

type ProfileDelivery struct {
	Usecase ProfileUsecase
}

type ProfileUsecase interface {
	UpdateProfile(userID uint64, username *string) (*models.User, error)
	GetProfile(userID uint64) (*models.User, error)
	UploadAvatar(file io.Reader, filename, contentType string, size int64) (*models.File, error)
}

func NewProfileDelivery(u ProfileUsecase) *ProfileDelivery {
	return &ProfileDelivery{
		Usecase: u,
	}
}

func (d *ProfileDelivery) UpdateProfile(w http.ResponseWriter, r *http.Request) {
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

	username, avatarFileID, err := d.parseMultipartForm(r)
	if err != nil {
		apiutils.WriteError(w, http.StatusBadRequest, fmt.Sprintf("error parsing multipart form: %v", err))
		return
	}

	if username == nil && avatarFileID == nil {
		apiutils.WriteError(w, http.StatusBadRequest, "at least one field must be provided")
		return
	}

	user, err := d.Usecase.UpdateProfile(userID, username)
	if err != nil {
		log.Error().Err(err).Msg("error updating profile")
		apiutils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("error updating profile: %v", err))
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, user)
}

func (d *ProfileDelivery) GetProfile(w http.ResponseWriter, r *http.Request) {
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

func (d *ProfileDelivery) parseMultipartForm(r *http.Request) (*string, *uint64, error) {
	err := r.ParseMultipartForm(constants.MaxAvatarFileSize)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing multipart form: %w", err)
	}

	var username *string
	var avatarFileID *uint64

	if usernameStr := r.FormValue("username"); usernameStr != "" {
		username = &usernameStr
	}

	file, header, err := r.FormFile("avatar")
	if err == nil {
		defer file.Close()

		fileModel, err := d.Usecase.UploadAvatar(file, header.Filename, header.Header.Get("Content-Type"), header.Size)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to upload file: %w", err)
		}

		avatarFileID = &fileModel.ID
	} else if err != http.ErrMissingFile {
		return nil, nil, fmt.Errorf("error reading avatar file: %w", err)
	}

	return username, avatarFileID, nil
}
