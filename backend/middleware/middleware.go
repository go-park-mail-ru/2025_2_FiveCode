package middleware

import (
	"backend/apiutils"
	"backend/config"
	"backend/logger"
	"backend/store"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
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
	"http://89.208.210.115:8001": true,
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()
		baseLogger := logger.FromContext(r.Context())
		reqLogger := baseLogger.With().Str("request_id", requestID).Logger()
		ctx := logger.ToContext(r.Context(), reqLogger)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterInterceptor) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func realIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	return ip
}

const maxBodyLogSize = 1024

func AccessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log := logger.FromContext(r.Context())

		bodyBytes, _ := io.ReadAll(r.Body)
		r.Body.Close()
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		wInt := &responseWriterInterceptor{ResponseWriter: w, statusCode: http.StatusOK}

		defer func() {
			duration := time.Since(start)
			status := wInt.statusCode

			var logEvent *zerolog.Event
			msg := ""

			switch {
			case status >= 500:
				logEvent = log.Error()
				msg = "Server error"
			case status >= 400:
				logEvent = log.Info()
				msg = "Client error"
			default:
				logEvent = log.Info()
				msg = "Request processed successfully"
			}

			if len(bodyBytes) > 0 && len(bodyBytes) < maxBodyLogSize {
				logEvent = logEvent.Bytes("body", bodyBytes)
			}

			logEvent.
				Str("method", r.Method).
				Str("remote_addr", r.RemoteAddr).
				Str("url", r.URL.Path).
				Dur("work_time", duration).
				Int("status", status).
				Str("user_agent", r.UserAgent()).
				Str("host", r.Host).
				Str("real_ip", realIP(r)).
				Int64("content_length", r.ContentLength).
				Str("start_time", start.Format(time.RFC3339)).
				Str("duration_human", duration.String()).
				Int64("duration_ms", duration.Milliseconds()).
				Msg(msg)
		}()

		next.ServeHTTP(wInt, r)
	})
}

func AuthMiddleware(s *store.Store) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := r.Cookie("session_id")
			if errors.Is(err, http.ErrNoCookie) {
				apiutils.WriteError(w, http.StatusBadRequest, "no session cookie")
				return
			}
			if err != nil {
				apiutils.WriteError(w, http.StatusInternalServerError, "internal server error")
				return
			}

			key := "session:" + session.Value
			val, err := s.Redis.Client.Get(r.Context(), key).Result()
			if errors.Is(err, redis.Nil) {
				apiutils.WriteError(w, http.StatusUnauthorized, "invalid session")
				return
			}
			if err != nil {
				apiutils.WriteError(w, http.StatusInternalServerError, "internal server error")
				return
			}

			var userID uint64
			_, err = fmt.Sscanf(val, "%d", &userID)
			if err != nil {
				apiutils.WriteError(w, http.StatusInternalServerError, "internal server error")
				return
			}

			ctx := WithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserAccessMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := GetUserID(r.Context())
			if !ok {
				apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
				return
			}

			vars := mux.Vars(r)
			idStr := vars["user_id"]
			if idStr == "" {
				apiutils.WriteError(w, http.StatusBadRequest, "missing user id")
				return
			}

			requestedUserID, err := strconv.ParseUint(idStr, 10, 64)
			if err != nil {
				apiutils.WriteError(w, http.StatusBadRequest, "invalid user id")
				return
			}

			if userID != requestedUserID {
				apiutils.WriteError(w, http.StatusForbidden, "access denied")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func CSRFMiddleware(conf *config.Config) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := logger.FromContext(r.Context())

			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			session, err := r.Cookie("session_id")
			if errors.Is(err, http.ErrNoCookie) {
				log.Warn().Msg("csrf check: no session cookie")
				apiutils.WriteErrorWithCode(w, http.StatusUnauthorized, "no session cookie", "session_missing")
				return
			}
			if err != nil {
				log.Error().Err(err).Msg("csrf check: error getting session cookie")
				apiutils.WriteErrorWithCode(w, http.StatusInternalServerError, "internal server error", "internal_error")
				return
			}

			csrfToken := r.Header.Get("X-CSRF-Token")
			if csrfToken == "" {
				log.Warn().Msg("csrf check: missing csrf token in header")
				apiutils.WriteErrorWithCode(w, http.StatusForbidden, "missing csrf token", "csrf_token_missing")
				return
			}

			secretKey := []byte(conf.Auth.CSRF.SecretKey)
			ttlMinutes := conf.Auth.CSRF.TokenTTLMinutes

			err = apiutils.ValidateCSRFToken(csrfToken, session.Value, secretKey, ttlMinutes)
			if err != nil {
				log.Warn().Err(err).Msg("csrf check: invalid csrf token")

				switch {
				case errors.Is(err, apiutils.ErrTokenExpired):
					apiutils.WriteErrorWithCode(w, http.StatusForbidden, "csrf token expired", "csrf_token_expired")
				case errors.Is(err, apiutils.ErrSessionMismatch):
					apiutils.WriteErrorWithCode(w, http.StatusForbidden, "csrf token session mismatch", "csrf_token_invalid")
				default:
					apiutils.WriteErrorWithCode(w, http.StatusForbidden, "invalid csrf token", "csrf_token_invalid")
				}
				return
			}

			log.Debug().Msg("csrf check: token valid")
			next.ServeHTTP(w, r)
		})
	}
}
