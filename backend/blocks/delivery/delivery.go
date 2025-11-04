package delivery

import (
	"backend/apiutils"
	"backend/logger"
	"backend/middleware"
	"backend/models"
	namederrors "backend/named_errors"
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type BlocksUsecase interface {
	CreateBlock(ctx context.Context, userID, noteID uint64, afterBlockID *uint64) (*models.Block, error)
	GetBlocks(ctx context.Context, userID, noteID uint64) ([]models.BlockWithContent, error)
	GetBlock(ctx context.Context, userID, blockID uint64) (*models.BlockWithContent, error)
	UpdateBlock(ctx context.Context, userID, blockID uint64, text string, formats []models.BlockTextFormat) (*models.BlockWithContent, error)
	DeleteBlock(ctx context.Context, userID, blockID uint64) error
	UpdateBlockPosition(ctx context.Context, userID, blockID uint64, afterBlockID *uint64) (*models.Block, error)
}

type BlocksDelivery struct {
	Usecase BlocksUsecase
}

func NewBlocksDelivery(usecase BlocksUsecase) *BlocksDelivery {
	return &BlocksDelivery{
		Usecase: usecase,
	}
}

type CreateBlockRequest struct {
	BeforeBlockID *uint64 `json:"before_block_id"`
}

func (d *BlocksDelivery) CreateBlock(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	vars := mux.Vars(r)
	noteID, err := strconv.ParseUint(vars["note_id"], 10, 64)
	if err != nil {
		log.Warn().Err(err).Str("note_id", vars["note_id"]).Msg("invalid note id")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close request body")
		}
	}()
	var req CreateBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("invalid request body")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	block, err := d.Usecase.CreateBlock(r.Context(), userID, noteID, req.BeforeBlockID)
	if err != nil {
		handleBlockError(w, r.Context(), err)
		return
	}

	apiutils.WriteJSON(w, http.StatusCreated, block)
}

func (d *BlocksDelivery) GetBlocks(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	vars := mux.Vars(r)
	noteID, err := strconv.ParseUint(vars["note_id"], 10, 64)
	if err != nil {
		log.Warn().Err(err).Str("note_id", vars["note_id"]).Msg("invalid note id")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	blocks, err := d.Usecase.GetBlocks(r.Context(), userID, noteID)
	if err != nil {
		handleBlockError(w, r.Context(), err)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"note_id": noteID,
		"blocks":  blocks,
	})
}

func (d *BlocksDelivery) GetBlock(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	vars := mux.Vars(r)
	blockID, err := strconv.ParseUint(vars["block_id"], 10, 64)
	if err != nil {
		log.Warn().Err(err).Str("block_id", vars["block_id"]).Msg("invalid block id")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid block id")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	block, err := d.Usecase.GetBlock(r.Context(), userID, blockID)
	if err != nil {
		handleBlockError(w, r.Context(), err)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, block)
}

type UpdateBlockRequest struct {
	Text    string                 `json:"text"`
	Formats []BlockTextFormatInput `json:"formats"`
}

type BlockTextFormatInput struct {
	StartOffset   int     `json:"start_offset"`
	EndOffset     int     `json:"end_offset"`
	Bold          *bool   `json:"bold,omitempty"`
	Italic        *bool   `json:"italic,omitempty"`
	Underline     *bool   `json:"underline,omitempty"`
	Strikethrough *bool   `json:"strikethrough,omitempty"`
	Link          *string `json:"link,omitempty"`
	Font          *string `json:"font,omitempty"`
	Size          *int    `json:"size,omitempty"`
}

func (d *BlocksDelivery) UpdateBlock(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	vars := mux.Vars(r)
	blockID, err := strconv.ParseUint(vars["block_id"], 10, 64)
	if err != nil {
		log.Warn().Err(err).Str("block_id", vars["block_id"]).Msg("invalid block id")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid block id")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close request body")
		}
	}()
	var req UpdateBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("invalid request body")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	formats := convertToFormats(req.Formats)

	block, err := d.Usecase.UpdateBlock(r.Context(), userID, blockID, req.Text, formats)
	if err != nil {
		handleBlockError(w, r.Context(), err)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, block)
}

func (d *BlocksDelivery) DeleteBlock(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	vars := mux.Vars(r)
	blockID, err := strconv.ParseUint(vars["block_id"], 10, 64)
	if err != nil {
		log.Warn().Err(err).Str("block_id", vars["block_id"]).Msg("invalid block id")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid block id")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	err = d.Usecase.DeleteBlock(r.Context(), userID, blockID)
	if err != nil {
		handleBlockError(w, r.Context(), err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type UpdateBlockPositionRequest struct {
	BeforeBlockID *uint64 `json:"before_block_id"`
}

func (d *BlocksDelivery) UpdateBlockPosition(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	vars := mux.Vars(r)
	blockID, err := strconv.ParseUint(vars["block_id"], 10, 64)
	if err != nil {
		log.Warn().Err(err).Str("block_id", vars["block_id"]).Msg("invalid block id")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid block id")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close request body")
		}
	}()
	var req UpdateBlockPositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("invalid request body")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	block, err := d.Usecase.UpdateBlockPosition(r.Context(), userID, blockID, req.BeforeBlockID)
	if err != nil {
		handleBlockError(w, r.Context(), err)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, block)
}

func handleBlockError(w http.ResponseWriter, ctx context.Context, err error) {
	log := logger.FromContext(ctx)
	switch err {
	case namederrors.ErrNotFound:
		log.Warn().Err(err).Msg("block or note not found")
		apiutils.WriteError(w, http.StatusNotFound, "block or note not found")
	case namederrors.ErrNoAccess:
		log.Warn().Err(err).Msg("access to note denied")
		apiutils.WriteError(w, http.StatusForbidden, "no access to this note")
	default:
		log.Error().Err(err).Msg("internal server error in blocks delivery")
		apiutils.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}

func convertToFormats(inputs []BlockTextFormatInput) []models.BlockTextFormat {
	formats := make([]models.BlockTextFormat, len(inputs))
	for i, input := range inputs {
		formats[i] = models.BlockTextFormat{
			StartOffset:   input.StartOffset,
			EndOffset:     input.EndOffset,
			Bold:          apiutils.GetBool(input.Bold, false),
			Italic:        apiutils.GetBool(input.Italic, false),
			Underline:     apiutils.GetBool(input.Underline, false),
			Strikethrough: apiutils.GetBool(input.Strikethrough, false),
			Link:          input.Link,
			Font:          models.TextFont(apiutils.GetString(input.Font, string(models.FontInter))),
			Size:          apiutils.GetInt(input.Size, 12),
		}
	}
	return formats
}