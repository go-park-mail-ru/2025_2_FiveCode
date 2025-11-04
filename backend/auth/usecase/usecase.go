package Usecase

import (
	"backend/logger"
	"backend/models"
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type AuthRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	CreateSession(ctx context.Context, userID uint64) (string, error)
	DeleteSession(ctx context.Context, sessionID string) error
	GetUserIDBySession(ctx context.Context, sessionID string) (uint64, error)
}

type AuthUsecase struct {
	Repository AuthRepository
}

func NewAuthUsecase(repository AuthRepository) *AuthUsecase {
	return &AuthUsecase{Repository: repository}
}

func (uc *AuthUsecase) Login(ctx context.Context, email string, password string) (*models.User, string, error) {
	log := logger.FromContext(ctx)
	log.Info().Str("email", email).Msg("user login attempt")

	user, err := uc.Repository.GetUserByEmail(ctx, email)
	if err != nil {
		log.Error().Err(err).Str("email", email).Msg("failed to get user by email")
		return nil, "", fmt.Errorf("failed to get user by email: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Warn().Str("email", email).Msg("wrong password provided")
		return nil, "", fmt.Errorf("wrong password: %w", err)
	}

	sessionID, err := uc.Repository.CreateSession(ctx, user.ID)
	if err != nil {
		log.Error().Err(err).Uint64("user_id", user.ID).Msg("failed to create session")
		return nil, "", fmt.Errorf("failed to create session: %w", err)
	}

	return user, sessionID, nil
}

func (uc *AuthUsecase) Logout(ctx context.Context, sessionID string) error {
	log := logger.FromContext(ctx)
	log.Info().Msg("user logout attempt")
	err := uc.Repository.DeleteSession(ctx, sessionID)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete session")
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}
