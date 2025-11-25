package repository

import (
	"backend/notes_service/internal/constants"
	"backend/notes_service/internal/models"
	"backend/notes_service/logger"
	"backend/pkg/store"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type BlocksRepository struct {
	db store.DB
}

func NewBlocksRepository(db store.DB) *BlocksRepository {
	return &BlocksRepository{db: db}
}

func (r *BlocksRepository) CreateTextBlock(ctx context.Context, noteID uint64, position float64, userID uint64) (*models.Block, error) {
	log := logger.FromContext(ctx)
	now := time.Now().UTC()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("CreateTextBlock: begin tx failed")
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Error().Err(err).Msg("CreateTextBlock: rollback failed")
		}
	}()

	insertBlockQuery := `
		INSERT INTO block (note_id, type, position, last_edited_by, created_at, updated_at)
		VALUES ($1, 'text', $2, $3, $4, $5)
		RETURNING id
	`
	var blockID uint64
	if err := tx.QueryRowContext(ctx, insertBlockQuery, noteID, position, userID, now, now).Scan(&blockID); err != nil {
		log.Error().Err(err).Msg("CreateTextBlock: insert block failed")
		return nil, fmt.Errorf("failed to create block: %w", err)
	}

	insertTextQuery := `
		INSERT INTO block_text (block_id, text, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`
	if _, err := tx.ExecContext(ctx, insertTextQuery, blockID, "", now, now); err != nil {
		log.Error().Err(err).Uint64("block_id", blockID).Msg("CreateTextBlock: insert block_text failed")
		return nil, fmt.Errorf("failed to create block_text: %w", err)
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("CreateTextBlock: commit failed")
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.GetBlockByID(ctx, blockID)
}

func (r *BlocksRepository) CreateAttachmentBlock(ctx context.Context, noteID uint64, position float64, fileID uint64, userID uint64) (*models.Block, error) {
	log := logger.FromContext(ctx)
	now := time.Now().UTC()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("CreateAttachmentBlock: begin tx failed")
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Error().Err(err).Msg("CreateAttachmentBlock: rollback failed")
		}
	}()

	insertBlockQuery := `
		INSERT INTO block (note_id, type, position, last_edited_by, created_at, updated_at)
		VALUES ($1, 'attachment', $2, $3, $4, $5)
		RETURNING id
	`
	var blockID uint64
	if err := tx.QueryRowContext(ctx, insertBlockQuery, noteID, position, userID, now, now).Scan(&blockID); err != nil {
		log.Error().Err(err).Msg("CreateAttachmentBlock: insert block failed")
		return nil, fmt.Errorf("failed to create block: %w", err)
	}

	insertAttachQuery := `
		INSERT INTO block_attachment (block_id, file_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`
	if _, err := tx.ExecContext(ctx, insertAttachQuery, blockID, fileID, now, now); err != nil {
		log.Error().Err(err).Uint64("block_id", blockID).Msg("CreateAttachmentBlock: insert block_attachment failed")
		return nil, fmt.Errorf("failed to create block_attachment: %w", err)
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("CreateAttachmentBlock: commit failed")
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.GetBlockByID(ctx, blockID)
}

