package repository

import (
	"backend/models"
	namederrors "backend/named_errors"
	"backend/store"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

type BlocksRepository struct {
	Store *store.Store
}

func NewBlocksRepository(store *store.Store) *BlocksRepository {
	return &BlocksRepository{
		Store: store,
	}
}

func (r *BlocksRepository) CreateBlock(ctx context.Context, noteID uint64, blockType models.BlockType, position float64) (*models.Block, error) {
	log.Info().Uint64("note_id", noteID).Str("type", string(blockType)).Float64("position", position).Msg("CreateBlock: start")

	r.Store.Mu.Lock()
	defer r.Store.Mu.Unlock()

	tx, err := r.Store.Postgres.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("CreateBlock: begin transaction failed")
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	blockQuery := `
		INSERT INTO block (note_id, type, position)
		VALUES ($1, $2, $3)
		RETURNING id, note_id, type, position, created_at, updated_at, last_edited_by
	`

	block := &models.Block{}
	var lastEditedBy sql.NullInt64

	err = tx.QueryRowContext(ctx, blockQuery, noteID, blockType, position).Scan(
		&block.ID,
		&block.NoteID,
		&block.Type,
		&block.Position,
		&block.CreatedAt,
		&block.UpdatedAt,
		&lastEditedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create block: %w", err)
	}

	if blockType == models.BlockTypeText {
		textQuery := `
			INSERT INTO block_text (block_id, text)
			VALUES ($1, $2)
		`
		_, err = tx.ExecContext(ctx, textQuery, block.ID, "")
		if err != nil {
			return nil, fmt.Errorf("failed to create block_text: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return block, nil
}

func (r *BlocksRepository) GetBlocksByNoteID(ctx context.Context, noteID uint64) ([]models.BlockWithContent, error) {
	r.Store.Mu.RLock()
	defer r.Store.Mu.RUnlock()

	query := `
		SELECT b.id, b.note_id, b.type, b.position, b.created_at, b.updated_at,
		       bt.text,
		       COALESCE(
		           json_agg(
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
		           ) FILTER (WHERE btf.id IS NOT NULL),
		           '[]'
		       ) as formats
		FROM block b
		LEFT JOIN block_text bt ON b.id = bt.block_id
		LEFT JOIN block_text_format btf ON bt.id = btf.block_text_id
		WHERE b.note_id = $1
		GROUP BY b.id, b.note_id, b.type, b.position, b.created_at, b.updated_at, bt.text
		ORDER BY b.position ASC
	`

	rows, err := r.Store.Postgres.DB.QueryContext(ctx, query, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query blocks: %w", err)
	}
	defer rows.Close()

	blocks := make([]models.BlockWithContent, 0)

	for rows.Next() {
		var block models.BlockWithContent
		var text sql.NullString
		var formatsJSON []byte

		err := rows.Scan(
			&block.ID,
			&block.NoteID,
			&block.Type,
			&block.Position,
			&block.CreatedAt,
			&block.UpdatedAt,
			&text,
			&formatsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan block: %w", err)
		}

		if text.Valid {
			block.Text = text.String
		}

		if len(formatsJSON) > 0 {
			var formats []models.BlockTextFormat
			if err := json.Unmarshal(formatsJSON, &formats); err != nil {
				log.Warn().Err(err).Msg("Failed to unmarshal formats")
			} else {
				block.Formats = formats
			}
		}

		blocks = append(blocks, block)
	}

	return blocks, nil
}

func (r *BlocksRepository) GetBlockByID(ctx context.Context, blockID uint64) (*models.BlockWithContent, error) {

	r.Store.Mu.RLock()
	defer r.Store.Mu.RUnlock()

	query := `
		SELECT b.id, b.note_id, b.type, b.position, b.created_at, b.updated_at,
		       bt.text
		FROM block b
		LEFT JOIN block_text bt ON b.id = bt.block_id
		WHERE b.id = $1
	`

	block := &models.BlockWithContent{}
	var text sql.NullString

	err := r.Store.Postgres.DB.QueryRowContext(ctx, query, blockID).Scan(
		&block.ID,
		&block.NoteID,
		&block.Type,
		&block.Position,
		&block.CreatedAt,
		&block.UpdatedAt,
		&text,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, namederrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	if text.Valid {
		block.Text = text.String

		formatsQuery := `
			SELECT btf.id, btf.block_text_id, btf.start_offset, btf.end_offset, 
			       btf.bold, btf.italic, btf.underline, btf.strikethrough, 
			       btf.link, btf.font, btf.size
			FROM block_text_format btf
			JOIN block_text bt ON btf.block_text_id = bt.id
			WHERE bt.block_id = $1
			ORDER BY btf.start_offset
		`

		rows, err := r.Store.Postgres.DB.QueryContext(ctx, formatsQuery, blockID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to query formats")
			return nil, fmt.Errorf("failed to query formats: %w", err)
		}
		defer rows.Close()

		formats := make([]models.BlockTextFormat, 0)
		for rows.Next() {
			var format models.BlockTextFormat
			var link sql.NullString

			err := rows.Scan(
				&format.ID,
				&format.BlockTextID,
				&format.StartOffset,
				&format.EndOffset,
				&format.Bold,
				&format.Italic,
				&format.Underline,
				&format.Strikethrough,
				&link,
				&format.Font,
				&format.Size,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to scan format: %w", err)
			}

			if link.Valid {
				format.Link = &link.String
			}

			formats = append(formats, format)
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating formats: %w", err)
		}

		block.Formats = formats
	}

	log.Info().Uint64("block_id", blockID).Msg("GetBlockByID completed successfully")
	return block, nil
}

func (r *BlocksRepository) UpdateBlockText(ctx context.Context, blockID uint64, text string, formats []models.BlockTextFormat) (*models.BlockWithContent, error) {
	r.Store.Mu.Lock()

	tx, err := r.Store.Postgres.DB.BeginTx(ctx, nil)
	if err != nil {
		r.Store.Mu.Unlock()
		log.Error().Err(err).Uint64("block_id", blockID).Msg("UpdateBlockText: begin tx failed")
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	updateBlockQuery := `UPDATE block SET updated_at = $1 WHERE id = $2`
	if _, err = tx.ExecContext(ctx, updateBlockQuery, time.Now().UTC(), blockID); err != nil {
		r.Store.Mu.Unlock()
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
	if err = tx.QueryRowContext(ctx, updateTextQuery, text, time.Now().UTC(), blockID).Scan(&blockTextID); err != nil {
		r.Store.Mu.Unlock()
		log.Error().Err(err).Uint64("block_id", blockID).Msg("UpdateBlockText: update block_text failed")
		return nil, fmt.Errorf("failed to update block_text: %w", err)
	}

	deleteFormatsQuery := `DELETE FROM block_text_format WHERE block_text_id = $1`
	if _, err = tx.ExecContext(ctx, deleteFormatsQuery, blockTextID); err != nil {
		r.Store.Mu.Unlock()
		log.Error().Err(err).Uint64("block_id", blockID).Uint64("block_text_id", blockTextID).Msg("UpdateBlockText: delete old formats failed")
		return nil, fmt.Errorf("failed to delete old formats: %w", err)
	}

	if len(formats) > 0 {
		insertFormatQuery := `
			INSERT INTO block_text_format (block_text_id, start_offset, end_offset, bold, italic, underline, strikethrough, link, font, size)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`
		for _, f := range formats {
			var link interface{}
			if f.Link != nil {
				link = *f.Link
			}
			if _, err = tx.ExecContext(ctx, insertFormatQuery,
				blockTextID, f.StartOffset, f.EndOffset, f.Bold, f.Italic, f.Underline, f.Strikethrough, link, f.Font, f.Size,
			); err != nil {
				r.Store.Mu.Unlock()
				log.Error().Err(err).Uint64("block_id", blockID).Uint64("block_text_id", blockTextID).Msg("UpdateBlockText: insert format failed")
				return nil, fmt.Errorf("failed to insert format: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		r.Store.Mu.Unlock()
		log.Error().Err(err).Uint64("block_id", blockID).Msg("UpdateBlockText: commit failed")
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.Store.Mu.Unlock()

	block, err := r.GetBlockByID(ctx, blockID)
	if err != nil {
		log.Error().Err(err).Uint64("block_id", blockID).Msg("UpdateBlockText: GetBlockByID failed")
		return nil, err
	}
	log.Info().Uint64("block_id", blockID).Msg("UpdateBlockText: success")
	return block, nil
}

func (r *BlocksRepository) UpdateBlockPosition(ctx context.Context, blockID uint64, position float64) (*models.Block, error) {
	r.Store.Mu.Lock()
	defer r.Store.Mu.Unlock()

	query := `
		UPDATE block
		SET position = $1, updated_at = $2
		WHERE id = $3
		RETURNING id, note_id, type, position, created_at, updated_at
	`

	block := &models.Block{}
	err := r.Store.Postgres.DB.QueryRowContext(ctx, query, position, time.Now().UTC(), blockID).Scan(
		&block.ID,
		&block.NoteID,
		&block.Type,
		&block.Position,
		&block.CreatedAt,
		&block.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, namederrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update block position: %w", err)
	}

	return block, nil
}

func (r *BlocksRepository) DeleteBlock(ctx context.Context, blockID uint64) error {
	r.Store.Mu.Lock()
	defer r.Store.Mu.Unlock()

	query := `DELETE FROM block WHERE id = $1`

	result, err := r.Store.Postgres.DB.ExecContext(ctx, query, blockID)
	if err != nil {
		return fmt.Errorf("failed to delete block: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return namederrors.ErrNotFound
	}

	log.Info().Uint64("block_id", blockID).Msg("Block deleted")
	return nil
}

func (r *BlocksRepository) GetBlockNoteID(ctx context.Context, blockID uint64) (uint64, error) {
	r.Store.Mu.RLock()
	defer r.Store.Mu.RUnlock()

	query := `SELECT note_id FROM block WHERE id = $1`

	var noteID uint64
	err := r.Store.Postgres.DB.QueryRowContext(ctx, query, blockID).Scan(&noteID)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, namederrors.ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get block note_id: %w", err)
	}

	return noteID, nil
}

func (r *BlocksRepository) GetBlocksByNoteIDForPositionCalc(ctx context.Context, noteID uint64, excludeBlockID uint64) ([]struct {
	ID       uint64
	Position float64
}, error) {
	r.Store.Mu.RLock()
	defer r.Store.Mu.RUnlock()

	query := `
		SELECT id, position
		FROM block
		WHERE note_id = $1 AND id != $2
		ORDER BY position
	`

	rows, err := r.Store.Postgres.DB.QueryContext(ctx, query, noteID, excludeBlockID)
	if err != nil {
		return nil, fmt.Errorf("failed to query blocks for position calc: %w", err)
	}
	defer rows.Close()

	blocks := make([]struct {
		ID       uint64
		Position float64
	}, 0)

	for rows.Next() {
		var block struct {
			ID       uint64
			Position float64
		}
		if err := rows.Scan(&block.ID, &block.Position); err != nil {
			return nil, fmt.Errorf("failed to scan block: %w", err)
		}
		blocks = append(blocks, block)
	}

	return blocks, nil
}
