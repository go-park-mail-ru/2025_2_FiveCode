package usecase

import (
	"backend/models"
	namederrors "backend/named_errors"
	"context"
	"errors"
	"sort"
)

type BlocksUsecase struct {
	BlocksRepo BlocksRepository
	NotesRepo  NotesRepository
}

type BlocksRepository interface {
	CreateBlock(ctx context.Context, noteID uint64, blockType models.BlockType, position float64) (*models.Block, error)
	GetBlocksByNoteID(ctx context.Context, noteID uint64) ([]models.BlockWithContent, error)
	GetBlockByID(ctx context.Context, blockID uint64) (*models.BlockWithContent, error)
	UpdateBlockText(ctx context.Context, blockID uint64, text string, formats []models.BlockTextFormat) (*models.BlockWithContent, error)
	UpdateBlockPosition(ctx context.Context, blockID uint64, position float64) (*models.Block, error)
	DeleteBlock(ctx context.Context, blockID uint64) error
	GetBlockNoteID(ctx context.Context, blockID uint64) (uint64, error)
	GetBlocksByNoteIDForPositionCalc(ctx context.Context, noteID uint64, excludeBlockID uint64) ([]struct {
		ID       uint64
		Position float64
	}, error)
}

type NotesRepository interface {
	GetNoteById(ctx context.Context, noteID uint64) (*models.Note, error)
}

func NewBlocksUsecase(blocksRepo BlocksRepository, notesRepo NotesRepository) *BlocksUsecase {
	return &BlocksUsecase{
		BlocksRepo: blocksRepo,
		NotesRepo:  notesRepo,
	}
}

func (u *BlocksUsecase) CreateBlock(ctx context.Context, userID, noteID uint64, beforeBlockID *uint64) (*models.Block, error) {
	if err := u.checkNoteAccess(ctx, userID, noteID); err != nil {
		return nil, err
	}

	position, err := u.calculatePosition(ctx, noteID, beforeBlockID, 0)
	if err != nil {
		return nil, errors.New("failed to calculate position: " + err.Error())
	}

	block, err := u.BlocksRepo.CreateBlock(ctx, noteID, models.BlockTypeText, position)
	if err != nil {
		return nil, errors.New("failed to create block: " + err.Error())
	}

	return block, nil
}

func (u *BlocksUsecase) GetBlocks(ctx context.Context, userID, noteID uint64) ([]models.BlockWithContent, error) {
	if err := u.checkNoteAccess(ctx, userID, noteID); err != nil {
		return nil, err
	}

	blocks, err := u.BlocksRepo.GetBlocksByNoteID(ctx, noteID)
	if err != nil {
		return nil, errors.New("failed to get blocks: " + err.Error())
	}

	return blocks, nil
}

func (u *BlocksUsecase) GetBlock(ctx context.Context, userID, blockID uint64) (*models.BlockWithContent, error) {
	block, err := u.BlocksRepo.GetBlockByID(ctx, blockID)
	if err != nil {
		return nil, errors.New("failed to get block by id: " + err.Error())
	}

	if err := u.checkNoteAccess(ctx, userID, block.NoteID); err != nil {
		return nil, err
	}

	return block, nil
}

func (u *BlocksUsecase) UpdateBlock(ctx context.Context, userID, blockID uint64, text string, formats []models.BlockTextFormat) (*models.BlockWithContent, error) {
	noteID, err := u.BlocksRepo.GetBlockNoteID(ctx, blockID)
	if err != nil {
		return nil, errors.New("failed to get block note id: " + err.Error())
	}

	if err := u.checkNoteAccess(ctx, userID, noteID); err != nil {
		return nil, err
	}

	optimizedFormats := optimizeFormats(text, formats)

	block, err := u.BlocksRepo.UpdateBlockText(ctx, blockID, text, optimizedFormats)
	if err != nil {
		return nil, errors.New("failed to update block: " + err.Error())
	}

	return block, nil
}

func (u *BlocksUsecase) DeleteBlock(ctx context.Context, userID, blockID uint64) error {
	noteID, err := u.BlocksRepo.GetBlockNoteID(ctx, blockID)
	if err != nil {
		return errors.New("failed to get block note id: " + err.Error())
	}

	if err := u.checkNoteAccess(ctx, userID, noteID); err != nil {
		return err
	}

	if err := u.BlocksRepo.DeleteBlock(ctx, blockID); err != nil {
		return errors.New("failed to delete block: " + err.Error())
	}

	return nil
}

func (u *BlocksUsecase) UpdateBlockPosition(ctx context.Context, userID, blockID uint64, beforeBlockID *uint64) (*models.Block, error) {
	noteID, err := u.BlocksRepo.GetBlockNoteID(ctx, blockID)
	if err != nil {
		return nil, errors.New("failed to get block note id: " + err.Error())
	}

	if err := u.checkNoteAccess(ctx, userID, noteID); err != nil {
		return nil, err
	}

	position, err := u.calculatePosition(ctx, noteID, beforeBlockID, blockID)
	if err != nil {
		return nil, errors.New("failed to calculate position: " + err.Error())
	}

	block, err := u.BlocksRepo.UpdateBlockPosition(ctx, blockID, position)
	if err != nil {
		return nil, errors.New("failed to update position: " + err.Error())
	}

	return block, nil
}

func (u *BlocksUsecase) checkNoteAccess(ctx context.Context, userID, noteID uint64) error {
	note, err := u.NotesRepo.GetNoteById(ctx, noteID)
	if err != nil {
		return errors.New("failed to get note by id: " + err.Error())
	}

	if note.OwnerID != userID {
		return namederrors.ErrNoAccess
	}

	return nil
}

