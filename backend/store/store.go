package store

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists             = errors.New("user already exists")
	ErrInvalidEmailOrPassword = errors.New("invalid email or password")
)

type User struct {
	ID        uint64    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type Note struct {
	ID        uint64 `json:"id"`
	OwnerID   uint64 `json:"owner_id"`
	Title     string `json:"title"`
	Text      string `json:"text"`
	Favourite bool   `json:"favorite"`
	Folder    string `json:"folder"`
}

type Store struct {
	mu           sync.RWMutex
	users        map[uint64]*User
	usersByEmail map[string]uint64
	notes        map[uint64]*Note
	sessions     map[string]uint64

	nextUserID uint64
}

func (s *Store) InitFillStore() error {
	_, err := s.CreateUser("user@example.com", "password", "Superuser")
	if err != nil {
		return fmt.Errorf("init fill store: %w", err)
	}

	notes := []*Note{
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
		s.notes[note.ID] = note
	}
	return nil
}

func NewStore() *Store {
	return &Store{
		users:        make(map[uint64]*User),
		usersByEmail: make(map[string]uint64),
		notes:        make(map[uint64]*Note),
		sessions:     make(map[string]uint64),
		nextUserID:   1,
	}
}

func (s *Store) CreateDefaultNotes(userID uint64) {
	notes := []*Note{
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
		s.notes[note.ID] = note
	}
}

func (s *Store) CreateUser(email, password, username string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.usersByEmail[email]; ok {
		return nil, ErrUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("cannot hash password: %w", err)
	}

	user := &User{
		ID:        s.nextUserID,
		Email:     email,
		Username:  username,
		Password:  string(hashedPassword),
		CreatedAt: time.Now().UTC(),
	}
	s.users[user.ID] = user
	s.usersByEmail[email] = user.ID
	s.CreateDefaultNotes(user.ID)
	s.nextUserID++

	return user, nil
}

func (s *Store) AuthenticateUser(email, password string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userID, ok := s.usersByEmail[email]
	if !ok {
		return nil, ErrInvalidEmailOrPassword
	}
	user := s.users[userID]

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidEmailOrPassword
	}

	return user, nil
}

func (s *Store) CreateSession(userID uint64) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionID := uuid.NewString()
	s.sessions[sessionID] = userID

	return sessionID
}

func (s *Store) DeleteSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, sessionID)
}

func (s *Store) GetUserBySession(sessionID string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userID, ok := s.sessions[sessionID]
	if !ok {
		log.Info().Str("session_id", sessionID).Msg("session not found")
		return nil, false
	}
	user, ok := s.users[userID]

	return user, ok
}

func (s *Store) ListNotes(ownerID uint64) []Note {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Note, 0)
	for _, note := range s.notes {
		if note.OwnerID == ownerID {
			result = append(result, *note)
		}
	}

	return result
}
