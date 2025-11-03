package store

import (
	"backend/config"
	"backend/models"
	namederrors "backend/named_errors"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type Store struct {
	Mu       sync.RWMutex
	Minio    *MinioStorage
	Postgres *PostgresDB
	Redis    *RedisDB

	Users        map[uint64]*models.User
	UsersByEmail map[string]uint64
	Files        map[uint64]*models.File
	sessions     map[string]uint64

	nextUserID uint64
	nextFileID uint64
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
		return errors.New("Error to init Minio storage: " + err.Error())
	}

	s.Minio = minioStorage
	return nil
}

func (s *Store) InitFillStore() error {
	ctx := context.Background()
	email := "user@example.com"
	password := "password"

	var userID uint64
	var exists bool
	checkQuery := `SELECT id FROM "user" WHERE email = $1`
	err := s.Postgres.DB.QueryRowContext(ctx, checkQuery, email).Scan(&userID)

	if errors.Is(err, sql.ErrNoRows) {
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
	} else if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	} else {
		exists = true
	}

	user, err := s.CreateUser(email, password)
	if err != nil && !errors.Is(err, namederrors.ErrUserExists) {
		return fmt.Errorf("failed to create user in memory: %w", err)
	}
	if err == nil {
		s.Mu.Lock()
		delete(s.Users, user.ID)
		delete(s.UsersByEmail, email)

		user.ID = userID
		s.Users[userID] = user
		s.UsersByEmail[email] = userID
		s.Mu.Unlock()

	} else {
	}

	if !exists {
		notes := []struct {
			Title     string
			IsShared  bool
			CreatedAt time.Time
			UpdatedAt time.Time
		}{
			{
				Title:     "University Lectures",
				IsShared:  false,
				CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
				UpdatedAt: time.Now().Add(-5 * 24 * time.Hour),
			},
			{
				Title:     "Project Ideas",
				IsShared:  true,
				CreatedAt: time.Now().Add(-20 * 24 * time.Hour),
				UpdatedAt: time.Now().Add(-2 * 24 * time.Hour),
			},
			{
				Title:     "Shopping List",
				IsShared:  false,
				CreatedAt: time.Now().Add(-7 * 24 * time.Hour),
				UpdatedAt: time.Now().Add(-6 * time.Hour),
			},
			{
				Title:     "Random Note",
				IsShared:  false,
				CreatedAt: time.Now().Add(-10 * 24 * time.Hour),
				UpdatedAt: time.Now().Add(-8 * 24 * time.Hour),
			},
		}

		for _, note := range notes {
			insertNoteQuery := `
                INSERT INTO note (owner_id, title, is_archived, is_shared, created_at, updated_at)
                VALUES ($1, $2, $3, $4, $5, $6)
                RETURNING id
            `
			var noteID uint64
			err = s.Postgres.DB.QueryRowContext(
				ctx,
				insertNoteQuery,
				userID,
				note.Title,
				false,
				note.IsShared,
				note.CreatedAt,
				note.UpdatedAt,
			).Scan(&noteID)

			if err != nil {
				return fmt.Errorf("failed to create note '%s': %w", note.Title, err)
			}

		}
	} else {
	}

	log.Info().Msg("InitFillStore completed successfully")
	return nil
}

func NewStore() *Store {
	return &Store{
		Users:        make(map[uint64]*models.User),
		UsersByEmail: make(map[string]uint64),
		Files:        make(map[uint64]*models.File),
		sessions:     make(map[string]uint64),

		nextUserID: 1,
		nextFileID: 1,
	}
}

func (s *Store) CreateUser(email, password string) (*models.User, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if _, ok := s.UsersByEmail[email]; ok {
		return nil, namederrors.ErrUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("Cannot hash password:" + err.Error())
	}

	user := &models.User{
		ID:        s.nextUserID,
		Email:     email,
		Username:  fmt.Sprintf("user_%d", s.nextUserID),
		Password:  string(hashedPassword),
		CreatedAt: time.Now().UTC(),
	}
	s.Users[user.ID] = user
	s.UsersByEmail[email] = user.ID
	s.nextUserID++

	return user, nil
}

func (s *Store) AuthenticateUser(email, password string) (*models.User, error) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	userID, ok := s.UsersByEmail[email]
	if !ok {
		return nil, namederrors.ErrInvalidEmailOrPassword
	}
	user := s.Users[userID]

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, namederrors.ErrInvalidEmailOrPassword
	}

	return user, nil
}

func (s *Store) CreateSession(userID uint64) string {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	sessionID := uuid.NewString()
	s.sessions[sessionID] = userID

	return sessionID
}

func (s *Store) DeleteSession(sessionID string) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	delete(s.sessions, sessionID)
}

func (s *Store) GetUserBySession(sessionID string) (*models.User, bool) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	userID, ok := s.sessions[sessionID]
	if !ok {
		log.Info().Str("session_id", sessionID).Msg("session not found")
		return nil, false
	}
	user, ok := s.Users[userID]

	return user, ok
}

func (s *Store) GetUserIDBySession(sessionID string) (uint64, bool) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	userID, ok := s.sessions[sessionID]
	if !ok {
		log.Info().Str("session_id", sessionID).Msg("session not found")
		return 0, false
	}

	return userID, true
}

func (s *Store) UpdateUserProfile(userID uint64, username *string, password *string, avatarFileID *uint64) (*models.User, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	user, ok := s.Users[userID]
	if !ok {
		return nil, namederrors.ErrNotFound
	}

	if username != nil {
		user.Username = *username
	}
	if password != nil {
		user.Password = *password
	}
	if avatarFileID != nil {
		user.AvatarFileID = avatarFileID
	}

	now := time.Now().UTC()
	user.UpdatedAt = &now

	return user, nil
}

func (s *Store) GetUserByID(userID uint64) (*models.User, error) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	user, ok := s.Users[userID]
	if !ok {
		return nil, namederrors.ErrNotFound
	}

	return user, nil
}

func (s *Store) SaveFile(file *models.File) (*models.File, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	file.ID = s.nextFileID
	s.Files[file.ID] = file
	s.nextFileID++

	return file, nil
}

func (s *Store) GetFileByID(fileID uint64) (*models.File, error) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	file, ok := s.Files[fileID]
	if !ok {
		return nil, namederrors.ErrNotFound
	}

	return file, nil
}

func (s *Store) UpdateFile(file *models.File) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	_, ok := s.Files[file.ID]
	if !ok {
		return namederrors.ErrNotFound
	}

	s.Files[file.ID] = file
	return nil
}

func (s *Store) DeleteFile(fileID uint64) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	_, ok := s.Files[fileID]
	if !ok {
		return namederrors.ErrNotFound
	}

	delete(s.Files, fileID)
	return nil
}

func (s *Store) UploadFileToMinIO(ctx context.Context, filename string, file io.Reader, size int64, contentType string) (string, error) {
	if s.Minio == nil {
		return "", errors.New("minio storage not initialized")
	}

	objectName := fmt.Sprintf("%s-%s", uuid.New().String(), filename)
	bucketName := "notes-app"

	client := s.Minio.client
	_, err := client.PutObject(ctx, bucketName, objectName, file, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to MinIO: %w", err)
	}

	endpoint := client.EndpointURL()
	scheme := endpoint.Scheme
	url := fmt.Sprintf("%s://%s/%s/%s", scheme, endpoint.Host, bucketName, objectName)
	return url, nil
}
