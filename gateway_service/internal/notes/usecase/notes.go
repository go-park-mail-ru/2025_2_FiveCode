package usecase

import (
	"backend/gateway_service/internal/notes/models"
	"context"
	"errors"
)

func (u *NotesUsecase) GetAllNotes(ctx context.Context, userID uint64) ([]models.Note, error) {
	return u.repo.GetAllNotes(ctx, userID)
}

func (u *NotesUsecase) CreateNote(ctx context.Context, userID uint64, parentNoteID *uint64) (*models.Note, error) {
	if parentNoteID != nil {
		parentNote, err := u.repo.GetNoteById(ctx, userID, *parentNoteID)
		if err != nil {
			return nil, err
		}

		if parentNote.ParentNoteID != nil {
			return nil, errors.New("cannot create sub-note of a sub-note: maximum nesting level is 1")
		}
	}

	return u.repo.CreateNote(ctx, userID, parentNoteID)
}

func (u *NotesUsecase) GetNoteById(ctx context.Context, userID, noteID uint64) (*models.Note, error) {
	return u.repo.GetNoteById(ctx, userID, noteID)
}

func (u *NotesUsecase) UpdateNote(ctx context.Context, input *models.UpdateNoteInput) (*models.Note, error) {
	return u.repo.UpdateNote(ctx, input)
}

func (u *NotesUsecase) DeleteNote(ctx context.Context, userID, noteID uint64) error {
	return u.repo.DeleteNote(ctx, userID, noteID)
}

func (u *NotesUsecase) AddFavorite(ctx context.Context, userID, noteID uint64) error {
	return u.repo.AddFavorite(ctx, userID, noteID)
}

func (u *NotesUsecase) RemoveFavorite(ctx context.Context, userID, noteID uint64) error {
	return u.repo.RemoveFavorite(ctx, userID, noteID)
}

func (u *NotesUsecase) SearchNotes(ctx context.Context, userID uint64, query string) (*models.SearchNotesResponse, error) {
	return u.repo.SearchNotes(ctx, userID, query)
}

func (u *NotesUsecase) SetIcon(ctx context.Context, userID, noteID, iconFileID uint64) (*models.Note, error) {
	return u.repo.SetIcon(ctx, userID, noteID, iconFileID)
}

func (u *NotesUsecase) SetHeader(ctx context.Context, userID, noteID, headerFileID uint64) (*models.Note, error) {
	return u.repo.SetHeader(ctx, userID, noteID, headerFileID)
}
