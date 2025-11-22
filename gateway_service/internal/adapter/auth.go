package adapter

import (
	"context"

	authPB "backend/auth_service/pkg/auth/v1"
	"backend/gateway_service/internal/constants"
)

type AuthGrpcAdapter struct {
	client authPB.AuthClient
}

func NewAuthGrpcAdapter(client authPB.AuthClient) *AuthGrpcAdapter {
	return &AuthGrpcAdapter{client: client}
}

func (a *AuthGrpcAdapter) ValidateSession(ctx context.Context, sessionID string) (uint64, error) {
	resp, err := a.client.GetUserIDBySession(ctx, &authPB.GetUserIDBySessionRequest{
		SessionId: sessionID,
	})
	if err != nil {
		return 0, err
	}
	if !resp.GetIsValid() {
		return 0, constants.ErrInvalidSession
	}
	return resp.GetUserId(), nil
}
