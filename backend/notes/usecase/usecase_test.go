package usecase

import (
	"backend/logger"
	"backend/models"
	namederrors "backend/named_errors"
	"backend/notes/mock"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNotesUsecase_GetAllNotes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewNotesUsecase(mockRepo)

	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)

	tests := []struct {
		name          string
		userID        uint64
		setupMocks    func()
		expectedNotes []models.Note
		expectedError error
	}{
		{
			name:   "success",
			userID: 1,
			setupMocks: func() {
				mockRepo.EXPECT().GetNotes(ctx, uint64(1)).Return([]models.Note{
					{ID: 1, OwnerID: 1, Title: "Note 1"},
					{ID: 2, OwnerID: 1, Title: "Note 2"},
				}, nil)
			},
			expectedNotes: []models.Note{
				{ID: 1, OwnerID: 1, Title: "Note 1"},
				{ID: 2, OwnerID: 1, Title: "Note 2"},
			},
			expectedError: nil,
		},
		{
			name:   "repository error",
			userID: 1,
			setupMocks: func() {
				mockRepo.EXPECT().GetNotes(ctx, uint64(1)).Return(nil, errors.New("database error"))
			},
			expectedNotes: nil,
			expectedError: errors.New("failed to get notes"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			notes, err := usecase.GetAllNotes(ctx, tt.userID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, notes)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expectedNotes), len(notes))
			}
		})
	}
}

func TestNotesUsecase_CreateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewNotesUsecase(mockRepo)

	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)

	tests := []struct {
		name          string
		userID        uint64
		setupMocks    func()
		expectedNote  *models.Note
		expectedError error
	}{
		{
			name:   "success",
			userID: 1,
			setupMocks: func() {
				mockRepo.EXPECT().CreateNote(ctx, uint64(1)).Return(&models.Note{
					ID:        1,
					OwnerID:   1,
					Title:     "New Note",
					CreatedAt: time.Now(),
				}, nil)
			},
			expectedNote: &models.Note{
				ID:      1,
				OwnerID: 1,
				Title:   "New Note",
			},
			expectedError: nil,
		},
		{
			name:   "repository error",
			userID: 1,
			setupMocks: func() {
				mockRepo.EXPECT().CreateNote(ctx, uint64(1)).Return(nil, errors.New("database error"))
			},
			expectedNote:  nil,
			expectedError: errors.New("failed to create note"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			note, err := usecase.CreateNote(ctx, tt.userID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, note)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedNote.ID, note.ID)
				assert.Equal(t, tt.expectedNote.OwnerID, note.OwnerID)
			}
		})
	}
}

func TestNotesUsecase_GetNoteById(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewNotesUsecase(mockRepo)

	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)

	tests := []struct {
		name          string
		userID        uint64
		noteID        uint64
		setupMocks    func()
		expectedNote  *models.Note
		expectedError error
	}{
		{
			name:   "success",
			userID: 1,
			noteID: 1,
			setupMocks: func() {
				mockRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(&models.Note{
					ID:        1,
					OwnerID:   1,
					Title:     "Test Note",
					CreatedAt: time.Now(),
				}, nil)
			},
			expectedNote: &models.Note{
				ID:      1,
				OwnerID: 1,
				Title:   "Test Note",
			},
			expectedError: nil,
		},
		{
			name:   "no access",
			userID: 1,
			noteID: 1,
			setupMocks: func() {
				mockRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(&models.Note{
					ID:        1,
					OwnerID:   2,
					Title:     "Test Note",
					CreatedAt: time.Now(),
				}, nil)
			},
			expectedNote:  nil,
			expectedError: namederrors.ErrNoAccess,
		},
		{
			name:   "repository error",
			userID: 1,
			noteID: 1,
			setupMocks: func() {
				mockRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(nil, errors.New("database error"))
			},
			expectedNote:  nil,
			expectedError: errors.New("failed to get note"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			note, err := usecase.GetNoteById(ctx, tt.userID, tt.noteID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				if tt.expectedError == namederrors.ErrNoAccess {
					assert.Equal(t, namederrors.ErrNoAccess, err)
				} else {
					assert.Contains(t, err.Error(), tt.expectedError.Error())
				}
				assert.Nil(t, note)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedNote.ID, note.ID)
			}
		})
	}
}

func TestNotesUsecase_UpdateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewNotesUsecase(mockRepo)

	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)

	tests := []struct {
		name          string
		userID        uint64
		noteID        uint64
		title         *string
		isArchived    *bool
		setupMocks    func()
		expectedNote  *models.Note
		expectedError error
	}{
		{
			name:       "success",
			userID:     1,
			noteID:     1,
			title:      stringPtr("Updated Title"),
			isArchived: nil,
			setupMocks: func() {
				mockRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(&models.Note{
					ID:        1,
					OwnerID:   1,
					Title:     "Old Title",
					CreatedAt: time.Now(),
				}, nil)
				mockRepo.EXPECT().UpdateNote(ctx, uint64(1), stringPtr("Updated Title"), nil).Return(&models.Note{
					ID:        1,
					OwnerID:   1,
					Title:     "Updated Title",
					CreatedAt: time.Now(),
				}, nil)
			},
			expectedNote: &models.Note{
				ID:      1,
				OwnerID: 1,
				Title:   "Updated Title",
			},
			expectedError: nil,
		},
		{
			name:       "no access",
			userID:     1,
			noteID:     1,
			title:      stringPtr("Updated Title"),
			isArchived: nil,
			setupMocks: func() {
				mockRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(&models.Note{
					ID:        1,
					OwnerID:   2,
					Title:     "Old Title",
					CreatedAt: time.Now(),
				}, nil)
			},
			expectedNote:  nil,
			expectedError: namederrors.ErrNoAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			note, err := usecase.UpdateNote(ctx, tt.userID, tt.noteID, tt.title, tt.isArchived)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, note)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedNote.ID, note.ID)
			}
		})
	}
}

func TestNotesUsecase_DeleteNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockNotesRepository(ctrl)
	usecase := NewNotesUsecase(mockRepo)

	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)

	tests := []struct {
		name          string
		userID        uint64
		noteID        uint64
		setupMocks    func()
		expectedError error
	}{
		{
			name:   "success",
			userID: 1,
			noteID: 1,
			setupMocks: func() {
				mockRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(&models.Note{
					ID:        1,
					OwnerID:   1,
					Title:     "Test Note",
					CreatedAt: time.Now(),
				}, nil)
				mockRepo.EXPECT().DeleteNote(ctx, uint64(1)).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "no access",
			userID: 1,
			noteID: 1,
			setupMocks: func() {
				mockRepo.EXPECT().GetNoteById(ctx, uint64(1)).Return(&models.Note{
					ID:        1,
					OwnerID:   2,
					Title:     "Test Note",
					CreatedAt: time.Now(),
				}, nil)
			},
			expectedError: namederrors.ErrNoAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			err := usecase.DeleteNote(ctx, tt.userID, tt.noteID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
