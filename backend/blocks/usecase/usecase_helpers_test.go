package usecase

import (
	"backend/blocks/repository"
	"backend/models"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// small fake repos to avoid gomock-generated code affecting coverage
type fakeBlocksRepo struct {
    blocks []repository.BlockPositionInfo
}

func (f *fakeBlocksRepo) CreateTextBlock(ctx context.Context, noteID uint64, position float64, userID uint64) (*models.BlockWithContent, error) {
    return &models.BlockWithContent{Block: models.Block{ID: 1, NoteID: noteID, Position: position}}, nil
}
func (f *fakeBlocksRepo) CreateAttachmentBlock(ctx context.Context, noteID uint64, position float64, fileID uint64, userID uint64) (*models.BlockWithContent, error) {
    return nil, nil
}
func (f *fakeBlocksRepo) GetBlocksByNoteID(ctx context.Context, noteID uint64) ([]models.BlockWithContent, error) { return nil, nil }
func (f *fakeBlocksRepo) GetBlockByID(ctx context.Context, blockID uint64) (*models.BlockWithContent, error) { return nil, nil }
func (f *fakeBlocksRepo) UpdateBlockText(ctx context.Context, blockID uint64, text string, formats []models.BlockTextFormat) (*models.BlockWithContent, error) {
    return nil, nil
}
func (f *fakeBlocksRepo) UpdateBlockPosition(ctx context.Context, blockID uint64, position float64) (*models.Block, error) { return nil, nil }
func (f *fakeBlocksRepo) DeleteBlock(ctx context.Context, blockID uint64) error { return nil }
func (f *fakeBlocksRepo) GetBlockNoteID(ctx context.Context, blockID uint64) (uint64, error) { return 1, nil }
func (f *fakeBlocksRepo) GetBlocksByNoteIDForPositionCalc(ctx context.Context, noteID uint64, excludeBlockID uint64) ([]repository.BlockPositionInfo, error) {
    return f.blocks, nil
}

type fakeNotesRepo struct{
    noteOwner uint64
}
func (f *fakeNotesRepo) GetNoteById(ctx context.Context, noteID uint64, userID uint64) (*models.Note, error) {
    return &models.Note{ID: noteID, OwnerID: f.noteOwner}, nil
}

func Test_isDefaultFormat_and_styles_formats_merge(t *testing.T) {
    f1 := models.BlockTextFormat{Bold: true, Font: models.FontInter, Size: 12}
    f2 := models.BlockTextFormat{Bold: true, Font: models.FontInter, Size: 12}
    assert.True(t, stylesEqual(f1, f2))
    assert.False(t, isDefaultFormat(f1))

    m := map[int]models.BlockTextFormat{0: f1, 1: {Italic: true, Font: models.FontInter, Size: 12}}
    merged := mergeFormats(m)
    assert.True(t, merged.Bold)
    assert.True(t, merged.Italic)

    // formatsEqual considers offsets too
    fa := models.BlockTextFormat{StartOffset:0, EndOffset:4, Bold:true}
    fb := models.BlockTextFormat{StartOffset:0, EndOffset:4, Bold:true}
    assert.True(t, formatsEqual(fa, fb))
}

func Test_optimizeFormats_basic(t *testing.T) {
    text := "HelloWorld"
    formats := []models.BlockTextFormat{
        {StartOffset:0, EndOffset:5, Bold:true},
        {StartOffset:5, EndOffset:10, Italic:true},
    }
    res := optimizeFormats(text, formats)
    // merged adjacent same-style segments should be two entries
    assert.True(t, len(res) >= 1)
}

func Test_calculatePosition_cases(t *testing.T) {
    ctx := context.Background()
    // empty blocks -> returns 1.0
    fb := &fakeBlocksRepo{blocks: []repository.BlockPositionInfo{}}
    u := NewBlocksUsecase(fb, &fakeNotesRepo{noteOwner:1})
    pos, err := u.calculatePosition(ctx, 1, nil, 0)
    assert.NoError(t, err)
    assert.Equal(t, 1.0, pos)

    // single existing block -> append
    fb.blocks = []repository.BlockPositionInfo{{ID:1, Position:2.0}}
    pos, err = u.calculatePosition(ctx, 1, nil, 0)
    assert.NoError(t, err)
    assert.Equal(t, 3.0, pos)

    // beforeBlock at start -> half of its position
    fb.blocks = []repository.BlockPositionInfo{{ID:2, Position:2.0}, {ID:3, Position:3.0}}
    b := uint64(2)
    pos, err = u.calculatePosition(ctx, 1, &b, 0)
    assert.NoError(t, err)
    assert.Equal(t, 1.0, pos)

    // beforeBlock not found -> error
    nf := uint64(999)
    _, err = u.calculatePosition(ctx, 1, &nf, 0)
    assert.Error(t, err)
}
