package usecase

import (
	"backend/gateway_service/internal/user/models"
	"context"
)

type AuthRepository interface {
	CreateSession(ctx context.Context, userID uint64) (string, error)
	Logout(ctx context.Context, sessionID string) error
	GetCSRFToken(ctx context.Context, sessionID string) (string, error)
}

type UserRepository interface {
	VerifyUser(ctx context.Context, email, password string) (uint64, error)
	CreateUser(ctx context.Context, email, password, username string) (uint64, error)
	GetUser(ctx context.Context, userID uint64) (*models.User, error)
}

type AuthUsecase struct {
	authRepo AuthRepository
	userRepo UserRepository
}

func NewAuthUsecase(authRepo AuthRepository, userRepo UserRepository) *AuthUsecase {
	return &AuthUsecase{
		authRepo: authRepo,
		userRepo: userRepo,
	}
}

func (u *AuthUsecase) Login(ctx context.Context, email, password string) (string, *models.User, error) {
	userID, err := u.userRepo.VerifyUser(ctx, email, password)
	if err != nil {
		return "", nil, err
	}

	sessionID, err := u.authRepo.CreateSession(ctx, userID)
	if err != nil {
		return "", nil, err
	}

	user, err := u.userRepo.GetUser(ctx, userID)
	if err != nil {
		return "", nil, err
	}

	return sessionID, user, nil
}

func (u *AuthUsecase) Register(ctx context.Context, email, password string) (string, *models.User, error) {
	userID, err := u.userRepo.CreateUser(ctx, email, password, "")
	if err != nil {
		return "", nil, err
	}

	sessionID, err := u.authRepo.CreateSession(ctx, userID)
	if err != nil {
		return "", nil, err
	}

	user, err := u.userRepo.GetUser(ctx, userID)
	if err != nil {
		return "", nil, err
	}

	return sessionID, user, nil
}

func (u *AuthUsecase) Logout(ctx context.Context, sessionID string) error {
	return u.authRepo.Logout(ctx, sessionID)
}

func (u *AuthUsecase) GetCSRFToken(ctx context.Context, sessionID string) (string, error) {
	return u.authRepo.GetCSRFToken(ctx, sessionID)
}
