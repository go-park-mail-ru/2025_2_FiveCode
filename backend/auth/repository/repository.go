package repository

import (
	"backend/logger"
	"backend/models"
	namederrors "backend/named_errors"
	"backend/store"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type AuthRepository struct {
	Store *store.Store
}

func NewAuthRepository(store *store.Store) *AuthRepository {
	return &AuthRepository{Store: store}
}

func (r *AuthRepository) CreateSession(ctx context.Context, userID uint64) (string, error) {
	log := logger.FromContext(ctx)
	log.Info().Uint64("user_id", userID).Msg("creating session via redis store")
	return r.Store.Redis.CreateSession(ctx, userID, 30*24*time.Hour)
}

func (r *AuthRepository) DeleteSession(ctx context.Context, sessionID string) error {
	log := logger.FromContext(ctx)
	log.Info().Str("session_id", sessionID).Msg("deleting session via redis store")
	return r.Store.Redis.DeleteSession(ctx, sessionID)
}

func (r *AuthRepository) GetUserIDBySession(ctx context.Context, sessionID string) (uint64, error) {
	log := logger.FromContext(ctx)
	log.Info().Str("session_id", sessionID).Msg("getting user id by session via redis store")
	return r.Store.Redis.GetUserIDBySession(ctx, sessionID)
}

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	log := logger.FromContext(ctx)
	log.Info().Str("email", email).Msg("getting user by email from PostgreSQL")

	query := `SELECT id, email, password_hash, username, avatar_file_id, created_at, updated_at FROM "user" WHERE email = $1`

	user := &models.User{}
	var avatarFileID sql.NullInt64
	var updatedAt sql.NullTime

	err := r.Store.Postgres.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Username,
		&avatarFileID,
		&user.CreatedAt,
		&updatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		log.Warn().Str("email", email).Msg("user not found by email")
		return nil, namederrors.ErrNotFound
	}
	if err != nil {
		log.Error().Err(err).Str("email", email).Msg("failed to query user")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if avatarFileID.Valid {
		val := uint64(avatarFileID.Int64)
		user.AvatarFileID = &val
	}
	if updatedAt.Valid {
		user.UpdatedAt = &updatedAt.Time
	}

	return user, nil
}

func (r *AuthRepository) CreateUser(ctx context.Context, email, passwordHash string) (*models.User, error) {
	log := logger.FromContext(ctx)
	log.Info().Str("email", email).Msg("creating user in PostgreSQL")

	username := strings.Split(email, "@")[0]

	query := `
		INSERT INTO "user" (email, password_hash, username)
		VALUES ($1, $2, $3)
		RETURNING id, email, password_hash, username, avatar_file_id, created_at, updated_at
	`

	user := &models.User{}
	var avatarFileID sql.NullInt64
	var updatedAt sql.NullTime

	err := r.Store.Postgres.DB.QueryRowContext(ctx, query, email, passwordHash, username).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Username,
		&avatarFileID,
		&user.CreatedAt,
		&updatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			log.Warn().Str("email", email).Msg("user already exists")
			return nil, namederrors.ErrUserExists
		}
		log.Error().Err(err).Msg("failed to create user in PostgreSQL")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if avatarFileID.Valid {
		val := uint64(avatarFileID.Int64)
		user.AvatarFileID = &val
	}
	if updatedAt.Valid {
		user.UpdatedAt = &updatedAt.Time
	}

	log.Info().Uint64("user_id", user.ID).Msg("user created in PostgreSQL")
	return user, nil
}

func (r *AuthRepository) GetUserByID(ctx context.Context, userID uint64) (*models.User, error) {
	log := logger.FromContext(ctx)
	log.Info().Uint64("user_id", userID).Msg("getting user by id from PostgreSQL")

	query := `SELECT id, email, password_hash, username, avatar_file_id, created_at, updated_at FROM "user" WHERE id = $1`

	user := &models.User{}
	var avatarFileID sql.NullInt64
	var updatedAt sql.NullTime

	err := r.Store.Postgres.DB.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Username,
		&avatarFileID,
		&user.CreatedAt,
		&updatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		log.Warn().Uint64("user_id", userID).Msg("user not found by id")
		return nil, namederrors.ErrNotFound
	}
	if err != nil {
		log.Error().Err(err).Uint64("user_id", userID).Msg("failed to query user")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if avatarFileID.Valid {
		val := uint64(avatarFileID.Int64)
		user.AvatarFileID = &val
	}
	if updatedAt.Valid {
		user.UpdatedAt = &updatedAt.Time
	}

	return user, nil
}