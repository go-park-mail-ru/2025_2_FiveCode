package usecase

import (
	"backend/user_service/internal/constants"
	"backend/user_service/internal/models"
	"backend/user_service/logger"
	"context"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

//go:generate mockgen -source=usecase.go -destination=../mock/mock_usecase.go -package=mock
type UserRepository interface {
	CreateUser(ctx context.Context, email, passwordHash, username string) (uint64, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
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

func (uc *UserUsecase) CreateUser(ctx context.Context, email, password, username string) (*models.User, error) {
	log := logger.FromContext(ctx)

	if username == "" {
		username = strings.Split(email, "@")[0]
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	userID, err := uc.Repository.CreateUser(ctx, email, string(hashedPassword), username)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	log.Info().Uint64("user_id", userID).Msg("user created successfully")

	return uc.Repository.GetUserByID(ctx, userID)
}

func (uc *UserUsecase) VerifyUser(ctx context.Context, email, password string) (*models.User, error) {
	user, err := uc.Repository.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, constants.ErrInvalidEmailOrPassword
	}

	user.Password = ""

	return user, nil
}
