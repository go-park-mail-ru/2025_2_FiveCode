package delivery

import (
	"backend/gateway_service/internal/apiutils"
	"backend/gateway_service/internal/middleware"
	"backend/gateway_service/internal/utils"
	"backend/gateway_service/logger"
	"backend/notes_service/models"
	blockPB "backend/notes_service/pkg/block/v1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type baseCreateBlockRequest struct {
	Type          string  `json:"type"`
	BeforeBlockID *uint64 `json:"before_block_id,omitempty"`
}

func (d *NotesDelivery) CreateBlock(w http.ResponseWriter, r *http.Request) {
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

func (d *NotesDelivery) createCodeBlock(w http.ResponseWriter, r *http.Request, userID, noteID uint64, body []byte) {
	log := logger.FromContext(r.Context())

	var req createCodeBlockRequest
	if err := apiutils.StrictUnmarshal(body, &req); err != nil {
		log.Warn().Err(err).Msg("invalid payload for code block")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid payload for code block")
		return
	}

	grpcResp, err := d.usecase.CreateCodeBlock(r.Context(), &blockPB.CreateCodeBlockRequest{
		UserId:        userID,
		NoteId:        noteID,
		BeforeBlockId: req.BeforeBlockID,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create code block")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	block := utils.ProtoBlockToModel(grpcResp)
	apiutils.WriteJSON(w, http.StatusCreated, block)
}

type createTextBlockRequest struct {
	baseCreateBlockRequest
}

func (d *NotesDelivery) createTextBlock(w http.ResponseWriter, r *http.Request, userID, noteID uint64, body []byte) {
	log := logger.FromContext(r.Context())

	var req createTextBlockRequest
	if err := apiutils.StrictUnmarshal(body, &req); err != nil {
		log.Warn().Err(err).Msg("invalid payload for text block")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid payload for text block")
		return
	}

	grpcResp, err := d.usecase.CreateTextBlock(r.Context(), &blockPB.CreateTextBlockRequest{
		UserId:        userID,
		NoteId:        noteID,
		BeforeBlockId: req.BeforeBlockID,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create text block")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	block := utils.ProtoBlockToModel(grpcResp)
	apiutils.WriteJSON(w, http.StatusCreated, block)
}

type createAttachmentBlockRequest struct {
	baseCreateBlockRequest
	FileID uint64 `json:"file_id"`
}

func (d *NotesDelivery) createAttachmentBlock(w http.ResponseWriter, r *http.Request, userID, noteID uint64, body []byte) {
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

	grpcResp, err := d.usecase.CreateAttachmentBlock(r.Context(), &blockPB.CreateAttachmentBlockRequest{
		UserId:        userID,
		NoteId:        noteID,
		BeforeBlockId: req.BeforeBlockID,
		FileId:        req.FileID,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create attachment block")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	block := utils.ProtoBlockToModel(grpcResp)
	apiutils.WriteJSON(w, http.StatusCreated, block)
}

func (d *NotesDelivery) GetBlocks(w http.ResponseWriter, r *http.Request) {
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

	grpcResp, err := d.usecase.GetBlocks(r.Context(), userID, noteID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get blocks")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	blocks := utils.ProtoBlocksToModels(grpcResp.Blocks)

	apiutils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"note_id": noteID,
		"blocks":  blocks,
	})
}

func (d *NotesDelivery) GetBlock(w http.ResponseWriter, r *http.Request) {
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

	grpcResp, err := d.usecase.GetBlock(r.Context(), userID, blockID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get block")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	block := utils.ProtoBlockToModel(grpcResp)
	apiutils.WriteJSON(w, http.StatusOK, block)
}

type UpdateBlockDeliveryRequest struct {
	Type    string          `json:"type"`
	Content json.RawMessage `json:"content"`
}

func (d *NotesDelivery) UpdateBlock(w http.ResponseWriter, r *http.Request) {
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

	grpcReq, err := convertToGrpcUpdateBlockRequest(userID, blockID, deliveryReq)
	if err != nil {
		log.Warn().Err(err).Msg("failed to convert request")
		apiutils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	grpcResp, err := d.usecase.UpdateBlock(r.Context(), grpcReq)
	if err != nil {
		log.Error().Err(err).Msg("failed to update block")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	block := utils.ProtoBlockToModel(grpcResp)
	apiutils.WriteJSON(w, http.StatusOK, block)
}

func (d *NotesDelivery) DeleteBlock(w http.ResponseWriter, r *http.Request) {
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

	err = d.usecase.DeleteBlock(r.Context(), userID, blockID)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete block")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type UpdateBlockPositionRequest struct {
	BeforeBlockID *uint64 `json:"before_block_id"`
}

func (d *NotesDelivery) UpdateBlockPosition(w http.ResponseWriter, r *http.Request) {
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

	grpcResp, err := d.usecase.UpdateBlockPosition(r.Context(), &blockPB.UpdateBlockPositionRequest{
		UserId:        userID,
		BlockId:       blockID,
		BeforeBlockId: req.BeforeBlockID,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to update block position")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	block := utils.ProtoBlockToModel(grpcResp)
	apiutils.WriteJSON(w, http.StatusOK, block)
}

func convertToGrpcUpdateBlockRequest(userID, blockID uint64, deliveryReq UpdateBlockDeliveryRequest) (*blockPB.UpdateBlockRequest, error) {
	grpcReq := &blockPB.UpdateBlockRequest{
		UserId:  userID,
		BlockId: blockID,
		Type:    deliveryReq.Type,
	}

	switch deliveryReq.Type {
	case models.BlockTypeText:
		var textContent models.UpdateTextContent
		if err := json.Unmarshal(deliveryReq.Content, &textContent); err != nil {
			return nil, fmt.Errorf("failed to unmarshal text content: %w", err)
		}
		grpcReq.Content = &blockPB.UpdateBlockRequest_TextContent{
			TextContent: utils.ModelTextContentToProto(&textContent),
		}

	case models.BlockTypeCode:
		var codeContent models.UpdateCodeContent
		if err := json.Unmarshal(deliveryReq.Content, &codeContent); err != nil {
			return nil, fmt.Errorf("failed to unmarshal code content: %w", err)
		}
		grpcReq.Content = &blockPB.UpdateBlockRequest_CodeContent{
			CodeContent: utils.ModelCodeContentToProto(&codeContent),
		}

	default:
		return nil, fmt.Errorf("unsupported block type: %s", deliveryReq.Type)
	}

	return grpcReq, nil
}