func (r *BlocksRepository) CreateCodeBlock(ctx context.Context, noteID uint64, position float64, userID uint64) (*models.Block, error) {
	log := logger.FromContext(ctx)
	now := time.Now().UTC()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("CreateCodeBlock: failed to begin transaction")
		return nil, fmt.Errorf("CreateCodeBlock: failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Error().Err(err).Msg("CreateCodeBlock: rollback failed")
		}
	}()

	insertBlockQuery := `
		INSERT INTO block (note_id, type, position, last_edited_by, created_at, updated_at)
		VALUES ($1, 'code', $2, $3, $4, $5)
		RETURNING id
	`
	var blockID uint64
	if err := tx.QueryRowContext(ctx, insertBlockQuery, noteID, position, userID, now, now).Scan(&blockID); err != nil {
		log.Error().Err(err).Msg("CreateCodeBlock: failed to create block")
		return nil, fmt.Errorf("CreateCodeBlock: failed to create block: %w", err)
	}

	insertCodeQuery := `
		INSERT INTO block_code (block_id, language, code_text, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	if _, err := tx.ExecContext(ctx, insertCodeQuery, blockID, "javascript", "", now, now); err != nil {
		log.Error().Err(err).Msg("CreateCodeBlock: failed to create block_code")
		return nil, fmt.Errorf("CreateCodeCode: failed to create block_code: %w", err)
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("CreateCodeBlock: failed to commit transaction")
		return nil, fmt.Errorf("CreateCodeBlock: failed to commit transaction: %w", err)
	}

	return r.GetBlockByID(ctx, blockID)
}

func (r *BlocksRepository) UpdateCodeBlock(ctx context.Context, blockID uint64, language, codeText string) (*models.Block, error) {
	log := logger.FromContext(ctx)
	now := time.Now().UTC()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("UpdateCodeBlock: failed to begin transaction")
		return nil, fmt.Errorf("UpdateCodeBlock: failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Error().Err(err).Msg("UpdateCodeBlock: rollback failed")
		}
	}()

	updateBlockQuery := `UPDATE block SET updated_at = $1 WHERE id = $2`
	if _, err = tx.ExecContext(ctx, updateBlockQuery, now, blockID); err != nil {
		log.Error().Err(err).Msg("UpdateCodeBlock: failed to update block timestamp")
		return nil, fmt.Errorf("UpdateCodeBlock: failed to update block timestamp: %w", err)
	}

	updateCodeQuery := `
        INSERT INTO block_code (block_id, language, code_text, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (block_id) DO UPDATE 
        SET language = EXCLUDED.language, code_text = EXCLUDED.code_text, updated_at = EXCLUDED.updated_at
	`
	if _, err := tx.ExecContext(ctx, updateCodeQuery, blockID, language, codeText, now, now); err != nil {
		log.Error().Err(err).Msg("UpdateCodeBlock: failed to update/insert block_code")
		return nil, fmt.Errorf("UpdateCodeBlock: failed to update/insert block_code: %w", err)
	}

	if err = tx.Commit(); err != nil {
		log.Error().Err(err).Msg("UpdateCodeBlock: failed to commit transaction")
		return nil, fmt.Errorf("UpdateCodeBlock: failed to commit transaction: %w", err)
	}
	return r.GetBlockByID(ctx, blockID)
}

func (r *BlocksRepository) GetBlockByID(ctx context.Context, blockID uint64) (*models.Block, error) {
	log := logger.FromContext(ctx)

	blocks, err := r.GetBlocksByIDs(ctx, []uint64{blockID})
	if err != nil {
		return nil, err
	}

	if len(blocks) == 0 {
		log.Warn().Uint64("block_id", blockID).Msg("block not found")
		return nil, constants.ErrNotFound
	}

	return &blocks[0], nil
}

func (r *BlocksRepository) GetBlocksByIDs(ctx context.Context, blockIDs []uint64) ([]models.Block, error) {
	if len(blockIDs) == 0 {
		return []models.Block{}, nil
	}

	baseBlocks, err := r.getBaseBlocksByIDs(ctx, blockIDs)
	if err != nil {
		return nil, err
	}

	if len(baseBlocks) == 0 {
		return []models.Block{}, nil
	}

	var textBlockIDs, codeBlockIDs, attachmentBlockIDs []uint64
	blockMap := make(map[uint64]models.BaseBlock)

	for _, base := range baseBlocks {
		blockMap[base.ID] = base
		switch base.Type {
		case models.BlockTypeText:
			textBlockIDs = append(textBlockIDs, base.ID)
		case models.BlockTypeCode:
			codeBlockIDs = append(codeBlockIDs, base.ID)
		case models.BlockTypeAttachment:
			attachmentBlockIDs = append(attachmentBlockIDs, base.ID)
		}
	}

	textContents, err := r.getTextContentsBatch(ctx, textBlockIDs)
	if err != nil {
		return nil, err
	}

	codeContents, err := r.getCodeContentsBatch(ctx, codeBlockIDs)
	if err != nil {
		return nil, err
	}

	attachmentContents, err := r.getAttachmentContentsBatch(ctx, attachmentBlockIDs)
	if err != nil {
		return nil, err
	}

	blocks := make([]models.Block, 0, len(baseBlocks))
	for _, base := range baseBlocks {
		block := models.Block{BaseBlock: base}

		switch base.Type {
		case models.BlockTypeText:
			if content, ok := textContents[base.ID]; ok {
				block.Content = content
			}
		case models.BlockTypeCode:
			if content, ok := codeContents[base.ID]; ok {
				block.Content = content
			}
		case models.BlockTypeAttachment:
			if content, ok := attachmentContents[base.ID]; ok {
				block.Content = content
			}
		}

		blocks = append(blocks, block)
	}

	return blocks, nil
}

func (r *BlocksRepository) GetBlocksByNoteID(ctx context.Context, noteID uint64) ([]models.Block, error) {
	log := logger.FromContext(ctx)

	query := `
		SELECT id
		FROM block
		WHERE note_id = $1
		ORDER BY position ASC
	`

	rows, err := r.db.QueryContext(ctx, query, noteID)
	if err != nil {
		log.Error().Err(err).Msg("failed to query block IDs")
		return nil, fmt.Errorf("failed to query block IDs: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Error().Err(err).Msg("GetBlocksByNoteID: failed to close rows")
		}
	}()

	var blockIDs []uint64
	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			log.Error().Err(err).Msg("failed to scan block ID")
			return nil, fmt.Errorf("failed to scan block ID: %w", err)
		}
		blockIDs = append(blockIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating block IDs: %w", err)
	}

	return r.GetBlocksByIDs(ctx, blockIDs)
}

func (r *BlocksRepository) getBaseBlocksByIDs(ctx context.Context, blockIDs []uint64) ([]models.BaseBlock, error) {
	if len(blockIDs) == 0 {
		return []models.BaseBlock{}, nil
	}

	log := logger.FromContext(ctx)

	query := `
		SELECT id, note_id, type, position, created_at, updated_at
		FROM block
		WHERE id = ANY($1)
		ORDER BY position ASC
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(blockIDs))
	if err != nil {
		log.Error().Err(err).Msg("failed to query base blocks")
		return nil, fmt.Errorf("failed to query base blocks: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Error().Err(err).Msg("getBaseBlocksByIDs: failed to close rows")
		}
	}()

	var blocks []models.BaseBlock
	for rows.Next() {
		var block models.BaseBlock
		if err := rows.Scan(&block.ID, &block.NoteID, &block.Type, &block.Position, &block.CreatedAt, &block.UpdatedAt); err != nil {
			log.Error().Err(err).Msg("failed to scan base block")
			return nil, fmt.Errorf("failed to scan base block: %w", err)
		}
		blocks = append(blocks, block)
	}

	return blocks, rows.Err()
}

