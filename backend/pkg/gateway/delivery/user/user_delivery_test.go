package delivery

import (
	"backend/logger"
	"backend/models"
	namederrors "backend/named_errors"
	"backend/user/mock"
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

func TestUserDelivery_GetProfileBySession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockUserUsecase(ctrl)
	delivery := NewUserDelivery(mockUsecase)

	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)

	tests := []struct {
		name           string
		cookie         *http.Cookie
		setupMocks     func()
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:   "success",
			cookie: &http.Cookie{Name: "session_id", Value: "test-session"},
			setupMocks: func() {
				mockUsecase.EXPECT().GetUserBySession(gomock.Any(), "test-session").Return(&models.User{
					ID:        1,
					Email:     "test@example.com",
					Username:  "testuser",
					CreatedAt: time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no cookie",
			cookie:         nil,
			setupMocks:     func() {},
			expectedStatus: http.StatusOK,
			expectedBody:   nil,
		},
		{
			name:   "invalid session",
			cookie: &http.Cookie{Name: "session_id", Value: "invalid"},
			setupMocks: func() {
				mockUsecase.EXPECT().GetUserBySession(gomock.Any(), "invalid").Return(nil, namederrors.ErrInvalidSession)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			req := httptest.NewRequest("GET", "/profile", nil)
			req = req.WithContext(ctx)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			rr := httptest.NewRecorder()

			delivery.GetProfileBySession(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestUserDelivery_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockUserUsecase(ctrl)
	delivery := NewUserDelivery(mockUsecase)

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
			body: updateProfileRequest{
				Username: stringPtr("newusername"),
			},
			setupMocks: func() {
				mockUsecase.EXPECT().UpdateProfile(gomock.Any(), stringPtr("newusername"), nil, nil).Return(&models.User{
					ID:       1,
					Username: "newusername",
				}, nil)
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
			name: "no fields provided",
			body: updateProfileRequest{},
			setupMocks: func() {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid username length",
			body: updateProfileRequest{
				Username: stringPtr("ab"),
			},
			setupMocks: func() {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "user not found",
			body: updateProfileRequest{
				Username: stringPtr("newusername"),
			},
			setupMocks: func() {
				mockUsecase.EXPECT().UpdateProfile(gomock.Any(), stringPtr("newusername"), nil, nil).Return(nil, namederrors.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
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

			req := httptest.NewRequest("PUT", "/profile", bytes.NewBuffer(bodyBytes))
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			delivery.UpdateProfile(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestUserDelivery_GetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockUserUsecase(ctrl)
	delivery := NewUserDelivery(mockUsecase)

	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)

	tests := []struct {
		name           string
		setupMocks     func()
		expectedStatus int
	}{
		{
			name: "success",
			setupMocks: func() {
				mockUsecase.EXPECT().GetProfile(gomock.Any()).Return(&models.User{
					ID:        1,
					Email:     "test@example.com",
					Username:  "testuser",
					CreatedAt: time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "user not found",
			setupMocks: func() {
				mockUsecase.EXPECT().GetProfile(gomock.Any()).Return(nil, namederrors.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "internal error",
			setupMocks: func() {
				mockUsecase.EXPECT().GetProfile(gomock.Any()).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			req := httptest.NewRequest("GET", "/profile", nil)
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			delivery.GetProfile(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestUserDelivery_GetProfileBySession_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockUserUsecase(ctrl)
	delivery := NewUserDelivery(mockUsecase)

	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)

	mockUsecase.EXPECT().GetUserBySession(gomock.Any(), "session").Return(nil, assert.AnError)

	req := httptest.NewRequest("GET", "/profile", nil)
	req = req.WithContext(ctx)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "session"})
	rr := httptest.NewRecorder()

	delivery.GetProfileBySession(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestUserDelivery_UpdateProfile_WithPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockUserUsecase(ctrl)
	delivery := NewUserDelivery(mockUsecase)

	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)

	body := updateProfileRequest{
		Password: stringPtr("newpassword123"),
	}
	bodyBytes, _ := json.Marshal(body)

	mockUsecase.EXPECT().UpdateProfile(gomock.Any(), nil, stringPtr("newpassword123"), nil).Return(&models.User{
		ID:       1,
		Username: "testuser",
	}, nil)

	req := httptest.NewRequest("PUT", "/profile", bytes.NewBuffer(bodyBytes))
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	delivery.UpdateProfile(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestUserDelivery_UpdateProfile_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockUserUsecase(ctrl)
	delivery := NewUserDelivery(mockUsecase)

	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)

	body := updateProfileRequest{
		Username: stringPtr("newusername"),
	}
	bodyBytes, _ := json.Marshal(body)

	mockUsecase.EXPECT().UpdateProfile(gomock.Any(), stringPtr("newusername"), nil, nil).Return(nil, assert.AnError)

	req := httptest.NewRequest("PUT", "/profile", bytes.NewBuffer(bodyBytes))
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	delivery.UpdateProfile(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func stringPtr(s string) *string {
	return &s
}
