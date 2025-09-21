package handler

import (
	"backend/store"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

type Handler struct {
	Store *store.Store
}

func NewHandler(s *store.Store) *Handler {
	return &Handler{Store: s}
}

func WriteJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var request registerRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if request.Email == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "no email provided"})
		return
	}
	if request.Password == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "no password provided"})
		return
	}

	user, err := h.Store.CreateUser(request.Email, request.Password)
	if errors.Is(err, store.ErrUserExists) {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "user already exists"})
		return
	}
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("error creating user: %s", err)})
		return
	}

	WriteJSON(w, http.StatusCreated, user)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if request.Email == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "no email provided"})
		return
	}
	if request.Password == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "no password provided"})
		return
	}

	user, err := h.Store.AuthenticateUser(request.Email, request.Password)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": fmt.Sprintf("error authenticating user: %s", err)})
		return
	}

	sessionID := h.Store.CreateSession(user.ID)
	expiration := time.Now().Add(10 * time.Hour)
	session := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  expiration,
		HttpOnly: true,
	}
	http.SetCookie(w, session)

	WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session_id")
	if err == nil {
		h.Store.DeleteSession(session.Value)
		session.Expires = time.Now().Add(-1 * time.Hour)
		http.SetCookie(w, session)
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func (h *Handler) ListNotes(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r.Context())
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	notes := h.Store.ListNotes(userID)

	WriteJSON(w, http.StatusOK, notes)
}

const frontendHost = "localhost"
const frontendPort = "3000"

func NewRouter(s *store.Store) http.Handler {
	h := NewHandler(s)

	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/register", h.Register).Methods("POST")
	api.HandleFunc("/login", h.Login).Methods("POST")
	api.HandleFunc("/logout", h.Logout).Methods("POST")

	protected := api.PathPrefix("/notes").Subrouter()
	protected.Use(MakeAuthMiddleware(s))
	protected.HandleFunc("", h.ListNotes).Methods("GET")

	corsOpts := handlers.AllowedOrigins([]string{fmt.Sprintf("http://%s:%s", frontendHost, frontendPort)})
	corsMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	corsHeaders := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})

	return handlers.CORS(corsOpts, corsMethods, corsHeaders)(r)
}

func MakeAuthMiddleware(s *store.Store) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := r.Cookie("session_id")
			if err != nil {
				WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}

			user, ok := s.GetUserBySession(session.Value)
			if !ok {
				WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}

			ctx := WithUserID(r.Context(), user.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})

	}
}

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
