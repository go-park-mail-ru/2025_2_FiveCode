package repository

import (
	"backend/logger"
	"backend/models"
	namederrors "backend/named_errors"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type AuthRepository struct {
	db    *sql.DB
	redis *redis.Client
}

func NewAuthRepository(db *sql.DB, redisClient *redis.Client) *AuthRepository {
	return &AuthRepository{
		db:    db,
		redis: redisClient,
	}
}

func (r *AuthRepository) CreateSession(ctx context.Context, userID uint64) (string, error) {
	log := logger.FromContext(ctx)
	log.Info().Uint64("user_id", userID).Msg("creating session in redis")

	sessionID := uuid.NewString()
	key := "session:" + sessionID
	duration := 30 * 24 * time.Hour

	err := r.redis.Set(ctx, key, userID, duration).Err()
	if err != nil {
		log.Error().Err(err).Msg("failed to set session in redis")
		return "", err
	}

	return sessionID, nil
}

func (r *AuthRepository) GetUserIDBySession(ctx context.Context, sessionID string) (uint64, error) {
	log := logger.FromContext(ctx)
	log.Info().Str("session_id", sessionID).Msg("getting user id by session from redis")

	key := "session:" + sessionID

	val, err := r.redis.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		log.Warn().Str("key", key).Msg("session not found in redis")
		return 0, namederrors.ErrInvalidSession
	}
	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("failed to get session from redis")
		return 0, err
	}

	var userID uint64
	_, err = fmt.Sscanf(val, "%d", &userID)
	if err != nil {
		log.Error().Err(err).Str("value", val).Msg("failed to parse userID from redis session value")
		return 0, fmt.Errorf("failed to parse userID from session: %w", err)
	}

	return userID, nil
}

func (r *AuthRepository) DeleteSession(ctx context.Context, sessionID string) error {
	log := logger.FromContext(ctx)
	log.Info().Str("session_id", sessionID).Msg("deleting session from redis")

	key := "session:" + sessionID

	err := r.redis.Del(ctx, key).Err()
	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("failed to delete session from redis")
		return err
	}

	return nil
}

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	log := logger.FromContext(ctx)
	log.Info().Str("email", email).Msg("getting user by email from PostgreSQL")

	query := `SELECT id, email, password_hash, username, avatar_file_id, created_at, updated_at FROM "user" WHERE email = $1`

	user := &models.User{}
	var avatarFileID sql.NullInt64
	var updatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, email).Scan(
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

	err := r.db.QueryRowContext(ctx, query, email, passwordHash, username).Scan(
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

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
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
