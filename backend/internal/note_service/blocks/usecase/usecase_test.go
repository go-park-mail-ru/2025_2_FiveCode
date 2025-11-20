package usecase

import (
	"backend/blocks/repository"
	"backend/models"
	models2 "backend/pkg/models"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// minimal fake repos used by tests
type bu_fakeBlocksRepo struct{}

func (b *bu_fakeBlocksRepo) CreateTextBlock(ctx context.Context, noteID uint64, position float64, userID uint64) (*models.BlockWithContent, error) {
	return &models.BlockWithContent{Block: models2.Block{ID: 10, NoteID: noteID, Position: position}}, nil
}
func (b *bu_fakeBlocksRepo) CreateAttachmentBlock(ctx context.Context, noteID uint64, position float64, fileID uint64, userID uint64) (*models.BlockWithContent, error) {
	return &models.BlockWithContent{Block: models2.Block{ID: 11, NoteID: noteID, Position: position}}, nil
}
func (b *bu_fakeBlocksRepo) GetBlocksByNoteID(ctx context.Context, noteID uint64) ([]models.BlockWithContent, error) {
	return []models.BlockWithContent{}, nil
}
func (b *bu_fakeBlocksRepo) GetBlockByID(ctx context.Context, blockID uint64) (*models.BlockWithContent, error) {
	return &models.BlockWithContent{Block: models2.Block{ID: blockID, NoteID: 1, Position: 1.0}}, nil
}
func (b *bu_fakeBlocksRepo) UpdateBlockText(ctx context.Context, blockID uint64, text string, formats []models2.BlockTextFormat) (*models.BlockWithContent, error) {
	return &models.BlockWithContent{Block: models2.Block{ID: blockID}}, nil
}
func (b *bu_fakeBlocksRepo) UpdateBlockPosition(ctx context.Context, blockID uint64, position float64) (*models2.Block, error) {
	return &models2.Block{ID: blockID, Position: position}, nil
}
func (b *bu_fakeBlocksRepo) DeleteBlock(ctx context.Context, blockID uint64) error { return nil }
func (b *bu_fakeBlocksRepo) GetBlockNoteID(ctx context.Context, blockID uint64) (uint64, error) {
	return 1, nil
}
func (b *bu_fakeBlocksRepo) GetBlocksByNoteIDForPositionCalc(ctx context.Context, noteID uint64, excludeBlockID uint64) ([]repository.BlockPositionInfo, error) {
	return []repository.BlockPositionInfo{{ID: 1, Position: 1.0}, {ID: 2, Position: 2.0}}, nil
}

type bu_fakeNotesRepo struct{ owner uint64 }

func (f *bu_fakeNotesRepo) GetNoteById(ctx context.Context, noteID uint64, userID uint64) (*models2.Note, error) {
	return &models2.Note{ID: noteID, OwnerID: f.owner}, nil
}

// errBlocksRepo simulates repository errors for blocks usecase tests
type errBlocksRepo struct{}

func (r *errBlocksRepo) CreateTextBlock(ctx context.Context, noteID uint64, position float64, userID uint64) (*models.BlockWithContent, error) {
	return nil, nil
}
func (r *errBlocksRepo) CreateAttachmentBlock(ctx context.Context, noteID uint64, position float64, fileID uint64, userID uint64) (*models.BlockWithContent, error) {
	return nil, nil
}
func (r *errBlocksRepo) GetBlocksByNoteID(ctx context.Context, noteID uint64) ([]models.BlockWithContent, error) {
	return nil, nil
}
func (r *errBlocksRepo) GetBlockByID(ctx context.Context, blockID uint64) (*models.BlockWithContent, error) {
	return nil, nil
}
func (r *errBlocksRepo) UpdateBlockText(ctx context.Context, blockID uint64, text string, formats []models2.BlockTextFormat) (*models.BlockWithContent, error) {
	return nil, fmt.Errorf("upd")
}
func (r *errBlocksRepo) UpdateBlockPosition(ctx context.Context, blockID uint64, position float64) (*models2.Block, error) {
	return nil, fmt.Errorf("pos")
}
func (r *errBlocksRepo) DeleteBlock(ctx context.Context, blockID uint64) error {
	return fmt.Errorf("del")
}
func (r *errBlocksRepo) GetBlockNoteID(ctx context.Context, blockID uint64) (uint64, error) {
	return 0, fmt.Errorf("nb")
}
func (r *errBlocksRepo) GetBlocksByNoteIDForPositionCalc(ctx context.Context, noteID uint64, excludeBlockID uint64) ([]repository.BlockPositionInfo, error) {
	return nil, fmt.Errorf("poscalc")
}

func TestBlocksUsecase_CreateAndPosition(t *testing.T) {
	blocksRepo := &bu_fakeBlocksRepo{}
	notesRepo := &bu_fakeNotesRepo{owner: 1}
	u := NewBlocksUsecase(blocksRepo, notesRepo)

	// Create text block
	b, err := u.CreateTextBlock(context.Background(), 1, 1, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(10), b.ID)

	// Update position
	nb, err := u.UpdateBlockPosition(context.Background(), 1, 10, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(10), nb.ID)
}

func TestBlocksUsecase_CreateAttachment_FileIDRequired(t *testing.T) {
	blocksRepo := &bu_fakeBlocksRepo{}
	notesRepo := &bu_fakeNotesRepo{owner: 1}
	u := NewBlocksUsecase(blocksRepo, notesRepo)

	_, err := u.CreateAttachmentBlock(context.Background(), 1, 1, nil, 0)
	assert.Error(t, err)
}

func TestBlocksUsecase_AccessDeniedAndPositionCalc(t *testing.T) {
	// note owner differs so access denied
	blocksRepo := &bu_fakeBlocksRepo{}
	notesRepo := &bu_fakeNotesRepo{owner: 2}
	u := NewBlocksUsecase(blocksRepo, notesRepo)

	_, err := u.CreateTextBlock(context.Background(), 1, 1, nil)
	if !assert.Error(t, err) {
		t.Fatalf("expected access denied error")
	}

	// test calculatePosition when beforeBlock not found
	// use fake notes repo with correct owner
	notesRepo2 := &bu_fakeNotesRepo{owner: 1}
	u2 := NewBlocksUsecase(blocksRepo, notesRepo2)

	// beforeBlock points to non-existing id -> should return error from calculatePosition wrapped
	before := uint64(999)
	_, err = u2.CreateTextBlock(context.Background(), 1, 1, &before)
	assert.Error(t, err)
}

func TestBlocksUsecase_UpdateAndDelete_ErrorFlows(t *testing.T) {
	// use package-level errBlocksRepo that simulates repository errors
	notesRepo := &bu_fakeNotesRepo{owner: 1}
	er := &errBlocksRepo{}
	u := NewBlocksUsecase(er, notesRepo)

	// Update text should propagate error
	_, err := u.UpdateBlock(context.Background(), 1, 1, "x", nil)
	assert.Error(t, err)

	// Update position should propagate error
	_, err = u.UpdateBlockPosition(context.Background(), 1, 1, nil)
	assert.Error(t, err)

	// Delete should propagate error
	err = u.DeleteBlock(context.Background(), 1, 1)
	assert.Error(t, err)
}

// more coverage: calculatePosition branches
func TestBlocksUsecase_CalculatePositionVarious(t *testing.T) {
	blocksRepo := &bu_fakeBlocksRepo{}
	notesRepo := &bu_fakeNotesRepo{owner: 1}
	u := NewBlocksUsecase(blocksRepo, notesRepo)

	// nil before -> position should be > 0
	b, err := u.CreateTextBlock(context.Background(), 1, 0, nil)
	assert.NoError(t, err)
	assert.NotNil(t, b)

	// middle insertion: before exists in fake GetBlocksByNoteIDForPositionCalc
	before := uint64(2)
	_, err = u.CreateTextBlock(context.Background(), 1, 0, &before)
	assert.NoError(t, err)
}
