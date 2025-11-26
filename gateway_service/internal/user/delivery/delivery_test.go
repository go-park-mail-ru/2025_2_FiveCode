package delivery

import (
	"backend/gateway_service/internal/middleware"
	"backend/gateway_service/internal/user/delivery/mock"
	"backend/gateway_service/internal/user/models"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUserDelivery_GetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockUserUsecase(ctrl)
	delivery := NewUserDelivery(mockUsecase)

	userID := uint64(1)
	user := &models.User{ID: userID, Username: "test"}

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/profile", nil)
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().GetProfile(gomock.Any(), userID).Return(user, nil)

		delivery.GetProfile(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/profile", nil)
		rr := httptest.NewRecorder()

		delivery.GetProfile(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestUserDelivery_GetProfileBySession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockUserUsecase(ctrl)
	delivery := NewUserDelivery(mockUsecase)

	sessionID := "session-123"
	user := &models.User{ID: 1, Username: "test"}

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/me", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().GetProfileBySession(gomock.Any(), sessionID).Return(user, nil)

		delivery.GetProfileBySession(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("NoCookie", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/me", nil)
		rr := httptest.NewRecorder()

		delivery.GetProfileBySession(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestUserDelivery_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockUserUsecase(ctrl)
	delivery := NewUserDelivery(mockUsecase)

	userID := uint64(1)
	username := "newname"
	input := &models.UpdateUserInput{ID: userID, Username: &username}
	updatedUser := &models.User{ID: userID, Username: username}

	t.Run("Success", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"username": username,
		})
		req, _ := http.NewRequest(http.MethodPut, "/profile", bytes.NewBuffer(reqBody))
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().UpdateProfile(gomock.Any(), input).Return(updatedUser, nil)

		delivery.UpdateProfile(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPut, "/profile", nil)
		rr := httptest.NewRecorder()

		delivery.UpdateProfile(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestUserDelivery_DeleteProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockUserUsecase(ctrl)
	delivery := NewUserDelivery(mockUsecase)

	userID := uint64(1)
	sessionID := "session-123"

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, "/profile", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().DeleteProfile(gomock.Any(), userID, sessionID).Return(nil)

		delivery.DeleteProfile(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, "/profile", nil)
		rr := httptest.NewRecorder()

		delivery.DeleteProfile(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}
