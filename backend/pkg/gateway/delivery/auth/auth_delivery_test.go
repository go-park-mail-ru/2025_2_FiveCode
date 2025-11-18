package delivery

import (
	"backend/auth/mock"
	"backend/logger"
	"backend/models"
	namederrors "backend/named_errors"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestAuthDelivery_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockAuthUsecase(ctrl)
	delivery := NewAuthDelivery(mockUsecase, 24*time.Hour)

	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)

	tests := []struct {
		name           string
		body           interface{}
		setupMocks     func()
		expectedStatus int
	}{
		{
			name: "success",
			body: loginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func() {
				mockUsecase.EXPECT().Login(gomock.Any(), "test@example.com", "password123").Return(&models.User{
					ID:        1,
					Email:     "test@example.com",
					Username:  "test",
					CreatedAt: time.Now(),
				}, "session-id", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid json",
			body: "invalid json",
			setupMocks: func() {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid email or password",
			body: loginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			setupMocks: func() {
				mockUsecase.EXPECT().Login(gomock.Any(), "test@example.com", "wrongpassword").Return(nil, "", namederrors.ErrInvalidEmailOrPassword)
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			var bodyBytes []byte
			var err error
			if str, ok := tt.body.(string); ok {
				bodyBytes = []byte(str)
			} else {
				bodyBytes, err = json.Marshal(tt.body)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(bodyBytes))
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			delivery.Login(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestAuthDelivery_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockAuthUsecase(ctrl)
	delivery := NewAuthDelivery(mockUsecase, 24*time.Hour)

	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)

	tests := []struct {
		name           string
		body           interface{}
		setupMocks     func()
		expectedStatus int
	}{
		{
			name: "success",
			body: registerRequest{
				Email:           "newuser@example.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			setupMocks: func() {
				mockUsecase.EXPECT().Register(gomock.Any(), "newuser@example.com", "password123").Return(&models.User{
					ID:        1,
					Email:     "newuser@example.com",
					Username:  "newuser",
					CreatedAt: time.Now(),
				}, "session-id", nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "passwords do not match",
			body: registerRequest{
				Email:           "newuser@example.com",
				Password:        "password123",
				ConfirmPassword: "different",
			},
			setupMocks: func() {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "user already exists",
			body: registerRequest{
				Email:           "existing@example.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			setupMocks: func() {
				mockUsecase.EXPECT().Register(gomock.Any(), "existing@example.com", "password123").Return(nil, "", namederrors.ErrUserExists)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			var bodyBytes []byte
			var err error
			if str, ok := tt.body.(string); ok {
				bodyBytes = []byte(str)
			} else {
				bodyBytes, err = json.Marshal(tt.body)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(bodyBytes))
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			delivery.Register(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestAuthDelivery_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockAuthUsecase(ctrl)
	delivery := NewAuthDelivery(mockUsecase, 24*time.Hour)

	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)

	tests := []struct {
		name           string
		cookie         *http.Cookie
		setupMocks     func()
		expectedStatus int
	}{
		{
			name:   "success",
			cookie: &http.Cookie{Name: "session_id", Value: "session-id"},
			setupMocks: func() {
				mockUsecase.EXPECT().Logout(gomock.Any(), "session-id").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no cookie",
			cookie:         nil,
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid session",
			cookie: &http.Cookie{Name: "session_id", Value: "invalid"},
			setupMocks: func() {
				mockUsecase.EXPECT().Logout(gomock.Any(), "invalid").Return(namederrors.ErrInvalidSession)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			req := httptest.NewRequest("POST", "/logout", nil)
			req = req.WithContext(ctx)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			rr := httptest.NewRecorder()

			delivery.Logout(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
