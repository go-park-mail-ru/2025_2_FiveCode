package store

import (
	"backend/models"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	tests := []struct {
		name     string
		preSetup func(t *testing.T, s *Store)
		email    string
		password string
		wantErr  bool
	}{
		{
			name:     "create new user",
			email:    "u1@example.com",
			password: "secret123",
			wantErr:  false,
		},
		{
			name: "duplicate email",
			preSetup: func(t *testing.T, s *Store) {
				_, err := s.CreateUser("dup@example.com", "pass1234")
				require.NoError(t, err, "failed to create user")
			},
			email:    "dup@example.com",
			password: "pass1234",
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := NewStore()
			if test.preSetup != nil {
				test.preSetup(t, s)
			}
			user, err := s.CreateUser(test.email, test.password)

			if test.wantErr {
				require.Error(t, err, "expected error for CreateUser(%s)", test.email)
				return
			}
			require.NoError(t, err, "unexpected CreateUser error")
			require.Equal(t, test.email, user.Email)

			authenticatedUser, err := s.AuthenticateUser(test.email, test.password)
			require.NoError(t, err, "AuthenticateUser failed")
			require.Equal(t, test.email, authenticatedUser.Email)

			_, err = s.AuthenticateUser(test.email, "wrongpw")
			require.Error(t, err, "expected error for wrong password")
		})
	}

	t.Run("Sessions Create/Get/Delete", func(t *testing.T) {
		s := NewStore()
		user, err := s.CreateUser("sess@example.com", "pw123")
		require.NoError(t, err, "CreateUser failed")

		sessionID := s.CreateSession(user.ID)
		require.NotEmpty(t, sessionID, "expected non-empty session id")

		got, ok := s.GetUserBySession(sessionID)
		require.True(t, ok, "GetUserBySession did not find session")
		require.Equal(t, "sess@example.com", got.Email)

		s.DeleteSession(sessionID)
		_, ok = s.GetUserBySession(sessionID)
		require.False(t, ok, "expected session to be deleted")
	})
}

func TestListNotes(t *testing.T) {
	s := NewStore()

	user1, err := s.CreateUser("user1@example.com", "password")
	require.NoError(t, err)

	user1DefaultNotes := s.ListNotes(user1.ID)
	require.Len(t, user1DefaultNotes, 4, "User1 should have 4 default notes")

	user2, err := s.CreateUser("user2@example.com", "password")
	require.NoError(t, err)

	user2DefaultNotes := s.ListNotes(user2.ID)
	require.Len(t, user2DefaultNotes, 4, "User2 should have 4 default notes")

	note1 := &models.Note{ID: 200, OwnerID: user1.ID, Title: "User1 Note 1", Text: "Text 1", Favourite: false, Folder: "Work"}
	note2 := &models.Note{ID: 201, OwnerID: user1.ID, Title: "User1 Note 2", Text: "Text 2", Favourite: true, Folder: "Personal"}
	s.Notes[note1.ID] = note1
	s.Notes[note2.ID] = note2

	note3 := &models.Note{ID: 202, OwnerID: user2.ID, Title: "User2 Note 1", Text: "Text 3", Favourite: false, Folder: "Work"}
	s.Notes[note3.ID] = note3

	user1Notes := s.ListNotes(user1.ID)
	require.Len(t, user1Notes, 6)

	noteTitles := make(map[string]bool)
	for _, note := range user1Notes {
		noteTitles[note.Title] = true
	}
	require.True(t, noteTitles["User1 Note 1"])
	require.True(t, noteTitles["User1 Note 2"])

	user2Notes := s.ListNotes(user2.ID)
	require.Len(t, user2Notes, 5)

	noteTitles = make(map[string]bool)
	for _, note := range user2Notes {
		noteTitles[note.Title] = true
	}
	require.True(t, noteTitles["User2 Note 1"])

	noNotes := s.ListNotes(999)
	require.Len(t, noNotes, 0)
}
