package delivery

import (
	"backend/gateway_service/internal/auth/delivery/mock"
	"backend/gateway_service/internal/user/models"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAuthDelivery_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockAuthUsecase(ctrl)
	delivery := NewAuthDelivery(mockUsecase, time.Hour)

	email := "test@example.com"
	password := "password123"
	user := &models.User{ID: 1, Email: email, Username: "test"}
	sessionID := "session-123"

	t.Run("Success", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]string{
			"email":    email,
			"password": password,
		})
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().Login(gomock.Any(), email, password).Return(sessionID, user, nil)

		delivery.Login(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		cookie := rr.Result().Cookies()[0]
		assert.Equal(t, "session_id", cookie.Name)
		assert.Equal(t, sessionID, cookie.Value)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer([]byte("invalid")))
		rr := httptest.NewRecorder()

		delivery.Login(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("ValidationFailed", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]string{
			"email": "invalid-email",
		})
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
		rr := httptest.NewRecorder()

		delivery.Login(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("LoginFailed", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]string{
			"email":    email,
			"password": password,
		})
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBody))
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().Login(gomock.Any(), email, password).Return("", nil, errors.New("failed"))

		delivery.Login(rr, req)

		assert.NotEqual(t, http.StatusOK, rr.Code)
	})
}

func TestAuthDelivery_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockAuthUsecase(ctrl)
	delivery := NewAuthDelivery(mockUsecase, time.Hour)

	email := "test@example.com"
	password := "password123"
	user := &models.User{ID: 1, Email: email, Username: "test"}
	sessionID := "session-123"

	t.Run("Success", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]string{
			"email":            email,
			"password":         password,
			"confirm_password": password,
		})
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(reqBody))
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().Register(gomock.Any(), email, password).Return(sessionID, user, nil)

		delivery.Register(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("PasswordMismatch", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]string{
			"email":            email,
			"password":         password,
			"confirm_password": "other",
		})
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(reqBody))
		rr := httptest.NewRecorder()

		delivery.Register(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestAuthDelivery_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockAuthUsecase(ctrl)
	delivery := NewAuthDelivery(mockUsecase, time.Hour)

	sessionID := "session-123"

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/logout", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().Logout(gomock.Any(), sessionID).Return(nil)

		delivery.Logout(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.True(t, rr.Result().Cookies()[0].Expires.Before(time.Now()))
	})

	t.Run("NoCookie", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/logout", nil)
		rr := httptest.NewRecorder()

		delivery.Logout(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestAuthDelivery_GetCSRFToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockAuthUsecase(ctrl)
	delivery := NewAuthDelivery(mockUsecase, time.Hour)

	sessionID := "session-123"
	token := "csrf-token"

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/csrf", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().GetCSRFToken(gomock.Any(), sessionID).Return(token, nil)

		delivery.GetCSRFToken(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("NoCookie", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/csrf", nil)
		rr := httptest.NewRecorder()

		delivery.GetCSRFToken(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}
