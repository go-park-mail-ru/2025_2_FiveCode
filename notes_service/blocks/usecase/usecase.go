package usecase

import (
	blocksRepository "backend/notes_service/blocks/repository"
	"backend/notes_service/internal/constants"
	"backend/notes_service/internal/models"
	"backend/notes_service/logger"
	"context"
	"fmt"
	"sort"
)

type BlocksUsecase struct {
	BlocksRepo BlocksRepository
	NotesRepo  NotesRepository
}

//go:generate mockgen -source=usecase.go -destination=../mock/mock_usecase.go -package=mock
type BlocksRepository interface {
	CreateTextBlock(ctx context.Context, noteID uint64, position float64, userID uint64) (*models.Block, error)
	CreateAttachmentBlock(ctx context.Context, noteID uint64, position float64, fileID uint64, userID uint64) (*models.Block, error)
	CreateCodeBlock(ctx context.Context, noteID uint64, position float64, userID uint64) (*models.Block, error)
	GetBlocksByNoteID(ctx context.Context, noteID uint64) ([]models.Block, error)
	GetBlockByID(ctx context.Context, blockID uint64) (*models.Block, error)
	UpdateBlockText(ctx context.Context, blockID uint64, text string, formats []models.BlockTextFormat) (*models.Block, error)
	UpdateBlockPosition(ctx context.Context, blockID uint64, position float64) (*models.Block, error)
	UpdateCodeBlock(ctx context.Context, blockID uint64, language, codeText string) (*models.Block, error)
	DeleteBlock(ctx context.Context, blockID uint64) error
	GetBlockNoteID(ctx context.Context, blockID uint64) (uint64, error)
	GetBlocksByNoteIDForPositionCalc(ctx context.Context, noteID uint64, excludeBlockID uint64) ([]blocksRepository.BlockPositionInfo, error)
}

type NotesRepository interface {
	GetNoteById(ctx context.Context, noteID uint64, userID uint64) (*models.Note, error)
}

func NewBlocksUsecase(blocksRepo BlocksRepository, notesRepo NotesRepository) *BlocksUsecase {
	return &BlocksUsecase{
		BlocksRepo: blocksRepo,
		NotesRepo:  notesRepo,
	}
}

func (u *BlocksUsecase) CreateTextBlock(ctx context.Context, userID, noteID uint64, beforeBlockID *uint64) (*models.Block, error) {
	log := logger.FromContext(ctx).With().
		Uint64("user_id", userID).
		Uint64("note_id", noteID).
		Str("type", models.BlockTypeText).
		Logger()

	if err := u.checkNoteAccess(ctx, userID, noteID); err != nil {
		return nil, err
	}

	position, err := u.calculatePosition(ctx, noteID, beforeBlockID, 0)
	if err != nil {
		log.Error().Err(err).Msg("failed to calculate position")
		return nil, fmt.Errorf("failed to calculate position: %w", err)
	}

	block, err := u.BlocksRepo.CreateTextBlock(ctx, noteID, position, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to create text block")
		return nil, fmt.Errorf("failed to create text block: %w", err)
	}

	return block, nil
}

func (u *BlocksUsecase) CreateAttachmentBlock(ctx context.Context, userID, noteID uint64, beforeBlockID *uint64, fileID uint64) (*models.Block, error) {
	log := logger.FromContext(ctx).With().
		Uint64("user_id", userID).
		Uint64("note_id", noteID).
		Str("type", models.BlockTypeAttachment).
		Uint64("file_id", fileID).
		Logger()

	if fileID == 0 {
		return nil, fmt.Errorf("file_id is required")
	}

	if err := u.checkNoteAccess(ctx, userID, noteID); err != nil {
		return nil, err
	}

	position, err := u.calculatePosition(ctx, noteID, beforeBlockID, 0)
	if err != nil {
		log.Error().Err(err).Msg("failed to calculate position")
		return nil, fmt.Errorf("failed to calculate position: %w", err)
	}

	block, err := u.BlocksRepo.CreateAttachmentBlock(ctx, noteID, position, fileID, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to create attachment block")
		return nil, fmt.Errorf("failed to create attachment block: %w", err)
	}

	return block, nil
}

func (u *BlocksUsecase) CreateCodeBlock(ctx context.Context, userID, noteID uint64, beforeBlockID *uint64) (*models.Block, error) {
	log := logger.FromContext(ctx).With().
		Uint64("user_id", userID).
		Uint64("note_id", noteID).
		Str("type", models.BlockTypeCode).
		Logger()

	if err := u.checkNoteAccess(ctx, userID, noteID); err != nil {
		return nil, err
	}
	position, err := u.calculatePosition(ctx, noteID, beforeBlockID, 0)
	if err != nil {
		log.Error().Err(err).Msg("failed to calculate position for code block")
		return nil, fmt.Errorf("failed to calculate position for code block: %w", err)
	}
	return u.BlocksRepo.CreateCodeBlock(ctx, noteID, position, userID)
}

