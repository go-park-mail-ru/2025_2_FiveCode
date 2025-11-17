package delivery

import (
	"backend/file/mock"
	"backend/models"
	namederrors "backend/named_errors"
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestFileDelivery_UploadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockFileUsecase(ctrl)
	delivery := NewFileDelivery(mockUsecase)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("test content"))
	writer.Close()

	mockUsecase.EXPECT().UploadFile(gomock.Any(), gomock.Any(), "test.txt", gomock.Any(), gomock.Any()).Return(&models.File{
		ID:        1,
		URL:       "http://example.com/file.txt",
		MimeType:  "text/plain",
		SizeBytes: 12,
	}, nil)

	req := httptest.NewRequest("POST", "/files", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ctx := context.Background()
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	delivery.UploadFile(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestFileDelivery_GetFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockFileUsecase(ctrl)
	delivery := NewFileDelivery(mockUsecase)

	mockUsecase.EXPECT().GetFile(gomock.Any(), uint64(1)).Return(&models.File{
		ID:        1,
		URL:       "http://example.com/file.txt",
		MimeType:  "text/plain",
		SizeBytes: 10,
	}, nil)

	req := httptest.NewRequest("GET", "/files/1", nil)
	req = mux.SetURLVars(req, map[string]string{"file_id": "1"})
	rr := httptest.NewRecorder()

	delivery.GetFile(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestFileDelivery_GetFile_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockFileUsecase(ctrl)
	delivery := NewFileDelivery(mockUsecase)

	mockUsecase.EXPECT().GetFile(gomock.Any(), uint64(999)).Return(nil, namederrors.ErrNotFound)

	req := httptest.NewRequest("GET", "/files/999", nil)
	req = mux.SetURLVars(req, map[string]string{"file_id": "999"})
	rr := httptest.NewRecorder()

	delivery.GetFile(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestFileDelivery_DeleteFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockFileUsecase(ctrl)
	delivery := NewFileDelivery(mockUsecase)

	mockUsecase.EXPECT().DeleteFile(gomock.Any(), uint64(1)).Return(nil)

	req := httptest.NewRequest("DELETE", "/files/1", nil)
	req = mux.SetURLVars(req, map[string]string{"file_id": "1"})
	rr := httptest.NewRecorder()

	delivery.DeleteFile(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestFileDelivery_DeleteFile_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockFileUsecase(ctrl)
	delivery := NewFileDelivery(mockUsecase)

	mockUsecase.EXPECT().DeleteFile(gomock.Any(), uint64(999)).Return(namederrors.ErrNotFound)

	req := httptest.NewRequest("DELETE", "/files/999", nil)
	req = mux.SetURLVars(req, map[string]string{"file_id": "999"})
	rr := httptest.NewRecorder()

	delivery.DeleteFile(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestFileDelivery_UploadFile_NoFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockFileUsecase(ctrl)
	delivery := NewFileDelivery(mockUsecase)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest("POST", "/files", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ctx := context.Background()
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	delivery.UploadFile(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestFileDelivery_GetFile_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockFileUsecase(ctrl)
	delivery := NewFileDelivery(mockUsecase)

	req := httptest.NewRequest("GET", "/files/invalid", nil)
	req = mux.SetURLVars(req, map[string]string{"file_id": "invalid"})
	rr := httptest.NewRecorder()

	delivery.GetFile(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
