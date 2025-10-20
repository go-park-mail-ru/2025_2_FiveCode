package store

import (
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
	Mu           sync.RWMutex
	Users        map[uint64]*models.User
	UsersByEmail map[string]uint64
	Notes        map[uint64]*models.Note
	Files        map[uint64]*models.File
	sessions     map[string]uint64

	nextUserID uint64
	nextFileID uint64
}

func (s *Store) InitFillStore() error {
	_, err := s.CreateUser("user@example.com", "password")
	if err != nil {
		return fmt.Errorf("init fill store: %w", err)
	}

	notes := []*models.Note{
		{
			ID:        1,
			OwnerID:   1,
			Title:     "University note",
			Text:      "Lecture notes for math and history",
			Favourite: true,
			Folder:    "University",
		},
		{
			ID:        2,
			OwnerID:   1,
			Title:     "Project idea",
			Text:      "Brainstorming app features and sketches",
			Favourite: false,
			Folder:    "University",
		},
		{
			ID:        3,
			OwnerID:   1,
			Title:     "Shopping list",
			Text:      "Milk, bread, eggs, and vegetables",
			Favourite: false,
			Folder:    "Personal",
		},
		{
			ID:        4,
			OwnerID:   1,
			Title:     "Note №4",
			Text:      "Random text of the note",
			Favourite: false,
			Folder:    "Personal",
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
		Files:        make(map[uint64]*models.File),
		sessions:     make(map[string]uint64),
		nextUserID:   1,
		nextFileID:   1,
	}
}

func (s *Store) CreateDefaultNotes(userID uint64) {
	notes := []*models.Note{
		{
			ID:        userID*1000 + 1,
			OwnerID:   userID,
			Title:     "Books to read",
			Text:      "The Three Musketeers, Animal Farm, Angels and Demons",
			Favourite: false,
			Folder:    "Personal",
		},
		{
			ID:        userID*1000 + 2,
			OwnerID:   userID,
			Title:     "Homework",
			Text:      "Write an essay",
			Favourite: false,
			Folder:    "University",
		},
		{
			ID:        userID*1000 + 3,
			OwnerID:   userID,
			Title:     "My wishes",
			Text:      "I want to be a millionaire",
			Favourite: true,
			Folder:    "Personal",
		},
		{
			ID:        userID*1000 + 4,
			OwnerID:   userID,
			Title:     "Films to watch",
			Text:      "Harry Potter, The Lord of the Rings, Avatar",
			Favourite: false,
			Folder:    "Personal",
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
		Username:  fmt.Sprintf("user_%d", s.nextUserID),
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
		if note.OwnerID == ownerID {
			result = append(result, *note)
		}
	}

	return result
}

func (s *Store) UpdateUserProfile(userID uint64, username *string, avatarFileID *uint64) (*models.User, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	user, ok := s.Users[userID]
	if !ok {
		return nil, namederrors.ErrNotFound
	}

	if username != nil {
		user.Username = *username
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