func (u *BlocksUsecase) calculatePosition(ctx context.Context, noteID uint64, beforeBlockID *uint64, excludeBlockID uint64) (float64, error) {
	blocks, err := u.BlocksRepo.GetBlocksByNoteIDForPositionCalc(ctx, noteID, excludeBlockID)
	if err != nil {
		return 0, err
	}

	if len(blocks) == 0 {
		return 1.0, nil
	}

	if beforeBlockID == nil {
		maxPos := blocks[0].Position
		for _, b := range blocks {
			if b.Position > maxPos {
				maxPos = b.Position
			}
		}
		return maxPos + 1.0, nil
	}

	var beforeBlock *struct {
		ID       uint64
		Position float64
	}
	for i := range blocks {
		if blocks[i].ID == *beforeBlockID {
			beforeBlock = &blocks[i]
			break
		}
	}

	if beforeBlock == nil {
		return 0, errors.New("before_block not found")
	}

	var prevBlock *struct {
		ID       uint64
		Position float64
	}
	for i := range blocks {
		if blocks[i].Position < beforeBlock.Position {
			if prevBlock == nil || blocks[i].Position > prevBlock.Position {
				prevBlock = &blocks[i]
			}
		}
	}

	if prevBlock == nil {
		return beforeBlock.Position / 2.0, nil
	}

	return (prevBlock.Position + beforeBlock.Position) / 2.0, nil
}

func optimizeFormats(text string, formats []models.BlockTextFormat) []models.BlockTextFormat {
	if len(formats) == 0 {
		return []models.BlockTextFormat{}
	}

	textLen := len(text)

	validFormats := make([]models.BlockTextFormat, 0)
	for _, f := range formats {
		if f.StartOffset >= 0 && f.EndOffset <= textLen && f.StartOffset < f.EndOffset {
			if !isDefaultFormat(f) {
				validFormats = append(validFormats, f)
			}
		}
	}

	if len(validFormats) == 0 {
		return []models.BlockTextFormat{}
	}

	type event struct {
		offset int
		format models.BlockTextFormat
		isEnd  bool
	}

	events := make([]event, 0)
	for _, f := range validFormats {
		events = append(events, event{offset: f.StartOffset, format: f, isEnd: false})
		events = append(events, event{offset: f.EndOffset, format: f, isEnd: true})
	}

	sort.Slice(events, func(i, j int) bool {
		if events[i].offset == events[j].offset {
			return !events[i].isEnd && events[j].isEnd
		}
		return events[i].offset < events[j].offset
	})

	activeFormats := make(map[int]models.BlockTextFormat)
	result := make([]models.BlockTextFormat, 0)
	lastOffset := 0
	formatIndex := 0

	for _, ev := range events {
		if len(activeFormats) > 0 && ev.offset > lastOffset {
			merged := mergeFormats(activeFormats)
			merged.StartOffset = lastOffset
			merged.EndOffset = ev.offset
			result = append(result, merged)
		}

		lastOffset = ev.offset

		if ev.isEnd {
			for idx, f := range activeFormats {
				if formatsEqual(f, ev.format) {
					delete(activeFormats, idx)
					break
				}
			}
		} else {
			activeFormats[formatIndex] = ev.format
			formatIndex++
		}
	}

	if len(result) == 0 {
		return result
	}

	merged := make([]models.BlockTextFormat, 0)
	current := result[0]

	for i := 1; i < len(result); i++ {
		if current.EndOffset == result[i].StartOffset && stylesEqual(current, result[i]) {
			current.EndOffset = result[i].EndOffset
		} else {
			merged = append(merged, current)
			current = result[i]
		}
	}
	merged = append(merged, current)

	return merged
}

func isDefaultFormat(f models.BlockTextFormat) bool {
	return !f.Bold && !f.Italic && !f.Underline && !f.Strikethrough &&
		f.Link == nil && f.Font == models.FontInter && f.Size == 12
}

func formatsEqual(f1, f2 models.BlockTextFormat) bool {
	return f1.StartOffset == f2.StartOffset &&
		f1.EndOffset == f2.EndOffset &&
		f1.Bold == f2.Bold &&
		f1.Italic == f2.Italic &&
		f1.Underline == f2.Underline &&
		f1.Strikethrough == f2.Strikethrough &&
		((f1.Link == nil && f2.Link == nil) || (f1.Link != nil && f2.Link != nil && *f1.Link == *f2.Link)) &&
		f1.Font == f2.Font &&
		f1.Size == f2.Size
}

func stylesEqual(f1, f2 models.BlockTextFormat) bool {
	return f1.Bold == f2.Bold &&
		f1.Italic == f2.Italic &&
		f1.Underline == f2.Underline &&
		f1.Strikethrough == f2.Strikethrough &&
		((f1.Link == nil && f2.Link == nil) || (f1.Link != nil && f2.Link != nil && *f1.Link == *f2.Link)) &&
		f1.Font == f2.Font &&
		f1.Size == f2.Size
}

func mergeFormats(formats map[int]models.BlockTextFormat) models.BlockTextFormat {
	result := models.BlockTextFormat{
		Font: models.FontInter,
		Size: 12,
	}

	for _, f := range formats {
		if f.Bold {
			result.Bold = true
		}
		if f.Italic {
			result.Italic = true
		}
		if f.Underline {
			result.Underline = true
		}
		if f.Strikethrough {
			result.Strikethrough = true
		}
		if f.Link != nil {
			result.Link = f.Link
		}
		result.Font = f.Font
		result.Size = f.Size
	}

	return result
}
