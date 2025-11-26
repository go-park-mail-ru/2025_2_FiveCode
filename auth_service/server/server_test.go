package server

import (
	"backend/auth_service/internal/constants"
	"backend/auth_service/mock"
	pb "backend/auth_service/pkg/auth/v1"
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServer_CreateSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockAuthUsecase(ctrl)
	server := NewServer(mockUsecase)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		userID := uint64(1)
		sessionID := "session_123"

		mockUsecase.EXPECT().
			CreateSession(ctx, userID).
			Return(sessionID, nil)

		resp, err := server.CreateSession(ctx, &pb.CreateSessionRequest{UserId: userID})
		assert.NoError(t, err)
		assert.Equal(t, sessionID, resp.SessionId)
	})

	t.Run("Missing UserID", func(t *testing.T) {
		resp, err := server.CreateSession(ctx, &pb.CreateSessionRequest{UserId: 0})
		assert.Error(t, err)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("Internal Error", func(t *testing.T) {
		userID := uint64(1)
		mockUsecase.EXPECT().
			CreateSession(ctx, userID).
			Return("", errors.New("internal error"))

		resp, err := server.CreateSession(ctx, &pb.CreateSessionRequest{UserId: userID})
		assert.Error(t, err)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Internal, st.Code())
	})
}

func TestServer_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockAuthUsecase(ctrl)
	server := NewServer(mockUsecase)
	ctx := context.Background()
	sessionID := "session_123"

	t.Run("Success", func(t *testing.T) {
		mockUsecase.EXPECT().
			Logout(ctx, sessionID).
			Return(nil)

		resp, err := server.Logout(ctx, &pb.LogoutRequest{SessionId: sessionID})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("Invalid Session", func(t *testing.T) {
		mockUsecase.EXPECT().
			Logout(ctx, sessionID).
			Return(constants.ErrInvalidSession)

		resp, err := server.Logout(ctx, &pb.LogoutRequest{SessionId: sessionID})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("Internal Error", func(t *testing.T) {
		mockUsecase.EXPECT().
			Logout(ctx, sessionID).
			Return(errors.New("internal error"))

		resp, err := server.Logout(ctx, &pb.LogoutRequest{SessionId: sessionID})
		assert.Error(t, err)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Internal, st.Code())
	})
}

func TestServer_GetUserIDBySession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockAuthUsecase(ctrl)
	server := NewServer(mockUsecase)
	ctx := context.Background()
	sessionID := "session_123"

	t.Run("Success", func(t *testing.T) {
		userID := uint64(1)
		mockUsecase.EXPECT().
			GetUserIDBySession(ctx, sessionID).
			Return(userID, nil)

		resp, err := server.GetUserIDBySession(ctx, &pb.GetUserIDBySessionRequest{SessionId: sessionID})
		assert.NoError(t, err)
		assert.Equal(t, userID, resp.UserId)
		assert.True(t, resp.IsValid)
	})

	t.Run("Invalid Session", func(t *testing.T) {
		mockUsecase.EXPECT().
			GetUserIDBySession(ctx, sessionID).
			Return(uint64(0), constants.ErrInvalidSession)

		resp, err := server.GetUserIDBySession(ctx, &pb.GetUserIDBySessionRequest{SessionId: sessionID})
		assert.NoError(t, err)
		assert.False(t, resp.IsValid)
		assert.Equal(t, uint64(0), resp.UserId)
	})

	t.Run("Internal Error", func(t *testing.T) {
		mockUsecase.EXPECT().
			GetUserIDBySession(ctx, sessionID).
			Return(uint64(0), errors.New("internal error"))

		resp, err := server.GetUserIDBySession(ctx, &pb.GetUserIDBySessionRequest{SessionId: sessionID})
		assert.Error(t, err)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Internal, st.Code())
	})
}

func TestServer_GetCSRFToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockAuthUsecase(ctrl)
	server := NewServer(mockUsecase)
	ctx := context.Background()
	sessionID := "session_123"
	token := "csrf_token"

	t.Run("Success", func(t *testing.T) {
		mockUsecase.EXPECT().
			GenerateCSRFToken(ctx, sessionID).
			Return(token, nil)

		resp, err := server.GetCSRFToken(ctx, &pb.GetCSRFTokenRequest{SessionId: sessionID})
		assert.NoError(t, err)
		assert.Equal(t, token, resp.Token)
	})

	t.Run("Internal Error", func(t *testing.T) {
		mockUsecase.EXPECT().
			GenerateCSRFToken(ctx, sessionID).
			Return("", errors.New("internal error"))

		resp, err := server.GetCSRFToken(ctx, &pb.GetCSRFTokenRequest{SessionId: sessionID})
		assert.Error(t, err)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Internal, st.Code())
	})
}

func TestRegisterService(t *testing.T) {
	s := grpc.NewServer()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUsecase := mock.NewMockAuthUsecase(ctrl)
	RegisterService(s, mockUsecase)
}
