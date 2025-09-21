package store

import (
	"testing"
)

func TestStore(t *testing.T) {
	t.Run("CreateUser and AuthenticateUser", func(t *testing.T) {
		tests := []struct {
			name     string
			preSetup func(s *Store)
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
				preSetup: func(s *Store) {
					_, _ = s.CreateUser("dup@example.com", "pass1234")
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
					test.preSetup(s)
				}
				user, err := s.CreateUser(test.email, test.password)

				if test.wantErr {
					if err == nil {
						t.Fatalf("expected error for CreateUser(%s)", test.email)
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected CreateUser error: %v", err)
				}
				if user.Email != test.email {
					t.Fatalf("expected email %s, got %s", test.email, user.Email)
				}

				authenticatedUser, err := s.AuthenticateUser(test.email, test.password)
				if err != nil {
					t.Fatalf("AuthenticateUser failed: %v", err)
				}
				if authenticatedUser.Email != test.email {
					t.Fatalf("expected auth email %s, got %s", test.email, authenticatedUser.Email)
				}

				_, err = s.AuthenticateUser(test.email, "wrongpw")
				if err == nil {
					t.Fatalf("expected error for wrong password")
				}
			})
		}
	})

	t.Run("Sessions Create/Get/Delete", func(t *testing.T) {
		s := NewStore()
		user, err := s.CreateUser("sess@example.com", "pw123")
		if err != nil {
			t.Fatalf("CreateUser failed: %v", err)
		}

		sessionID := s.CreateSession(user.ID)
		if sessionID == "" {
			t.Fatalf("expected non-empty session id")
		}

		got, ok := s.GetUserBySession(sessionID)
		if !ok {
			t.Fatalf("GetUserBySession did not find session")
		} else if got.Email != "sess@example.com" {
			t.Fatalf("expected email sess@example.com, got %s", got.Email)
		}

		s.DeleteSession(sessionID)
		_, ok = s.GetUserBySession(sessionID)
		if ok {
			t.Fatalf("expected session to be deleted")
		}
	})
}
