package delivery

import (
	"backend/apiutils"
	"backend/logger"
	"backend/middleware"
	"backend/models"
	namederrors "backend/named_errors"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

//go:generate mockgen -source=delivery.go -destination=../mock/mock_delivery.go -package=mock
type BlocksUsecase interface {
	GetBlocks(ctx context.Context, userID, noteID uint64) ([]models.Block, error)
	GetBlock(ctx context.Context, userID, blockID uint64) (*models.Block, error)
	UpdateBlock(ctx context.Context, userID uint64, req *models.UpdateBlockRequest) (*models.Block, error)
	CreateTextBlock(ctx context.Context, userID, noteID uint64, beforeBlockID *uint64) (*models.Block, error)
	CreateCodeBlock(ctx context.Context, userID, noteID uint64, beforeBlockID *uint64) (*models.Block, error)
	CreateAttachmentBlock(ctx context.Context, userID, noteID uint64, beforeBlockID *uint64, fileID uint64) (*models.Block, error)
	DeleteBlock(ctx context.Context, userID, blockID uint64) error
	UpdateBlockPosition(ctx context.Context, userID, blockID uint64, beforeBlockID *uint64) (*models.Block, error)
}

type BlocksDelivery struct {
	Usecase BlocksUsecase
}

func NewBlocksDelivery(usecase BlocksUsecase) *BlocksDelivery {
	return &BlocksDelivery{
		Usecase: usecase,
	}
}

type baseCreateBlockRequest struct {
	Type          string  `json:"type"`
	BeforeBlockID *uint64 `json:"before_block_id,omitempty"`
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
			log.Error().Err(err).Msg("failed to close body")
		}
	}()

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		log.Warn().Err(err).Msg("failed to read body")
		apiutils.WriteError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	var base baseCreateBlockRequest
	if err := json.Unmarshal(body, &base); err != nil {
		log.Warn().Err(err).Msg("invalid request (type)")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	switch base.Type {
	case models.BlockTypeText:
		d.createTextBlock(w, r, userID, noteID, body)
	case models.BlockTypeAttachment:
		d.createAttachmentBlock(w, r, userID, noteID, body)
	case models.BlockTypeCode:
		d.createCodeBlock(w, r, userID, noteID, body)
	default:
		apiutils.WriteError(w, http.StatusBadRequest, "unsupported block type")
	}
}

type createCodeBlockRequest struct {
	baseCreateBlockRequest
}

func (d *BlocksDelivery) createCodeBlock(w http.ResponseWriter, r *http.Request, userID, noteID uint64, body []byte) {
	log := logger.FromContext(r.Context())
	log.Info().Uint64("user_id", userID).Uint64("note_id", noteID).Msg("handling create code block request")

	var req createCodeBlockRequest
	if err := apiutils.StrictUnmarshal(body, &req); err != nil {
		log.Warn().Err(err).Msg("invalid payload for code block")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid payload for code block")
		return
	}

	block, err := d.Usecase.CreateCodeBlock(r.Context(), userID, noteID, req.BeforeBlockID)
	if err != nil {
		handleBlockError(w, r.Context(), err)
		return
	}
	apiutils.WriteJSON(w, http.StatusCreated, block)
}

type createTextBlockRequest struct {
	baseCreateBlockRequest
}

func (d *BlocksDelivery) createTextBlock(w http.ResponseWriter, r *http.Request, userID, noteID uint64, body []byte) {
	log := logger.FromContext(r.Context())

	var req createTextBlockRequest
	if err := apiutils.StrictUnmarshal(body, &req); err != nil {
		log.Warn().Err(err).Msg("invalid payload for text block")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid payload for text block")
		return
	}

	block, err := d.Usecase.CreateTextBlock(r.Context(), userID, noteID, req.BeforeBlockID)
	if err != nil {
		handleBlockError(w, r.Context(), err)
		return
	}
	apiutils.WriteJSON(w, http.StatusCreated, block)
}

type createAttachmentBlockRequest struct {
	baseCreateBlockRequest
	FileID uint64 `json:"file_id"`
}

func (d *BlocksDelivery) createAttachmentBlock(w http.ResponseWriter, r *http.Request, userID, noteID uint64, body []byte) {
	log := logger.FromContext(r.Context())

	var req createAttachmentBlockRequest
	if err := apiutils.StrictUnmarshal(body, &req); err != nil {
		log.Warn().Err(err).Msg("invalid payload for attachment block")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid payload for attachment block")
		return
	}

	if req.FileID == 0 {
		log.Warn().Msg("file_id is required for attachment block")
		apiutils.WriteError(w, http.StatusBadRequest, "file_id is required")
		return
	}

	block, err := d.Usecase.CreateAttachmentBlock(r.Context(), userID, noteID, req.BeforeBlockID, req.FileID)
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

type UpdateBlockDeliveryRequest struct {
	Type    string          `json:"type"`
	Content json.RawMessage `json:"content"`
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

	var deliveryReq UpdateBlockDeliveryRequest
	if err := json.NewDecoder(r.Body).Decode(&deliveryReq); err != nil {
		log.Warn().Err(err).Msg("invalid request body")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	usecaseReq, err := parseUpdateBlockRequest(blockID, deliveryReq)
	if err != nil {
		log.Warn().Err(err).Msg("failed to parse request")
		apiutils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	block, err := d.Usecase.UpdateBlock(r.Context(), userID, usecaseReq)
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
	switch {
	case errors.Is(err, namederrors.ErrNotFound):
		log.Warn().Err(err).Msg("block or note not found")
		apiutils.WriteError(w, http.StatusNotFound, "block or note not found")
	case errors.Is(err, namederrors.ErrNoAccess):
		log.Warn().Err(err).Msg("access to note denied")
		apiutils.WriteError(w, http.StatusForbidden, "no access to this note")
	default:
		log.Error().Err(err).Msg("internal server error in blocks delivery")
		apiutils.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}

func parseUpdateBlockRequest(blockID uint64, deliveryReq UpdateBlockDeliveryRequest) (*models.UpdateBlockRequest, error) {
	req := &models.UpdateBlockRequest{
		BlockID: blockID,
		Type:    deliveryReq.Type,
	}

	switch deliveryReq.Type {
	case models.BlockTypeText:
		var textContent models.UpdateTextContent
		if err := json.Unmarshal(deliveryReq.Content, &textContent); err != nil {
			return nil, fmt.Errorf("failed to unmarshal text content: %w", err)
		}
		req.Content = textContent

	case models.BlockTypeCode:
		var codeContent models.UpdateCodeContent
		if err := json.Unmarshal(deliveryReq.Content, &codeContent); err != nil {
			return nil, fmt.Errorf("failed to unmarshal code content: %w", err)
		}
		req.Content = codeContent

	default:
		return nil, fmt.Errorf("unsupported block type: %s", deliveryReq.Type)
	}

	return req, nil
}
