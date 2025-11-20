package usecase

import (
	"backend/pkg/apiutils"
	"backend/pkg/constants"
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	Repository AuthRepository
	CSRFSecret []byte
}

//go:generate mockgen -source=usecase.go -destination=../mock/mock_usecase.go -package=mock
type AuthRepository interface {
	CreateSession(ctx context.Context, userID uint64) (string, error)
	GetUserIDByEmail(ctx context.Context, email string) (uint64, error)
	DeleteSession(ctx context.Context, sessionID string) error
	GetUserIDBySession(ctx context.Context, sessionID string) (uint64, error)
	CreateUser(ctx context.Context, email, passwordHash string) (uint64, error)
	GetUserHashedPasswordByID(ctx context.Context, userID uint64) (string, error)
}

func NewAuthUsecase(repository AuthRepository, csrfSecret []byte) *AuthUsecase {
	return &AuthUsecase{
		Repository: repository,
		CSRFSecret: csrfSecret,
	}
}

func (u *AuthUsecase) Login(ctx context.Context, email, password string) (uint64, string, error) {
	userID, err := u.Repository.GetUserIDByEmail(ctx, email)
	if err != nil {
		return 0, "", constants.ErrInvalidEmailOrPassword
	}

	userPasswordHash, err := u.Repository.GetUserHashedPasswordByID(ctx, userID)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get user hashed password: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userPasswordHash), []byte(password)); err != nil {
		return 0, "", constants.ErrInvalidEmailOrPassword
	}

	sessionID, err := u.Repository.CreateSession(ctx, userID)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create session: %w", err)
	}

	return userID, sessionID, nil
}

func (u *AuthUsecase) Register(ctx context.Context, email, password string) (uint64, string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, "", fmt.Errorf("failed to hash password: %w", err)
	}

	userID, err := u.Repository.CreateUser(ctx, email, string(hashedPassword))
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
	token, err := apiutils.GenerateCSRFToken(sessionID, u.CSRFSecret)
	if err != nil {
		return "", fmt.Errorf("failed to generate csrf token: %w", err)
	}

	return token, nil
}
