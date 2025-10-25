package notesDelivery

import (
	"backend/apiutils"
	"backend/middleware"
	"backend/models"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type NotesUsecase interface {
	GetAllNotes(userID uint64) ([]models.Note, error)
	CreateNote(userID uint64) (*models.Note, error)
	GetNoteById(userID uint64, noteID uint64) (*models.Note, error)
	UpdateNote(userID uint64, noteID uint64) (*models.Note, error)
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
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	notes, err := d.Usecase.GetAllNotes(userID)
	if err != nil {
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to get notes")
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, notes)
}

func (d *NotesDelivery) CreateNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	note, err := d.Usecase.CreateNote(userID)
	if err != nil {
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to create note")
	}

	apiutils.WriteJSON(w, http.StatusCreated, note)
}

func (d *NotesDelivery) GetNoteById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteID, err := strconv.ParseUint(vars["note_id"], 10, 64)
	if err != nil {
		apiutils.WriteError(w, http.StatusBadRequest, "invalid note id")
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	note, err := d.Usecase.GetNoteById(userID, noteID)
	if err != nil {
		apiutils.WriteError(w, http.StatusInternalServerError, "failed to get note")
	}

	apiutils.WriteJSON(w, http.StatusOK, note)
}

func (d *NotesDelivery) UpdateNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteID, err := strconv.ParseUint(vars["note_id"], 10, 64)
	if err != nil {
		apiutils.WriteError(w, http.StatusBadRequest, "invalid note id")
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

}
