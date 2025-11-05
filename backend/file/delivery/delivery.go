package delivery

import (
	"backend/apiutils"
	"backend/logger"
	"backend/models"
	namederrors "backend/named_errors"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

const (
	MaxFileSize = 50 * 1024 * 1024 // 50MB
)

type FileDelivery struct {
	Usecase FileUsecase
}

//go:generate mockgen -source=delivery.go -destination=../mock/mock_delivery.go -package=mock
type FileUsecase interface {
	UploadFile(ctx context.Context, file io.Reader, filename, contentType string, size int64) (*models.File, error)
	GetFile(ctx context.Context, fileID uint64) (*models.File, error)
	DeleteFile(ctx context.Context, fileID uint64) error
}

func NewFileDelivery(usecase FileUsecase) *FileDelivery {
	return &FileDelivery{
		Usecase: usecase,
	}
}

func (d *FileDelivery) UploadFile(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	err := r.ParseMultipartForm(MaxFileSize)
	if err != nil {
		log.Warn().Err(err).Msg("failed to parse multipart form")
		apiutils.WriteError(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		log.Warn().Err(err).Msg("failed to get file from form")
		apiutils.WriteError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close file")
		}
	}()

	if header.Size > MaxFileSize {
		log.Warn().Int64("size", header.Size).Msg("file size exceeds limit")
		apiutils.WriteError(w, http.StatusBadRequest, fmt.Sprintf("file size exceeds maximum allowed size of %d bytes", MaxFileSize))
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	log.Info().
		Str("filename", header.Filename).
		Str("content_type", contentType).
		Int64("size", header.Size).
		Msg("uploading file")

	uploadedFile, err := d.Usecase.UploadFile(r.Context(), file, header.Filename, contentType, header.Size)
	if err != nil {
		log.Error().Err(err).Msg("failed to upload file")
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to upload file")
		return
	}

	log.Info().Uint64("file_id", uploadedFile.ID).Msg("file uploaded successfully")
	apiutils.WriteJSON(w, http.StatusCreated, uploadedFile)
}

func (d *FileDelivery) GetFile(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	vars := mux.Vars(r)
	fileID, err := strconv.ParseUint(vars["file_id"], 10, 64)
	if err != nil {
		log.Warn().Err(err).Str("file_id", vars["file_id"]).Msg("invalid file id")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid file id")
		return
	}

	file, err := d.Usecase.GetFile(r.Context(), fileID)
	if err != nil {
		if errors.Is(err, namederrors.ErrNotFound) {
			log.Warn().Uint64("file_id", fileID).Msg("file not found")
			apiutils.WriteError(w, http.StatusNotFound, "file not found")
			return
		}
		log.Error().Err(err).Uint64("file_id", fileID).Msg("failed to get file")
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to get file")
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, file)
}

func (d *FileDelivery) DeleteFile(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	vars := mux.Vars(r)
	fileID, err := strconv.ParseUint(vars["file_id"], 10, 64)
	if err != nil {
		log.Warn().Err(err).Str("file_id", vars["file_id"]).Msg("invalid file id")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid file id")
		return
	}

	err = d.Usecase.DeleteFile(r.Context(), fileID)
	if err != nil {
		if errors.Is(err, namederrors.ErrNotFound) {
			log.Warn().Uint64("file_id", fileID).Msg("file not found")
			apiutils.WriteError(w, http.StatusNotFound, "file not found")
			return
		}
		log.Error().Err(err).Uint64("file_id", fileID).Msg("failed to delete file")
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to delete file")
		return
	}

	log.Info().Uint64("file_id", fileID).Msg("file deleted successfully")
	w.WriteHeader(http.StatusNoContent)
}
