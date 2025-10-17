package notesUsecase

import (
	"backend/models"
	"fmt"
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
		return nil, fmt.Errorf("failed to get notes: %w", err)
	}
	return notes, nil
}
