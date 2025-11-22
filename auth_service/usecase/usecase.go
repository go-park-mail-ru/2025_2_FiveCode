package usecase

import (
	"backend/auth_service/internal/utils"
	"context"
	"fmt"
)

type AuthUsecase struct {
	Repository  AuthRepository
	UserService UserService
	CSRFSecret  []byte
}

//go:generate mockgen -source=usecase.go -destination=../mock/mock_usecase.go -package=mock
type AuthRepository interface {
	CreateSession(ctx context.Context, userID uint64) (string, error)
	DeleteSession(ctx context.Context, sessionID string) error
	GetUserIDBySession(ctx context.Context, sessionID string) (uint64, error)
}

//go:generate mockgen -source=usecase.go -destination=../mock/mock_usecase.go -package=mock
type UserService interface {
	VerifyUser(ctx context.Context, email, password string) (uint64, error)
	CreateUser(ctx context.Context, email, password string) (uint64, error)
}

func NewAuthUsecase(repository AuthRepository, userService UserService, csrfSecret []byte) *AuthUsecase {
	return &AuthUsecase{
		Repository:  repository,
		UserService: userService,
		CSRFSecret:  csrfSecret,
	}
}

func (u *AuthUsecase) Login(ctx context.Context, email, password string) (uint64, string, error) {
	userID, err := u.UserService.VerifyUser(ctx, email, password)
	if err != nil {
		return 0, "", fmt.Errorf("failed to verify user: %w", err)
	}

	sessionID, err := u.Repository.CreateSession(ctx, userID)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create session: %w", err)
	}

	return userID, sessionID, nil
}

func (u *AuthUsecase) Register(ctx context.Context, email, password string) (uint64, string, error) {
	userID, err := u.UserService.CreateUser(ctx, email, password)
	if err != nil {
		return 0, "", err
	}

	sessionID, err := u.Repository.CreateSession(ctx, userID)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create session: %w", err)
	}

	return userID, sessionID, nil
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
