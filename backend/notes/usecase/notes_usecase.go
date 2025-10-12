package notesUsecase

import (
	"backend/models"
	"github.com/pkg/errors"
)

type NotesUsecase struct {
	Repository NotesRepository
}

type NotesRepository interface {
	GetNotes(userID uint64) ([]models.Note, error)
}

func NewNotesUsecase(Repository NotesRepository) *NotesUsecase {
	return &NotesUsecase{
		Repository: Repository,
	}
}

func (u *NotesUsecase) GetAllNotes(ownerID uint64) ([]models.Note, error) {
	notes, err := u.Repository.GetNotes(ownerID)
	if err != nil {
		return nil, errors.Wrap(err, "could not get notes")
	}
	return notes, nil
}
