package middleware

import (
	"backend/apiutils"
	"backend/store"
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type ctxKey string

const UserIDKey ctxKey = "userID"

func WithUserID(ctx context.Context, id uint64) context.Context {
	return context.WithValue(ctx, UserIDKey, id)
}

func GetUserID(ctx context.Context) (uint64, bool) {
	value := ctx.Value(UserIDKey)
	if value == nil {
		return 0, false
	}
	id, ok := value.(uint64)
	return id, ok
}

var allowed = map[string]bool{
	"http://localhost:8030":      true,
	"http://127.0.0.1:8030":      true,
	"http://89.208.210.115:8030": true,
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func AuthMiddleware(s *store.Store) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := r.Cookie("session_id")
			if errors.Is(err, http.ErrNoCookie) {
				log.Info().Msg("no session cookie found in auth middleware")
				apiutils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "no session cookie"})
				return
			}
			if err != nil {
				log.Error().Err(err).Msg("error getting session cookie in auth middleware")
				apiutils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
				return
			}

			user, ok := s.GetUserBySession(session.Value)
			if !ok {
				apiutils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid session"})
				return
			}

			ctx := WithUserID(r.Context(), user.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserAccessMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := GetUserID(r.Context())
			if !ok {
				apiutils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "user not authenticated"})
				return
			}

			vars := mux.Vars(r)
			idStr := vars["user_id"]
			if idStr == "" {
				apiutils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "missing user id"})
				return
			}

			requestedUserID, err := strconv.ParseUint(idStr, 10, 64)
			if err != nil {
				apiutils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
				return
			}

			if userID != requestedUserID {
				apiutils.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "access denied"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
