package delivery

import (
	"backend/gateway_service/internal/apiutils"
	"backend/gateway_service/internal/middleware"
	"backend/gateway_service/internal/notes/models"
	"backend/gateway_service/logger"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (d *NotesDelivery) GetAllNotes(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusInternalServerError, "user not authenticated")
		return
	}

	notes, err := d.usecase.GetAllNotes(r.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get notes")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, notes)
}

func (d *NotesDelivery) CreateNote(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusInternalServerError, "user not authenticated")
		return
	}

	note, err := d.usecase.CreateNote(r.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to create note")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	apiutils.WriteJSON(w, http.StatusCreated, note)
}

func (d *NotesDelivery) GetNoteById(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	vars := mux.Vars(r)
	noteID, err := strconv.ParseUint(vars["note_id"], 10, 64)
	if err != nil {
		log.Warn().Err(err).Str("note_id", vars["note_id"]).Msg("invalid note id")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusInternalServerError, "user not authenticated")
		return
	}

	note, err := d.usecase.GetNoteById(r.Context(), userID, noteID)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("failed to get note")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, note)
}

type UpdateNoteRequest struct {
	Title      *string `json:"title"`
	IsArchived *bool   `json:"is_archived"`
}

func (d *NotesDelivery) UpdateNote(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	vars := mux.Vars(r)
	noteID, err := strconv.ParseUint(vars["note_id"], 10, 64)
	if err != nil {
		log.Warn().Err(err).Str("note_id", vars["note_id"]).Msg("invalid note id")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusInternalServerError, "user not authenticated")
		return
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close request body")
		}
	}()

	var req UpdateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("invalid request body")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == nil && req.IsArchived == nil {
		log.Warn().Msg("invalid request body: title and is_archived are both nil")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := &models.UpdateNoteInput{
		ID:         noteID,
		UserID:     userID,
		Title:      req.Title,
		IsArchived: req.IsArchived,
	}

	note, err := d.usecase.UpdateNote(r.Context(), input)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("failed to update note")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, note)
}

func (d *NotesDelivery) DeleteNote(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	vars := mux.Vars(r)
	noteID, err := strconv.ParseUint(vars["note_id"], 10, 64)
	if err != nil {
		log.Warn().Err(err).Str("note_id", vars["note_id"]).Msg("invalid note id")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusInternalServerError, "user not authenticated")
		return
	}

	err = d.usecase.DeleteNote(r.Context(), userID, noteID)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("failed to delete note")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, "note was successfully deleted")
}

func (d *NotesDelivery) AddFavorite(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	vars := mux.Vars(r)
	noteID, _ := strconv.ParseUint(vars["note_id"], 10, 64)

	err := d.usecase.AddFavorite(r.Context(), userID, noteID)
	if err != nil {
		log.Error().Err(err).Msg("failed to add favorite")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (d *NotesDelivery) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	vars := mux.Vars(r)
	noteID, _ := strconv.ParseUint(vars["note_id"], 10, 64)

	err := d.usecase.RemoveFavorite(r.Context(), userID, noteID)
	if err != nil {
		log.Error().Err(err).Msg("failed to remove favorite")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
