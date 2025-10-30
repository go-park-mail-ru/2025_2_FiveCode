package Delivery

import (
	"backend/apiutils"
	"backend/middleware"
	"backend/models"
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type NotesUsecase interface {
	GetAllNotes(ctx context.Context, userID uint64) ([]models.Note, error)
	CreateNote(ctx context.Context, userID uint64) (*models.Note, error)
	GetNoteById(ctx context.Context, userID uint64, noteID uint64) (*models.Note, error)
	UpdateNote(ctx context.Context, userID uint64, noteID uint64, title *string, isArchived *bool) (*models.Note, error)
	DeleteNote(ctx context.Context, userID uint64, noteID uint64) error
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
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		apiutils.WriteError(w, http.StatusInternalServerError, "user not authenticated")
		return
	}

	notes, err := d.Usecase.GetAllNotes(r.Context(), userID)
	if err != nil {
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to get notes")
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, notes)
	return
}

func (d *NotesDelivery) CreateNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		apiutils.WriteError(w, http.StatusInternalServerError, "user not authenticated")
		return
	}

	note, err := d.Usecase.CreateNote(r.Context(), userID)
	if err != nil {
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to create note")
		return
	}

	apiutils.WriteJSON(w, http.StatusCreated, note)
	return
}

func (d *NotesDelivery) GetNoteById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteID, err := strconv.ParseUint(vars["note_id"], 10, 64)
	if err != nil {
		apiutils.WriteError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		apiutils.WriteError(w, http.StatusInternalServerError, "user not authenticated")
		return
	}

	note, err := d.Usecase.GetNoteById(r.Context(), userID, noteID)
	if err != nil {
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to get note")
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, note)
	return
}

type UpdateNoteRequest struct {
	Title      *string `json:"title"`
	IsArchived *bool   `json:"is_archived"`
}

func (d *NotesDelivery) UpdateNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteID, err := strconv.ParseUint(vars["note_id"], 10, 64)
	if err != nil {
		apiutils.WriteError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		apiutils.WriteError(w, http.StatusInternalServerError, "user not authenticated")
		return
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			// тут будет лог
		}
	}()
	var req UpdateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiutils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == nil && req.IsArchived == nil {
		apiutils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	note, err := d.Usecase.UpdateNote(r.Context(), userID, noteID, req.Title, req.IsArchived)
	if err != nil {
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to update note")
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, note)
	return
}

func (d *NotesDelivery) DeleteNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteID, err := strconv.ParseUint(vars["note_id"], 10, 64)
	if err != nil {
		apiutils.WriteError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		apiutils.WriteError(w, http.StatusInternalServerError, "user not authenticated")
		return
	}

	err = d.Usecase.DeleteNote(r.Context(), userID, noteID)
	if err != nil {
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to delete note")
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, "note was successfully deleted")
	return
}