func (r *BlocksRepository) getTextContentsBatch(ctx context.Context, blockIDs []uint64) (map[uint64]models.TextContent, error) {
	if len(blockIDs) == 0 {
		return make(map[uint64]models.TextContent), nil
	}

	log := logger.FromContext(ctx)

	query := `
		SELECT 
		    bt.block_id,
		    bt.text,
		    COALESCE(
		        (SELECT json_agg(
		            json_build_object(
		                'id', btf.id,
		                'start_offset', btf.start_offset,
		                'end_offset', btf.end_offset,
		                'bold', btf.bold,
		                'italic', btf.italic,
		                'underline', btf.underline,
		                'strikethrough', btf.strikethrough,
		                'link', btf.link,
		                'font', btf.font,
		                'size', btf.size
		            ) ORDER BY btf.start_offset
		        ) FROM block_text_format btf WHERE btf.block_text_id = bt.id),
		        '[]'::json
		    ) as formats
		FROM block_text bt
		WHERE bt.block_id = ANY($1)
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(blockIDs))
	if err != nil {
		log.Error().Err(err).Msg("failed to query text contents")
		return nil, fmt.Errorf("failed to query text contents: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Error().Err(err).Msg("getTextContentsBatch: failed to close rows")
		}
	}()

	contents := make(map[uint64]models.TextContent)
	for rows.Next() {
		var blockID uint64
		var text string
		var formatsJSON []byte

		if err := rows.Scan(&blockID, &text, &formatsJSON); err != nil {
			log.Error().Err(err).Msg("failed to scan text content")
			return nil, fmt.Errorf("failed to scan text content: %w", err)
		}

		var formats []models.BlockTextFormat
		if err := json.Unmarshal(formatsJSON, &formats); err != nil {
			log.Warn().Err(err).Msg("failed to unmarshal formats")
			formats = []models.BlockTextFormat{}
		}

		contents[blockID] = models.TextContent{
			Text:    text,
			Formats: formats,
		}
	}

	return contents, rows.Err()
}

func (r *BlocksRepository) getCodeContentsBatch(ctx context.Context, blockIDs []uint64) (map[uint64]models.CodeContent, error) {
	if len(blockIDs) == 0 {
		return make(map[uint64]models.CodeContent), nil
	}

	log := logger.FromContext(ctx)

	query := `
		SELECT block_id, code_text, language
		FROM block_code
		WHERE block_id = ANY($1)
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(blockIDs))
	if err != nil {
		log.Error().Err(err).Msg("failed to query code contents")
		return nil, fmt.Errorf("failed to query code contents: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Error().Err(err).Msg("getCodeContentsBatch: failed to close rows")
		}
	}()

	contents := make(map[uint64]models.CodeContent)
	for rows.Next() {
		var blockID uint64
		var code, language string

		if err := rows.Scan(&blockID, &code, &language); err != nil {
			log.Error().Err(err).Msg("failed to scan code content")
			return nil, fmt.Errorf("failed to scan code content: %w", err)
		}

		contents[blockID] = models.CodeContent{
			Code:     code,
			Language: language,
		}
	}

	return contents, rows.Err()
}

