package delivery

import (
	"backend/gateway_service/internal/apiutils"
	"backend/gateway_service/internal/middleware"
	"backend/gateway_service/internal/notes/models"
	"backend/gateway_service/internal/websocket"
	"backend/pkg/logger"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

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

type CreateNoteRequest struct {
	ParentNoteID *uint64 `json:"parent_note_id,omitempty"`
}

func (d *NotesDelivery) CreateNote(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
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

	var req CreateNoteRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	note, err := d.usecase.CreateNote(r.Context(), userID, req.ParentNoteID)
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

	d.notifyNoteChanged(r.Context(), noteID, userID)

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

func (d *NotesDelivery) notifyNoteChanged(ctx context.Context, noteID uint64, userID uint64) {
	log := logger.FromContext(ctx)

	blocks, err := d.usecase.GetBlocks(ctx, userID, noteID)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("failed to get blocks for ws broadcast")
		return
	}

	note, err := d.usecase.GetNoteById(ctx, userID, noteID)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("failed to get note for ws broadcast")
		return
	}

	message := websocket.ServerMessage{
		Type:      websocket.MessageTypeNoteUpdate,
		NoteID:    int(noteID),
		UpdatedBy: int(userID),
		UpdatedAt: time.Now(),
		Blocks:    blocks,
		Title:     note.Title,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal ws message")
		return
	}

	d.wsHub.BroadcastToNote(int(noteID), data, int(userID))

	log.Debug().
		Uint64("note_id", noteID).
		Uint64("updated_by", userID).
		Msg("note update broadcasted via websocket")
}

type SearchNotesRequest struct {
	Query string `json:"query"`
}

func (d *NotesDelivery) SearchNotes(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

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

	var req SearchNotesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("invalid request body")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	searchResult, err := d.usecase.SearchNotes(r.Context(), userID, req.Query)
	if err != nil {
		log.Error().Err(err).Str("query", req.Query).Msg("failed to search notes")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, searchResult)
}

type SetIconRequest struct {
	IconFileID uint64 `json:"icon_file_id"`
}

func (d *NotesDelivery) SetIcon(w http.ResponseWriter, r *http.Request) {
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

	var req SetIconRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("invalid request body")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	note, err := d.usecase.SetIcon(r.Context(), userID, noteID, req.IconFileID)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("failed to set icon")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, note)
}

type SetHeaderRequest struct {
	HeaderFileID uint64 `json:"header_file_id"`
}

func (d *NotesDelivery) SetHeader(w http.ResponseWriter, r *http.Request) {
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

	var req SetHeaderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("invalid request body")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	note, err := d.usecase.SetHeader(r.Context(), userID, noteID, req.HeaderFileID)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("failed to set header")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, note)
}
