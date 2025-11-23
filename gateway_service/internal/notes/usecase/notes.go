package usecase

import (
	notePB "backend/notes_service/pkg/note/v1"
	"context"
)

func (u *NotesUsecase) GetAllNotes(ctx context.Context, userID uint64) (*notePB.GetAllNotesResponse, error) {
	return u.repo.GetAllNotes(ctx, userID)
}

func (u *NotesUsecase) CreateNote(ctx context.Context, userID uint64) (*notePB.Note, error) {
	return u.repo.CreateNote(ctx, userID)
}

func (u *NotesUsecase) GetNoteById(ctx context.Context, userID, noteID uint64) (*notePB.Note, error) {
	return u.repo.GetNoteById(ctx, userID, noteID)
}

func (u *NotesUsecase) UpdateNote(ctx context.Context, req *notePB.UpdateNoteRequest) (*notePB.Note, error) {
	return u.repo.UpdateNote(ctx, req)
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
