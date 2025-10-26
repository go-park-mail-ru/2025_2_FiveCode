package store

import (
	"backend/config"
	"backend/models"
	namederrors "backend/named_errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type Store struct {
	Minio        *MinioStorage
	Mu           sync.RWMutex
	Users        map[uint64]*models.User
	UsersByEmail map[string]uint64
	Notes        map[uint64]*models.Note
	sessions     map[string]uint64

	nextUserID uint64
	nextNoteID uint64
}

func (s *Store) InitMinioStorage(conf *config.Config) error {
	minioStorage, err := NewMinioStorage(
		conf.Minio.Endpoint,
		conf.Minio.AccessKey,
		conf.Minio.SecretKey,
		conf.Minio.Secure,
	)
	if err != nil {
		return err
	}

	s.Minio = minioStorage
	return nil
}

func (s *Store) InitFillStore() error {
	_, err := s.CreateUser("user@example.com", "password")
	if err != nil {
		return fmt.Errorf("init fill store: %w", err)
	}

	notes := []*models.Note{
		{
			ID:           1,
			OwnerID:      1,
			ParentNoteID: nil,
			Title:        "University Lectures",
			IconFileID:   nil,
			IsArchived:   false,
			IsShared:     false,
			CreatedAt:    time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt:    time.Now().Add(-5 * 24 * time.Hour),
			DeletedAt:    nil,
		},
		{
			ID:           2,
			OwnerID:      1,
			ParentNoteID: nil,
			Title:        "Project Ideas",
			IconFileID:   nil,
			IsArchived:   false,
			IsShared:     true,
			CreatedAt:    time.Now().Add(-20 * 24 * time.Hour),
			UpdatedAt:    time.Now().Add(-2 * 24 * time.Hour),
			DeletedAt:    nil,
		},
		{
			ID:           3,
			OwnerID:      1,
			ParentNoteID: nil,
			Title:        "Shopping List",
			IconFileID:   nil,
			IsArchived:   false,
			IsShared:     false,
			CreatedAt:    time.Now().Add(-7 * 24 * time.Hour),
			UpdatedAt:    time.Now().Add(-6 * time.Hour),
			DeletedAt:    nil,
		},
		{
			ID:           4,
			OwnerID:      1,
			ParentNoteID: nil,
			Title:        "Random Note",
			IconFileID:   nil,
			IsArchived:   false,
			IsShared:     false,
			CreatedAt:    time.Now().Add(-10 * 24 * time.Hour),
			UpdatedAt:    time.Now().Add(-8 * 24 * time.Hour),
			DeletedAt:    nil,
		},
	}
	for _, note := range notes {
		s.Notes[note.ID] = note
	}
	return nil
}

func NewStore() *Store {
	return &Store{
		Users:        make(map[uint64]*models.User),
		UsersByEmail: make(map[string]uint64),
		Notes:        make(map[uint64]*models.Note),
		sessions:     make(map[string]uint64),
		nextUserID:   1,
		nextNoteID:   1,
	}
}

func (s *Store) CreateDefaultNotes(userID uint64) {
	notes := []*models.Note{
		{
			ID:           1,
			OwnerID:      userID*1000 + 1,
			ParentNoteID: nil,
			Title:        "University Lectures",
			IconFileID:   nil,
			IsArchived:   false,
			IsShared:     false,
			CreatedAt:    time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt:    time.Now().Add(-5 * 24 * time.Hour),
			DeletedAt:    nil,
		},
		{
			ID:           2,
			OwnerID:      userID*1000 + 2,
			ParentNoteID: nil,
			Title:        "Project Ideas",
			IconFileID:   nil,
			IsArchived:   false,
			IsShared:     true,
			CreatedAt:    time.Now().Add(-20 * 24 * time.Hour),
			UpdatedAt:    time.Now().Add(-2 * 24 * time.Hour),
			DeletedAt:    nil,
		},
		{
			ID:           3,
			OwnerID:      userID*1000 + 3,
			ParentNoteID: nil,
			Title:        "Shopping List",
			IconFileID:   nil,
			IsArchived:   false,
			IsShared:     false,
			CreatedAt:    time.Now().Add(-7 * 24 * time.Hour),
			UpdatedAt:    time.Now().Add(-6 * time.Hour),
			DeletedAt:    nil,
		},
		{
			ID:           4,
			OwnerID:      userID*1000 + 4,
			ParentNoteID: nil,
			Title:        "Random Note",
			IconFileID:   nil,
			IsArchived:   false,
			IsShared:     false,
			CreatedAt:    time.Now().Add(-10 * 24 * time.Hour),
			UpdatedAt:    time.Now().Add(-8 * 24 * time.Hour),
			DeletedAt:    nil,
		},
	}
	for _, note := range notes {
		s.Notes[note.ID] = note
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
		return nil, fmt.Errorf("cannot hash password: %w", err)
	}

	user := &models.User{
		ID:        s.nextUserID,
		Email:     email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now().UTC(),
	}
	s.Users[user.ID] = user
	s.UsersByEmail[email] = user.ID
	s.CreateDefaultNotes(user.ID)
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

func (s *Store) ListNotes(ownerID uint64) []models.Note {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	result := make([]models.Note, 0)
	for _, note := range s.Notes {
		if note.OwnerID == ownerID && note.DeletedAt != nil {
			result = append(result, *note)
		}
	}

	return result
}

func (s *Store) CreateNote(userID uint64) (*models.Note, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	result := &models.Note{
		ID:        s.nextNoteID,
		OwnerID:   userID,
		Title:     "",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	s.Notes[result.ID] = result
	s.nextNoteID++
	return result, nil
}

func (s *Store) GetNoteById(noteID uint64) (*models.Note, error) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	result := s.Notes[noteID]
	if result == nil {
		return nil, namederrors.ErrNotFound
	}

	return result, nil
}

func (s *Store) DeleteNote(noteID uint64) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	delete(s.Notes, noteID)

	return nil
}

func (s *Store) UpdateNote(noteID uint64, title *string, isArchived *bool) (*models.Note, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if s.Notes[noteID] == nil {
		return nil, namederrors.ErrNotFound
	}

	if title != nil {
		s.Notes[noteID].Title = *title
	}

	if isArchived != nil {
		s.Notes[noteID].IsArchived = *isArchived
	}
	s.Notes[noteID].UpdatedAt = time.Now().UTC()

	return s.Notes[noteID], nil
}
