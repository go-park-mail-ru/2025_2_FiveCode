package repository

import (
	"backend/logger"
	"backend/middleware"
	"backend/models"
	namederrors "backend/named_errors"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) UpdateProfile(ctx context.Context, username *string, password *string, avatarFileID *uint64) (*models.User, error) {
	log := logger.FromContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		log.Error().Msg("user not authenticated in repository layer")
		return nil, fmt.Errorf("user not authenticated")
	}

	log.Info().Uint64("user_id", userID).Msg("updating user profile in PostgreSQL")

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
		return nil, namederrors.ErrNotFound
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

	log.Info().Uint64("user_id", userID).Msg("profile updated successfully")
	return user, nil
}

func (r *UserRepository) GetProfile(ctx context.Context) (*models.User, error) {
	log := logger.FromContext(ctx)

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		log.Error().Msg("user not authenticated in repository layer")
		return nil, fmt.Errorf("user not authenticated")
	}

	log.Info().Uint64("user_id", userID).Msg("getting user profile from PostgreSQL")

	return r.GetUserByID(ctx, userID)
}

func (r *UserRepository) GetUserByID(ctx context.Context, userID uint64) (*models.User, error) {
	log := logger.FromContext(ctx)
	log.Info().Uint64("user_id", userID).Msg("getting user by id from PostgreSQL")

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
		return nil, namederrors.ErrNotFound
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
	log.Info().Uint64("user_id", userID).Msg("deleting user from PostgreSQL")

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
		return namederrors.ErrNotFound
	}

	log.Info().Uint64("user_id", userID).Msg("user deleted successfully")
	return nil
}
