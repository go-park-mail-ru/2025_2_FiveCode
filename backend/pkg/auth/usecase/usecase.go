package usecase

import (
	"backend/constants"
	"backend/logger"
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
	log := logger.FromContext(ctx)
	log.Info().Str("email", email).Msg("user login attempt")

	userID, err := u.Repository.GetUserIDByEmail(ctx, email)
	if err != nil {
		log.Warn().Str("email", email).Msg("user not found")
		return 0, "", constants.ErrInvalidEmailOrPassword
	}

	userPasswordHash, err := u.Repository.GetUserHashedPasswordByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Uint64("user_id", userID).Msg("failed to get user hashed password")
		return 0, "", fmt.Errorf("failed to get user hashed password: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userPasswordHash), []byte(password)); err != nil {
		log.Warn().Str("email", email).Msg("invalid password")
		return 0, "", constants.ErrInvalidEmailOrPassword
	}

	sessionID, err := u.Repository.CreateSession(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to create session")
		return 0, "", fmt.Errorf("failed to create session: %w", err)
	}

	log.Info().Uint64("user_id", userID).Msg("user logged in successfully")
	return userID, sessionID, nil
}

func (u *AuthUsecase) Register(ctx context.Context, email, password string) (uint64, string, error) {
	log := logger.FromContext(ctx)
	log.Info().Str("email", email).Msg("user registration attempt")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("failed to hash password")
		return 0, "", fmt.Errorf("failed to hash password: %w", err)
	}

	userID, err := u.Repository.CreateUser(ctx, email, string(hashedPassword))
	if err != nil {
		log.Error().Err(err).Str("email", email).Msg("failed to create user")
		return 0, "", err
	}

	sessionID, err := u.Repository.CreateSession(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to create session")
		return 0, "", fmt.Errorf("failed to create session: %w", err)
	}

	log.Info().Uint64("user_id", userID).Msg("user registered successfully")
	return userID, sessionID, nil
}

func (u *AuthUsecase) Logout(ctx context.Context, sessionID string) error {
	log := logger.FromContext(ctx)
	log.Info().Str("session_id", sessionID).Msg("user logout")

	if err := u.Repository.DeleteSession(ctx, sessionID); err != nil {
		log.Error().Err(err).Msg("failed to delete session")
		return fmt.Errorf("failed to logout: %w", err)
	}

	return nil
}

func (u *AuthUsecase) GetUserIDBySession(ctx context.Context, sessionID string) (uint64, error) {
	log := logger.FromContext(ctx)
	log.Info().Str("session_id", sessionID).Msg("getting user ID by session")

	userID, err := u.Repository.GetUserIDBySession(ctx, sessionID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}
