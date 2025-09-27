package middleware

import (
	"backend/handler"
	"backend/store"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func MakeAuthMiddleware(s *store.Store) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := r.Cookie("session_id")
			if errors.Is(err, http.ErrNoCookie) {
				log.Info().Msg("no session cookie found in auth middleware")
				handler.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "no session found"})
				return
			}
			if err != nil {
				log.Error().Err(err).Msg("error getting session cookie in auth middleware")
				handler.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
				return
			}

			user, ok := s.GetUserBySession(session.Value)
			if !ok {
				handler.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}

			ctx := handler.WithUserID(r.Context(), user.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
