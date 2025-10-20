package router

import (
	authDelivery "backend/auth/delivery"
	authRepository "backend/auth/repository"
	authUsecase "backend/auth/usecase"
	"backend/config"
	"backend/initialize"
	notesDelivery "backend/notes/delivery"
	notesRepository "backend/notes/repository"
	notesUsecase "backend/notes/usecase"
	profileDelivery "backend/profile/delivery"
	profileRepository "backend/profile/repository"
	profileUsecase "backend/profile/usecase"
	"backend/store"
	userDelivery "backend/user/delivery"
	userRepository "backend/user/repository"
	userUsecase "backend/user/usecase"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewRouter(t *testing.T) {
	s := store.NewStore()
	conf := &config.Config{
		Cookie: config.CookieConfig{
			SessionDuration: 7,
		},
	}

	deliveries := &initialize.Deliveries{}

	authR := authRepository.NewAuthRepository(s)
	authUC := authUsecase.NewAuthUsecase(authR)
	deliveries.AuthDelivery = authDelivery.NewAuthDelivery(authUC, time.Duration(conf.Cookie.SessionDuration)*24*time.Hour)

	userR := userRepository.NewUserRepository(s)
	userUC := userUsecase.NewUserUsecase(userR)
	deliveries.UserDelivery = userDelivery.NewUserDelivery(userUC)

	notesR := notesRepository.NewNotesRepository(s)
	notesUC := notesUsecase.NewNotesUsecase(notesR)
	deliveries.NotesDelivery = notesDelivery.NewNotesDelivery(notesUC)

	profileR := profileRepository.NewProfileRepository(s)
	profileUC := profileUsecase.NewProfileUsecase(profileR)
	deliveries.ProfileDelivery = profileDelivery.NewProfileDelivery(profileUC)

	router := NewRouter(s, deliveries)
	require.NotNil(t, router, "router should not be nil")

	tests := []struct {
		name     string
		method   string
		path     string
		wantCode int
	}{
		{
			name:   "register endpoint exists",
			method: "POST",
			path:   "/api/register",
		},
		{
			name:   "login endpoint exists",
			method: "POST",
			path:   "/api/login",
		},
		{
			name:   "logout endpoint exists",
			method: "POST",
			path:   "/api/logout",
		},
		{
			name:     "notes endpoint requires auth",
			method:   "GET",
			path:     "/api/user/1/notes",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "non-existent endpoint returns 404",
			method:   "GET",
			path:     "/api/none",
			wantCode: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.path, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if test.wantCode == 0 {
				require.NotEqual(t, http.StatusNotFound, rr.Code, "endpoint should exist")
			} else {
				require.Equal(t, test.wantCode, rr.Code, "should return expected status code")
			}
		})
	}
}
