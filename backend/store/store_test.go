package store

import (
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
