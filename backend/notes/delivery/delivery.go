package delivery

import (
	"backend/apiutils"
	"backend/logger"
	"backend/middleware"
	"backend/models"
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

//go:generate mockgen -source=delivery.go -destination=../mock/mock_delivery.go -package=mock
type NotesUsecase interface {
	GetAllNotes(ctx context.Context, userID uint64) ([]models.Note, error)
	CreateNote(ctx context.Context, userID uint64) (*models.Note, error)
	GetNoteById(ctx context.Context, userID uint64, noteID uint64) (*models.Note, error)
	UpdateNote(ctx context.Context, userID uint64, noteID uint64, title *string, isArchived *bool) (*models.Note, error)
	DeleteNote(ctx context.Context, userID uint64, noteID uint64) error
	AddFavorite(ctx context.Context, userID, noteID uint64) error
	RemoveFavorite(ctx context.Context, userID, noteID uint64) error
}

type NotesDelivery struct {
	Usecase NotesUsecase
}

func NewNotesDelivery(usecase NotesUsecase) *NotesDelivery {
	return &NotesDelivery{
		Usecase: usecase,
	}
}

func (d *NotesDelivery) GetAllNotes(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusInternalServerError, "user not authenticated")
		return
	}

	notes, err := d.Usecase.GetAllNotes(r.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get notes")
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to get notes")
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

	note, err := d.Usecase.CreateNote(r.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to create note")
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to create note")
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

	note, err := d.Usecase.GetNoteById(r.Context(), userID, noteID)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("failed to get note")
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to get note")
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

	note, err := d.Usecase.UpdateNote(r.Context(), userID, noteID, req.Title, req.IsArchived)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("failed to update note")
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to update note")
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

	err = d.Usecase.DeleteNote(r.Context(), userID, noteID)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("failed to delete note")
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to delete note")
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
	if err := d.Usecase.AddFavorite(r.Context(), userID, noteID); err != nil {
		log.Error().Err(err).Msg("failed to add favorite")
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to add favorite")
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
	if err := d.Usecase.RemoveFavorite(r.Context(), userID, noteID); err != nil {
		log.Error().Err(err).Msg("failed to remove favorite")
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to remove favorite")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
