package repository

import (
	authPB "backend/auth_service/pkg/auth/v1"
	"backend/gateway_service/internal/constants"
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AuthClient interface {
	CreateSession(ctx context.Context, in *authPB.CreateSessionRequest, opts ...grpc.CallOption) (*authPB.CreateSessionResponse, error)
	Logout(ctx context.Context, in *authPB.LogoutRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	GetCSRFToken(ctx context.Context, in *authPB.GetCSRFTokenRequest, opts ...grpc.CallOption) (*authPB.GetCSRFTokenResponse, error)
	GetUserIDBySession(ctx context.Context, in *authPB.GetUserIDBySessionRequest, opts ...grpc.CallOption) (*authPB.GetUserIDBySessionResponse, error)
}

type AuthRepository struct {
	client AuthClient
}

func NewAuthRepository(client AuthClient) *AuthRepository {
	return &AuthRepository{
		client: client,
	}
}

func (r *AuthRepository) CreateSession(ctx context.Context, userID uint64) (string, error) {
	resp, err := r.client.CreateSession(ctx, &authPB.CreateSessionRequest{
		UserId: userID,
	})
	if err != nil {
		return "", err
	}
	return resp.SessionId, nil
}

func (r *AuthRepository) Logout(ctx context.Context, sessionID string) error {
	_, err := r.client.Logout(ctx, &authPB.LogoutRequest{
		SessionId: sessionID,
	})
	return err
}

func (r *AuthRepository) GetCSRFToken(ctx context.Context, sessionID string) (string, error) {
	resp, err := r.client.GetCSRFToken(ctx, &authPB.GetCSRFTokenRequest{
		SessionId: sessionID,
	})
	if err != nil {
		return "", err
	}
	return resp.Token, nil
}

func (r *AuthRepository) GetUserIDBySession(ctx context.Context, sessionID string) (uint64, bool, error) {
	resp, err := r.client.GetUserIDBySession(ctx, &authPB.GetUserIDBySessionRequest{
		SessionId: sessionID,
	})
	if err != nil {
		return 0, false, err
	}
	return resp.UserId, resp.IsValid, nil
}

func (r *AuthRepository) ValidateSession(ctx context.Context, sessionID string) (uint64, error) {
	userID, isValid, err := r.GetUserIDBySession(ctx, sessionID)
	if err != nil {
		return 0, err
	}
	if !isValid {
		return 0, constants.ErrInvalidSession
	}
	return userID, nil
}
