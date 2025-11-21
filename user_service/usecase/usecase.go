package usecase

import (
	"backend/user_service/logger"
	"backend/user_service/models"
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

//go:generate mockgen -source=usecase.go -destination=../mock/mock_usecase.go -package=mock
type UserRepository interface {
	UpdateUser(ctx context.Context, userID uint64, username *string, password *string, avatarFileID *uint64) (*models.User, error)
	GetUserByID(ctx context.Context, userID uint64) (*models.User, error)
	DeleteUser(ctx context.Context, userID uint64) error
}

type UserUsecase struct {
	Repository UserRepository
}

func NewUserUsecase(UserRepository UserRepository) *UserUsecase {
	return &UserUsecase{
		Repository: UserRepository,
	}
}

func (uc *UserUsecase) UpdateUser(ctx context.Context, userID uint64, username *string, password *string, avatarFileID *uint64) (*models.User, error) {
	log := logger.FromContext(ctx)

	if password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
		if err != nil {
			log.Error().Err(err).Msg("failed to hash password")
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		passwordStr := string(hashedPassword)
		password = &passwordStr
	}

	user, err := uc.Repository.UpdateUser(ctx, userID, username, password, avatarFileID)
	if err != nil {
		log.Error().Err(err).Msg("failed to update profile in repository")
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return user, nil
}

func (uc *UserUsecase) GetUserByID(ctx context.Context, userID uint64) (*models.User, error) {
	log := logger.FromContext(ctx)

	user, err := uc.Repository.GetUserByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get profile from repository")
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	return user, nil
}

func (uc *UserUsecase) DeleteUser(ctx context.Context, userID uint64) error {
	log := logger.FromContext(ctx)

	err := uc.Repository.DeleteUser(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete user in repository")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
