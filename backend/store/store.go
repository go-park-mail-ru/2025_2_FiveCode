package store

import (
	"backend/config"
	"backend/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Store struct {
	Minio    *MinioStorage
	Postgres *PostgresDB
	Redis    *RedisDB
}

func (s *Store) InitRedis(conf *config.Config) error {
	rdb, err := NewRedisDB(
		conf.Storages.Redis.Host,
		conf.Storages.Redis.Port,
		conf.Storages.Redis.Password,
		conf.Storages.Redis.DB,
	)
	if err != nil {
		return fmt.Errorf("failed to init redis: %w", err)
	}

	s.Redis = rdb
	return nil
}

func (s *Store) InitPostgres(conf *config.Config) error {
	pg, err := NewPostgresDB(
		conf.Storages.Db.Host,
		conf.Storages.Db.Port,
		conf.Storages.Db.User,
		conf.Storages.Db.Password,
		conf.Storages.Db.DBName,
		conf.Storages.Db.SSLMode,
	)
	if err != nil {
		return fmt.Errorf("failed to init postgres: %w", err)
	}

	s.Postgres = pg
	return nil
}

func (s *Store) InitMinioStorage(conf *config.Config) error {
	minioStorage, err := NewMinioStorage(
		conf.Storages.Minio.Endpoint,
		conf.Storages.Minio.AccessKey,
		conf.Storages.Minio.SecretKey,
		conf.Storages.Minio.Secure,
	)
	if err != nil {
		return fmt.Errorf("error to init Minio storage: %w", err)
	}

	s.Minio = minioStorage
	return nil
}

func (s *Store) InitFillStore(ctx context.Context) error {
	log := logger.FromContext(ctx)
	email := "user@example.com"
	password := "password"

	var userID uint64
	var exists bool
	checkQuery := `SELECT id FROM "user" WHERE email = $1`
	err := s.Postgres.DB.QueryRowContext(ctx, checkQuery, email).Scan(&userID)

	if errors.Is(err, sql.ErrNoRows) {
		log.Info().Str("email", email).Msg("default user not found, creating...")
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		username := strings.Split(email, "@")[0]
		insertQuery := `
            INSERT INTO "user" (email, password_hash, username)
            VALUES ($1, $2, $3)
            RETURNING id
        `
		err = s.Postgres.DB.QueryRowContext(ctx, insertQuery, email, string(hashedPassword), username).Scan(&userID)
		if err != nil {
			return fmt.Errorf("failed to create user in PostgreSQL: %w", err)
		}
		exists = false
		log.Info().Uint64("user_id", userID).Msg("default user created in database")
	} else if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	} else {
		log.Info().Str("email", email).Msg("default user already exists in database")
		exists = true
	}

	if !exists {
		log.Info().Msg("creating default notes for new user")
		notes := []struct {
			Title     string
			IsShared  bool
			CreatedAt time.Time
			UpdatedAt time.Time
		}{
			{"University Lectures", false, time.Now().Add(-30 * 24 * time.Hour), time.Now().Add(-5 * 24 * time.Hour)},
			{"Project Ideas", true, time.Now().Add(-20 * 24 * time.Hour), time.Now().Add(-2 * 24 * time.Hour)},
			{"Shopping List", false, time.Now().Add(-7 * 24 * time.Hour), time.Now().Add(-6 * time.Hour)},
			{"Random Note", false, time.Now().Add(-10 * 24 * time.Hour), time.Now().Add(-8 * 24 * time.Hour)},
		}

		for _, note := range notes {
			insertNoteQuery := `
                INSERT INTO note (owner_id, title, is_archived, is_shared, created_at, updated_at)
                VALUES ($1, $2, $3, $4, $5, $6)
            `
			_, err = s.Postgres.DB.ExecContext(ctx, insertNoteQuery, userID, note.Title, false, note.IsShared, note.CreatedAt, note.UpdatedAt)
			if err != nil {
				return fmt.Errorf("failed to create note '%s': %w", note.Title, err)
			}
		}
		log.Info().Int("count", len(notes)).Msg("default notes created")
	}

	log.Info().Msg("InitFillStore completed successfully")
	return nil
}

func NewStore() *Store {
	return &Store{}
}
