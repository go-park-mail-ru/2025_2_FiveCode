package delivery

import (
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/gateway_service/internal/constants"
	"backend/gateway_service/internal/file/delivery/mock"
	"backend/gateway_service/internal/file/models"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestFileDelivery_UploadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockFileUsecase(ctrl)
	delivery := NewFileDelivery(mockUsecase)

	t.Run("Success", func(t *testing.T) {
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.txt")
		_, err := part.Write([]byte("test content"))
		assert.NoError(t, err)
		err = writer.Close()
		assert.NoError(t, err)

		req, _ := http.NewRequest(http.MethodPost, "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()

		file := &models.File{ID: 1, URL: "http://test.com/test.txt", SizeBytes: 12}
		mockUsecase.EXPECT().UploadFile(gomock.Any(), gomock.Any(), "test.txt", gomock.Any(), gomock.Any()).Return(file, nil)

		delivery.UploadFile(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("MissingFile", func(t *testing.T) {
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		err := writer.Close()
		assert.NoError(t, err)

		req, _ := http.NewRequest(http.MethodPost, "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()

		delivery.UploadFile(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.txt")
		_, err := part.Write([]byte("test content"))
		assert.NoError(t, err)
		err = writer.Close()
		assert.NoError(t, err)

		req, _ := http.NewRequest(http.MethodPost, "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().UploadFile(gomock.Any(), gomock.Any(), "test.txt", gomock.Any(), gomock.Any()).Return(nil, errors.New("usecase error"))

		delivery.UploadFile(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestFileDelivery_GetFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockFileUsecase(ctrl)
	delivery := NewFileDelivery(mockUsecase)

	fileID := uint64(1)
	file := &models.File{ID: fileID, URL: "http://test.com/test.txt"}

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/files/%d", fileID), nil)
		req = mux.SetURLVars(req, map[string]string{"file_id": fmt.Sprintf("%d", fileID)})
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().GetFile(gomock.Any(), fileID).Return(file, nil)

		delivery.GetFile(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/files/%d", fileID), nil)
		req = mux.SetURLVars(req, map[string]string{"file_id": fmt.Sprintf("%d", fileID)})
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().GetFile(gomock.Any(), fileID).Return(nil, constants.ErrNotFound)

		delivery.GetFile(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("InvalidID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/files/invalid", nil)
		req = mux.SetURLVars(req, map[string]string{"file_id": "invalid"})
		rr := httptest.NewRecorder()

		delivery.GetFile(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/files/%d", fileID), nil)
		req = mux.SetURLVars(req, map[string]string{"file_id": fmt.Sprintf("%d", fileID)})
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().GetFile(gomock.Any(), fileID).Return(nil, errors.New("usecase error"))

		delivery.GetFile(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestFileDelivery_DeleteFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockFileUsecase(ctrl)
	delivery := NewFileDelivery(mockUsecase)

	fileID := uint64(1)

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/files/%d", fileID), nil)
		req = mux.SetURLVars(req, map[string]string{"file_id": fmt.Sprintf("%d", fileID)})
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().DeleteFile(gomock.Any(), fileID).Return(nil)

		delivery.DeleteFile(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("NotFound", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/files/%d", fileID), nil)
		req = mux.SetURLVars(req, map[string]string{"file_id": fmt.Sprintf("%d", fileID)})
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().DeleteFile(gomock.Any(), fileID).Return(constants.ErrNotFound)

		delivery.DeleteFile(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("InvalidID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, "/files/invalid", nil)
		req = mux.SetURLVars(req, map[string]string{"file_id": "invalid"})
		rr := httptest.NewRecorder()

		delivery.DeleteFile(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/files/%d", fileID), nil)
		req = mux.SetURLVars(req, map[string]string{"file_id": fmt.Sprintf("%d", fileID)})
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().DeleteFile(gomock.Any(), fileID).Return(errors.New("usecase error"))

		delivery.DeleteFile(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
