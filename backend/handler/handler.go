package handler

import (
	"backend/apiutils"
	"backend/store"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

const sessionDuration = 30 * 24 * time.Hour

type Handler struct {
	Store *store.Store
}

func NewHandler(s *store.Store) *Handler {
	return &Handler{Store: s}
}

type registerRequest struct {
	Email           string `json:"email"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var request registerRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		apiutils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if request.Email == "" {
		apiutils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "no email provided"})
		return
	}
	if request.Username == "" {
		apiutils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "no username provided"})
		return
	}
	if request.Password == "" {
		apiutils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "no password provided"})
		return
	}
	if request.ConfirmPassword == "" {
		apiutils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "no confirm password provided"})
		return
	}
	if request.Password != request.ConfirmPassword {
		apiutils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "passwords do not match"})
		return
	}

	user, err := h.Store.CreateUser(request.Email, request.Password, request.Username)
	if errors.Is(err, store.ErrUserExists) {
		apiutils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "user already exists"})
		return
	}
	if err != nil {
		apiutils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("error creating user: %s", err)})
		return
	}

	apiutils.WriteJSON(w, http.StatusCreated, user)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		apiutils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if request.Email == "" {
		apiutils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "no email provided"})
		return
	}
	if request.Password == "" {
		apiutils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "no password provided"})
		return
	}

	user, err := h.Store.AuthenticateUser(request.Email, request.Password)
	if err != nil {
		apiutils.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": fmt.Sprintf("error authenticating user: %s", err)})
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

	apiutils.WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session_id")
	if errors.Is(err, http.ErrNoCookie) {
		log.Info().Msg("no session cookie found")
		apiutils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "no session cookie"})
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("error getting session cookie")
		apiutils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	h.Store.DeleteSession(session.Value)
	session.Expires = time.Now().Add(-1 * time.Hour)
	http.SetCookie(w, session)

	apiutils.WriteJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func (h *Handler) ListNotes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.ParseUint(vars["user_id"], 10, 64)
	if err != nil {
		apiutils.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user ID"})
		return
	}
	
	notes := h.Store.ListNotes(userID)
	apiutils.WriteJSON(w, http.StatusOK, notes)
}

func (h *Handler) CheckSession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if errors.Is(err, http.ErrNoCookie) {
		apiutils.WriteJSON(w, http.StatusOK, nil)
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("error reading session cookie")
		apiutils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	sessionID := cookie.Value

	user, ok := h.Store.GetUserBySession(sessionID)
	if !ok {
		log.Error().Err(err).Msg("error loading user for session")
		apiutils.WriteJSON(w, http.StatusOK, nil)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, user)
}
