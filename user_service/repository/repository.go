package repository

import (
	"backend/pkg/store"
	"backend/user_service/internal/constants"
	"backend/user_service/internal/models"
	"backend/user_service/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type UserRepository struct {
	db store.DB
}

func NewUserRepository(db store.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) UpdateUser(ctx context.Context, userID uint64, username *string, password *string, avatarFileID *uint64) (*models.User, error) {
	log := logger.FromContext(ctx)

	setClauses := make([]string, 0)
	args := make([]interface{}, 0)
	argPos := 1

	if username != nil {
		setClauses = append(setClauses, fmt.Sprintf("username = $%d", argPos))
		args = append(args, *username)
		argPos++
	}

	if password != nil {
		setClauses = append(setClauses, fmt.Sprintf("password_hash = $%d", argPos))
		args = append(args, *password)
		argPos++
	}

	if avatarFileID != nil {
		setClauses = append(setClauses, fmt.Sprintf("avatar_file_id = $%d", argPos))
		args = append(args, *avatarFileID)
		argPos++
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argPos))
	args = append(args, time.Now().UTC())
	argPos++

	args = append(args, userID)

	query := fmt.Sprintf(`
		UPDATE "user"
		SET %s
		WHERE id = $%d
		RETURNING id, email, password_hash, username, avatar_file_id, created_at, updated_at
	`, strings.Join(setClauses, ", "), argPos)

	user := &models.User{}
	var avatarFileIDResult sql.NullInt64
	var updatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Username,
		&avatarFileIDResult,
		&user.CreatedAt,
		&updatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		log.Warn().Uint64("user_id", userID).Msg("user not found")
		return nil, constants.ErrNotFound
	}
	if err != nil {
		log.Error().Err(err).Uint64("user_id", userID).Msg("failed to update profile in PostgreSQL")
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	if avatarFileIDResult.Valid {
		val := uint64(avatarFileIDResult.Int64)
		user.AvatarFileID = &val
	}
	if updatedAt.Valid {
		user.UpdatedAt = &updatedAt.Time
	}

	return user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, userID uint64) (*models.User, error) {
	log := logger.FromContext(ctx)

	query := `
		SELECT id, email, password_hash, username, avatar_file_id, created_at, updated_at
		FROM "user"
		WHERE id = $1
	`

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
		log.Warn().Uint64("user_id", userID).Msg("user not found")
		return nil, constants.ErrNotFound
	}
	if err != nil {
		log.Error().Err(err).Uint64("user_id", userID).Msg("failed to get user from PostgreSQL")
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

func (r *UserRepository) DeleteUser(ctx context.Context, userID uint64) error {
	log := logger.FromContext(ctx)

	query := `DELETE FROM "user" WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		log.Error().Err(err).Uint64("user_id", userID).Msg("failed to delete user")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		log.Warn().Uint64("user_id", userID).Msg("user not found for deletion")
		return constants.ErrNotFound
	}

	return nil
}

func (r *UserRepository) CreateUser(ctx context.Context, email, passwordHash, username string) (uint64, error) {
	log := logger.FromContext(ctx)

	now := time.Now().UTC()

	query := `
		INSERT INTO "user" (email, password_hash, username, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var userID uint64
	err := r.db.QueryRowContext(ctx, query, email, passwordHash, username, now, now).Scan(&userID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			log.Warn().Str("email", email).Msg("user already exists")
			return 0, constants.ErrUserExists
		}
		log.Error().Err(err).Msg("failed to create user in PostgreSQL")
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return userID, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	log := logger.FromContext(ctx)

	query := `
		SELECT id, email, password_hash, username, avatar_file_id, created_at, updated_at
		FROM "user"
		WHERE email = $1
	`

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
		return nil, constants.ErrNotFound
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
