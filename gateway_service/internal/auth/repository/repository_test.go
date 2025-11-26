package repository

import (
	authPB "backend/auth_service/pkg/auth/v1"
	"backend/gateway_service/internal/auth/repository/mock"
	"backend/gateway_service/internal/constants"
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestAuthRepository_CreateSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockAuthClient(ctrl)
	repo := NewAuthRepository(mockClient)

	userID := uint64(1)
	sessionID := "session-123"

	t.Run("Success", func(t *testing.T) {
		mockClient.EXPECT().CreateSession(gomock.Any(), &authPB.CreateSessionRequest{UserId: userID}).Return(&authPB.CreateSessionResponse{SessionId: sessionID}, nil)

		sid, err := repo.CreateSession(context.Background(), userID)
		assert.NoError(t, err)
		assert.Equal(t, sessionID, sid)
	})

	t.Run("Error", func(t *testing.T) {
		mockClient.EXPECT().CreateSession(gomock.Any(), &authPB.CreateSessionRequest{UserId: userID}).Return(nil, errors.New("failed"))

		_, err := repo.CreateSession(context.Background(), userID)
		assert.Error(t, err)
	})
}

func TestAuthRepository_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockAuthClient(ctrl)
	repo := NewAuthRepository(mockClient)

	sessionID := "session-123"

	t.Run("Success", func(t *testing.T) {
		mockClient.EXPECT().Logout(gomock.Any(), &authPB.LogoutRequest{SessionId: sessionID}).Return(&emptypb.Empty{}, nil)

		err := repo.Logout(context.Background(), sessionID)
		assert.NoError(t, err)
	})
}

func TestAuthRepository_GetCSRFToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockAuthClient(ctrl)
	repo := NewAuthRepository(mockClient)

	sessionID := "session-123"
	token := "csrf-token"

	t.Run("Success", func(t *testing.T) {
		mockClient.EXPECT().GetCSRFToken(gomock.Any(), &authPB.GetCSRFTokenRequest{SessionId: sessionID}).Return(&authPB.GetCSRFTokenResponse{Token: token}, nil)

		tkn, err := repo.GetCSRFToken(context.Background(), sessionID)
		assert.NoError(t, err)
		assert.Equal(t, token, tkn)
	})
}

func TestAuthRepository_GetUserIDBySession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockAuthClient(ctrl)
	repo := NewAuthRepository(mockClient)

	sessionID := "session-123"
	userID := uint64(1)

	t.Run("Success", func(t *testing.T) {
		mockClient.EXPECT().GetUserIDBySession(gomock.Any(), &authPB.GetUserIDBySessionRequest{SessionId: sessionID}).Return(&authPB.GetUserIDBySessionResponse{UserId: userID, IsValid: true}, nil)

		uid, valid, err := repo.GetUserIDBySession(context.Background(), sessionID)
		assert.NoError(t, err)
		assert.True(t, valid)
		assert.Equal(t, userID, uid)
	})
}

func TestAuthRepository_ValidateSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockAuthClient(ctrl)
	repo := NewAuthRepository(mockClient)

	sessionID := "session-123"
	userID := uint64(1)

	t.Run("Success", func(t *testing.T) {
		mockClient.EXPECT().GetUserIDBySession(gomock.Any(), &authPB.GetUserIDBySessionRequest{SessionId: sessionID}).Return(&authPB.GetUserIDBySessionResponse{UserId: userID, IsValid: true}, nil)

		uid, err := repo.ValidateSession(context.Background(), sessionID)
		assert.NoError(t, err)
		assert.Equal(t, userID, uid)
	})

	t.Run("Invalid", func(t *testing.T) {
		mockClient.EXPECT().GetUserIDBySession(gomock.Any(), &authPB.GetUserIDBySessionRequest{SessionId: sessionID}).Return(&authPB.GetUserIDBySessionResponse{UserId: 0, IsValid: false}, nil)

		_, err := repo.ValidateSession(context.Background(), sessionID)
		assert.Error(t, err)
		assert.Equal(t, constants.ErrInvalidSession, err)
	})
}
