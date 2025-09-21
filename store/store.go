package store

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists             = errors.New("user already exists")
	ErrInvalidEmailOrPassword = errors.New("invalid email or password")
)

type User struct {
	ID        uint64    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type Note struct {
	ID      uint64 `json:"id"`
	OwnerID uint64 `json:"owner_id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Store struct {
	mu           sync.RWMutex
	users        map[uint64]User
	usersByEmail map[string]uint64
	notes        map[uint64]Note
	sessions     map[string]uint64

	nextUserID uint64
}

func (s *Store) InitFillStore() {
	_, _ = s.CreateUser("user@example.com", "password")

	notes := []Note{
		{
			ID:      1,
			OwnerID: 1,
			Title:   "Заметка 1",
			Content: "Очень важная информация",
		},
		{
			ID:      2,
			OwnerID: 1,
			Title:   "Заметка 2",
			Content: "Не менее важная информация",
		},
	}
	for _, note := range notes {
		s.notes[note.ID] = note
	}
}

func NewStore() *Store {
	return &Store{
		users:        make(map[uint64]User),
		usersByEmail: make(map[string]uint64),
		notes:        make(map[uint64]Note),
		sessions:     make(map[string]uint64),
		nextUserID:   1,
	}
}

func (s *Store) CreateUser(email, password string) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.usersByEmail[email]; ok {
		return User{}, ErrUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("cannot hash password: %w", err)
	}

	user := User{
		ID:        s.nextUserID,
		Email:     email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now().UTC(),
	}
	s.users[user.ID] = user
	s.usersByEmail[email] = user.ID
	s.nextUserID++

	return user, nil
}

func (s *Store) AuthenticateUser(email, password string) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userID, ok := s.usersByEmail[email]
	if !ok {
		return User{}, ErrInvalidEmailOrPassword
	}
	user := s.users[userID]

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return User{}, ErrInvalidEmailOrPassword
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

func (s *Store) GetUserBySession(sessionID string) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userID, ok := s.sessions[sessionID]
	if !ok {
		return User{}, false
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
			result = append(result, note)
		}
	}

	return result
}