func (u *BlocksUsecase) UpdateBlock(ctx context.Context, userID uint64, req *models.UpdateBlockRequest) (*models.Block, error) {
	log := logger.FromContext(ctx).With().
		Uint64("user_id", userID).
		Uint64("block_id", req.BlockID).
		Logger()

	noteID, err := u.BlocksRepo.GetBlockNoteID(ctx, req.BlockID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get block note id")
		return nil, fmt.Errorf("failed to get block note id: %w", err)
	}

	if err := u.checkNoteAccess(ctx, userID, noteID); err != nil {
		return nil, err
	}

	block, err := u.BlocksRepo.GetBlockByID(ctx, req.BlockID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get block")
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	switch block.Type {
	case models.BlockTypeCode:
		codeContent, ok := req.Content.(models.UpdateCodeContent)
		if !ok {
			return nil, fmt.Errorf("invalid content type for code block update")
		}
		return u.BlocksRepo.UpdateCodeBlock(ctx, req.BlockID, codeContent.Language, codeContent.Code)

	case models.BlockTypeText:
		textContent, ok := req.Content.(models.UpdateTextContent)
		if !ok {
			return nil, fmt.Errorf("invalid content type for text block update")
		}
		optimizedFormats := optimizeFormats(textContent.Text, textContent.Formats)
		return u.BlocksRepo.UpdateBlockText(ctx, req.BlockID, textContent.Text, optimizedFormats)

	case models.BlockTypeAttachment:
		return nil, fmt.Errorf("updating attachment blocks is not supported")

	default:
		return nil, fmt.Errorf("unknown block type: %s", block.Type)
	}
}

func (u *BlocksUsecase) GetBlock(ctx context.Context, userID, blockID uint64) (*models.Block, error) {
	log := logger.FromContext(ctx).With().
		Uint64("user_id", userID).
		Uint64("block_id", blockID).
		Logger()

	block, err := u.BlocksRepo.GetBlockByID(ctx, blockID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get block by id from repository")
		return nil, fmt.Errorf("failed to get block by id: %w", err)
	}

	if err := u.checkNoteAccess(ctx, userID, block.NoteID); err != nil {
		return nil, err
	}

	return block, nil
}

func (u *BlocksUsecase) DeleteBlock(ctx context.Context, userID, blockID uint64) error {
	log := logger.FromContext(ctx).With().
		Uint64("user_id", userID).
		Uint64("block_id", blockID).
		Logger()

	noteID, err := u.BlocksRepo.GetBlockNoteID(ctx, blockID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get block note id for deletion")
		return fmt.Errorf("failed to get block note id: %w", err)
	}

	if err := u.checkNoteAccess(ctx, userID, noteID); err != nil {
		return err
	}

	if err := u.BlocksRepo.DeleteBlock(ctx, blockID); err != nil {
		log.Error().Err(err).Msg("failed to delete block in repository")
		return fmt.Errorf("failed to delete block: %w", err)
	}

	return nil
}

func (u *BlocksUsecase) GetBlocks(ctx context.Context, userID, noteID uint64) ([]models.Block, error) {
	log := logger.FromContext(ctx).With().
		Uint64("user_id", userID).
		Uint64("note_id", noteID).
		Logger()

	if err := u.checkNoteAccess(ctx, userID, noteID); err != nil {
		return nil, err
	}

	blocks, err := u.BlocksRepo.GetBlocksByNoteID(ctx, noteID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get blocks from repository")
		return nil, fmt.Errorf("failed to get blocks: %w", err)
	}

	return blocks, nil
}

func (u *BlocksUsecase) UpdateBlockPosition(ctx context.Context, userID, blockID uint64, beforeBlockID *uint64) (*models.Block, error) {
	log := logger.FromContext(ctx).With().
		Uint64("user_id", userID).
		Uint64("block_id", blockID).
		Logger()

	noteID, err := u.BlocksRepo.GetBlockNoteID(ctx, blockID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get block note id for position update")
		return nil, fmt.Errorf("failed to get block note id: %w", err)
	}

	if err := u.checkNoteAccess(ctx, userID, noteID); err != nil {
		return nil, err
	}

	position, err := u.calculatePosition(ctx, noteID, beforeBlockID, blockID)
	if err != nil {
		log.Error().Err(err).Msg("failed to calculate position for update")
		return nil, fmt.Errorf("failed to calculate position: %w", err)
	}

	block, err := u.BlocksRepo.UpdateBlockPosition(ctx, blockID, position)
	if err != nil {
		log.Error().Err(err).Msg("failed to update position in repository")
		return nil, fmt.Errorf("failed to update position: %w", err)
	}

	return block, nil
}

func (u *BlocksUsecase) checkNoteAccess(ctx context.Context, userID, noteID uint64) error {
	log := logger.FromContext(ctx)
	note, err := u.NotesRepo.GetNoteById(ctx, noteID, userID)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("failed to get note for access check")
		return fmt.Errorf("failed to get note by id: %w", err)
	}

	if note.OwnerID != userID {
		log.Warn().Uint64("user_id", userID).Uint64("note_id", noteID).Uint64("owner_id", note.OwnerID).Msg("user access denied to note")
		return constants.ErrNoAccess
	}

	return nil
}

func (u *BlocksUsecase) calculatePosition(ctx context.Context, noteID uint64, beforeBlockID *uint64, excludeBlockID uint64) (float64, error) {
	blocks, err := u.BlocksRepo.GetBlocksByNoteIDForPositionCalc(ctx, noteID, excludeBlockID)
	if err != nil {
		return 0, fmt.Errorf("failed to get blocks for position calc: %w", err)
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

	var beforeBlock *blocksRepository.BlockPositionInfo
	for i := range blocks {
		if blocks[i].ID == *beforeBlockID {
			beforeBlock = &blocks[i]
			break
		}
	}

	if beforeBlock == nil {
		return 0, fmt.Errorf("before_block not found")
	}

	var prevBlock *blocksRepository.BlockPositionInfo
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
