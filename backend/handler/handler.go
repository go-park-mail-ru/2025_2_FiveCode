package handler

import (
	"backend/store"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type ctxKey string

const UserIDKey ctxKey = "userID"
const sessionDuration = 30 * 24 * time.Hour

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

type Handler struct {
	Store *store.Store
}

func NewHandler(s *store.Store) *Handler {
	return &Handler{Store: s}
}

func WriteJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Error().Err(err).Msg("json encode error")
	}
}

type registerRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
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
	if request.ConfirmPassword == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "no confirm password provided"})
		return
	}
	if request.Password != request.ConfirmPassword {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "passwords do not match"})
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
	expiration := time.Now().Add(sessionDuration)
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
	if errors.Is(err, http.ErrNoCookie) {
		log.Info().Msg("no session cookie found")
		WriteJSON(w, http.StatusNotFound, map[string]string{"error": "no session found"})
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("error getting session cookie")
		WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	h.Store.DeleteSession(session.Value)
	session.Expires = time.Now().Add(-1 * time.Hour)
	http.SetCookie(w, session)

	WriteJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func (h *Handler) ListNotes(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := GetUserID(r.Context())
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "missing user id"})
		return
	}

	requestedUserID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	if currentUserID != requestedUserID {
		WriteJSON(w, http.StatusForbidden, map[string]string{"error": "access denied"})
		return
	}

	notes := h.Store.ListNotes(requestedUserID)
	WriteJSON(w, http.StatusOK, notes)
}
