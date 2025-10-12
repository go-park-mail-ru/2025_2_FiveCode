package notesDelivery

import (
	"backend/apiutils"
	"backend/models"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type NotesUsecase interface {
	GetAllNotes(userID uint64) ([]models.Note, error)
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
	vars := mux.Vars(r)
	userID, err := strconv.ParseUint(vars["user_id"], 10, 64)
	if err != nil {
		apiutils.WriteError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	notes, err := d.Usecase.GetAllNotes(userID)

	apiutils.WriteJSON(w, http.StatusOK, notes)
}
