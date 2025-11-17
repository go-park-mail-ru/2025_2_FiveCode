package usecase

import (
	"backend/logger"
	"backend/middleware"
	"backend/models"
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

//go:generate mockgen -source=usecase.go -destination=../mock/mock_usecase.go -package=mock
type UserRepository interface {
	UpdateProfile(ctx context.Context, username *string, password *string, avatarFileID *uint64) (*models.User, error)
	GetProfile(ctx context.Context) (*models.User, error)
	GetUserByID(ctx context.Context, userID uint64) (*models.User, error)
	DeleteUser(ctx context.Context, userID uint64) error
}

type AuthRepository interface {
	GetUserIDBySession(ctx context.Context, sessionID string) (uint64, error)
	DeleteSession(ctx context.Context, sessionID string) error
}

type UserUsecase struct {
	Repository UserRepository
	AuthRepo   AuthRepository
}

func NewUserUsecase(UserRepository UserRepository, AuthRepo AuthRepository) *UserUsecase {
	return &UserUsecase{
		Repository: UserRepository,
		AuthRepo:   AuthRepo,
	}
}

func (uc *UserUsecase) GetUserBySession(ctx context.Context, sessionID string) (*models.User, error) {
	log := logger.FromContext(ctx)
	log.Info().Msg("getting user by session")

	userID, err := uc.AuthRepo.GetUserIDBySession(ctx, sessionID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get user ID by session from auth repo")
		return nil, fmt.Errorf("failed to get user ID by session: %w", err)
	}

	user, err := uc.Repository.GetUserByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Uint64("user_id", userID).Msg("failed to get user by id")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (uc *UserUsecase) UpdateProfile(ctx context.Context, username *string, password *string, avatarFileID *uint64) (*models.User, error) {
	log := logger.FromContext(ctx)
	log.Info().Msg("updating user profile")

	if password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
		if err != nil {
			log.Error().Err(err).Msg("failed to hash password")
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		passwordStr := string(hashedPassword)
		password = &passwordStr
	}

	user, err := uc.Repository.UpdateProfile(ctx, username, password, avatarFileID)
	if err != nil {
		log.Error().Err(err).Msg("failed to update profile in repository")
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return user, nil
}

func (uc *UserUsecase) GetProfile(ctx context.Context) (*models.User, error) {
	log := logger.FromContext(ctx)
	log.Info().Msg("getting user profile")

	user, err := uc.Repository.GetProfile(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get profile from repository")
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	return user, nil
}

func (uc *UserUsecase) DeleteProfile(ctx context.Context, sessionID string) error {
	log := logger.FromContext(ctx)
	log.Info().Msg("deleting user profile")

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		log.Error().Msg("user not authenticated in usecase layer")
		return fmt.Errorf("user not authenticated")
	}

	err := uc.Repository.DeleteUser(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete user in repository")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if err := uc.AuthRepo.DeleteSession(ctx, sessionID); err != nil {
		log.Error().Err(err).Msg("failed to delete session after user deletion")
	}

	return nil
}