func (r *BlocksRepository) getAttachmentContentsBatch(ctx context.Context, blockIDs []uint64) (map[uint64]models.AttachmentContent, error) {
	if len(blockIDs) == 0 {
		return make(map[uint64]models.AttachmentContent), nil
	}

	log := logger.FromContext(ctx)

	query := `
		SELECT 
		    ba.block_id,
		    ba.file_id,
		    ba.caption
		FROM block_attachment ba
		WHERE ba.block_id = ANY($1)
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(blockIDs))
	if err != nil {
		log.Error().Err(err).Msg("failed to query attachment contents")
		return nil, fmt.Errorf("failed to query attachment contents: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Error().Err(err).Msg("getAttachmentContentsBatch: failed to close rows")
		}
	}()

	contents := make(map[uint64]models.AttachmentContent)
	for rows.Next() {
		var blockID uint64
		var fileID int64
		var caption sql.NullString

		if err := rows.Scan(&blockID, &fileID, &caption); err != nil {
			log.Error().Err(err).Msg("failed to scan attachment content")
			return nil, fmt.Errorf("failed to scan attachment content: %w", err)
		}

		content := models.AttachmentContent{
			URL:       fmt.Sprintf("file:%d", fileID),
			MimeType:  "application/octet-stream",
			SizeBytes: 0,
		}

		if caption.Valid {
			captionStr := caption.String
			content.Caption = &captionStr
		}
		
		// Ширину и высоту пока оставляем null, Gateway их заполнит

		contents[blockID] = content
	}

	return contents, rows.Err()
}

func (r *BlocksRepository) UpdateBlockText(ctx context.Context, blockID uint64, text string, formats []models.BlockTextFormat) (*models.Block, error) {
	log := logger.FromContext(ctx)
	now := time.Now().UTC()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Uint64("block_id", blockID).Msg("UpdateBlockText: begin tx failed")
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			log.Error().Err(err).Msg("CreateBlock: rollback failed")
		}
	}()

	updateBlockQuery := `UPDATE block SET updated_at = $1 WHERE id = $2`
	if _, err = tx.ExecContext(ctx, updateBlockQuery, now, blockID); err != nil {
		log.Error().Err(err).Uint64("block_id", blockID).Msg("UpdateBlockText: update block timestamp failed")
		return nil, fmt.Errorf("failed to update block timestamp: %w", err)
	}

	var blockTextID uint64
	updateTextQuery := `
		UPDATE block_text
		SET text = $1, updated_at = $2
		WHERE block_id = $3
		RETURNING id
	`
	if err = tx.QueryRowContext(ctx, updateTextQuery, text, now, blockID).Scan(&blockTextID); err != nil {
		log.Error().Err(err).Uint64("block_id", blockID).Msg("UpdateBlockText: update block_text failed")
		return nil, fmt.Errorf("failed to update block_text: %w", err)
	}

	deleteFormatsQuery := `DELETE FROM block_text_format WHERE block_text_id = $1`
	if _, err = tx.ExecContext(ctx, deleteFormatsQuery, blockTextID); err != nil {
		log.Error().Err(err).Uint64("block_text_id", blockTextID).Msg("UpdateBlockText: delete old formats failed")
		return nil, fmt.Errorf("failed to delete old formats: %w", err)
	}

	if len(formats) > 0 {
		insertFormatQuery := `
			INSERT INTO block_text_format (block_text_id, start_offset, end_offset, bold, italic, underline, strikethrough, link, font, size, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`
		for _, f := range formats {
			var link interface{}
			if f.Link != nil {
				link = *f.Link
			}
			if _, err = tx.ExecContext(ctx, insertFormatQuery,
				blockTextID, f.StartOffset, f.EndOffset, f.Bold, f.Italic, f.Underline, f.Strikethrough, link, f.Font, f.Size, now, now,
			); err != nil {
				log.Error().Err(err).Uint64("block_text_id", blockTextID).Msg("UpdateBlockText: insert format failed")
				return nil, fmt.Errorf("failed to insert format: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		log.Error().Err(err).Uint64("block_id", blockID).Msg("UpdateBlockText: commit failed")
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	block, err := r.GetBlockByID(ctx, blockID)
	if err != nil {
		log.Error().Err(err).Uint64("block_id", blockID).Msg("UpdateBlockText: GetBlockByID after update failed")
		return nil, err
	}
	return block, nil
}

func (r *BlocksRepository) UpdateBlockPosition(ctx context.Context, blockID uint64, position float64) (*models.Block, error) {
	log := logger.FromContext(ctx)
	now := time.Now().UTC()

	query := `
        UPDATE block
        SET position = $1, updated_at = $2
        WHERE id = $3
    `

	result, err := r.db.ExecContext(ctx, query, position, now, blockID)
	if err != nil {
		log.Error().Err(err).Msg("failed to update block position")
		return nil, fmt.Errorf("failed to update block position: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("failed to get rows affected")
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		log.Warn().Uint64("block_id", blockID).Msg("block not found for position update")
		return nil, constants.ErrNotFound
	}

	return r.GetBlockByID(ctx, blockID)
}

func (r *BlocksRepository) DeleteBlock(ctx context.Context, blockID uint64) error {
	log := logger.FromContext(ctx)

	query := `DELETE FROM block WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, blockID)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete block")
		return fmt.Errorf("failed to delete block: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("failed to get rows affected")
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		log.Warn().Uint64("block_id", blockID).Msg("block not found for deletion")
		return constants.ErrNotFound
	}

	return nil
}

