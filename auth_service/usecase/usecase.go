package usecase

import (
	"backend/auth_service/internal/utils"
	"context"
	"fmt"
)

type AuthUsecase struct {
	Repository  AuthRepository
	CSRFSecret  []byte
}

//go:generate mockgen -source=usecase.go -destination=../mock/mock_usecase.go -package=mock
type AuthRepository interface {
	CreateSession(ctx context.Context, userID uint64) (string, error)
	DeleteSession(ctx context.Context, sessionID string) error
	GetUserIDBySession(ctx context.Context, sessionID string) (uint64, error)
}
	
func NewAuthUsecase(repository AuthRepository, csrfSecret []byte) *AuthUsecase {
	return &AuthUsecase{
		Repository:  repository,
		CSRFSecret:  csrfSecret,
	}
}

func (u *AuthUsecase) CreateSession(ctx context.Context, userID uint64) (string, error) {
	sessionID, err := u.Repository.CreateSession(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	return sessionID, nil
}

func (u *AuthUsecase) Logout(ctx context.Context, sessionID string) error {
	if err := u.Repository.DeleteSession(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	return nil
}

func (u *AuthUsecase) GetUserIDBySession(ctx context.Context, sessionID string) (uint64, error) {
	userID, err := u.Repository.GetUserIDBySession(ctx, sessionID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (u *AuthUsecase) GenerateCSRFToken(ctx context.Context, sessionID string) (string, error) {
	token, err := utils.GenerateCSRFToken(sessionID, u.CSRFSecret)
	if err != nil {
		return "", fmt.Errorf("failed to generate csrf token: %w", err)
	}

	return token, nil
}