func (r *BlocksRepository) GetBlockNoteID(ctx context.Context, blockID uint64) (uint64, error) {
	log := logger.FromContext(ctx)

	query := `SELECT note_id FROM block WHERE id = $1`

	var noteID uint64
	err := r.db.QueryRowContext(ctx, query, blockID).Scan(&noteID)

	if errors.Is(err, sql.ErrNoRows) {
		log.Warn().Err(err).Uint64("block_id", blockID).Msg("block not found, cannot get note_id")
		return 0, constants.ErrNotFound
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to get block note_id")
		return 0, fmt.Errorf("failed to get block note_id: %w", err)
	}

	return noteID, nil
}

type BlockPositionInfo struct {
	ID       uint64
	Position float64
}

func (r *BlocksRepository) GetBlocksByNoteIDForPositionCalc(ctx context.Context, noteID uint64, excludeBlockID uint64) ([]BlockPositionInfo, error) {
	log := logger.FromContext(ctx)

	query := `
		SELECT id, position
		FROM block
		WHERE note_id = $1 AND id != $2
		ORDER BY position
	`

	rows, err := r.db.QueryContext(ctx, query, noteID, excludeBlockID)
	if err != nil {
		log.Error().Err(err).Msg("failed to query blocks for position calc")
		return nil, fmt.Errorf("failed to query blocks for position calc: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close rows")
		}
	}()

	var blocks []BlockPositionInfo

	for rows.Next() {
		var block BlockPositionInfo
		if err := rows.Scan(&block.ID, &block.Position); err != nil {
			log.Error().Err(err).Msg("failed to scan block for position calc")
			return nil, fmt.Errorf("failed to scan block: %w", err)
		}
		blocks = append(blocks, block)
	}

	return blocks, nil
}
