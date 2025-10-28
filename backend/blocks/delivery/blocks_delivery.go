package blocksDelivery

import (
	"backend/apiutils"
	"backend/middleware"
	"backend/models"
	namederrors "backend/named_errors"
	"context"
	"encoding/json"
	"io"
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
	vars := mux.Vars(r)
	noteID, err := strconv.ParseUint(vars["note_id"], 10, 64)
	if err != nil {
		apiutils.WriteError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	var req CreateBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiutils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	block, err := d.Usecase.CreateBlock(r.Context(), userID, noteID, req.BeforeBlockID)
	if err != nil {
		handleBlockError(w, err)
		return
	}

	apiutils.WriteJSON(w, http.StatusCreated, block)
}

func (d *BlocksDelivery) GetBlocks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteID, err := strconv.ParseUint(vars["note_id"], 10, 64)
	if err != nil {
		apiutils.WriteError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	blocks, err := d.Usecase.GetBlocks(r.Context(), userID, noteID)
	if err != nil {
		handleBlockError(w, err)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"note_id": noteID,
		"blocks":  blocks,
	})
}

func (d *BlocksDelivery) GetBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	blockID, err := strconv.ParseUint(vars["block_id"], 10, 64)
	if err != nil {
		apiutils.WriteError(w, http.StatusBadRequest, "invalid block id")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	block, err := d.Usecase.GetBlock(r.Context(), userID, blockID)
	if err != nil {
		handleBlockError(w, err)
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
	vars := mux.Vars(r)
	blockID, err := strconv.ParseUint(vars["block_id"], 10, 64)
	if err != nil {
		apiutils.WriteError(w, http.StatusBadRequest, "invalid block id")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	var req UpdateBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiutils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	formats := convertToFormats(req.Formats)

	block, err := d.Usecase.UpdateBlock(r.Context(), userID, blockID, req.Text, formats)
	if err != nil {
		handleBlockError(w, err)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, block)
}

func (d *BlocksDelivery) DeleteBlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	blockID, err := strconv.ParseUint(vars["block_id"], 10, 64)
	if err != nil {
		apiutils.WriteError(w, http.StatusBadRequest, "invalid block id")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	err = d.Usecase.DeleteBlock(r.Context(), userID, blockID)
	if err != nil {
		handleBlockError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type UpdateBlockPositionRequest struct {
	BeforeBlockID *uint64 `json:"before_block_id"`
}

func (d *BlocksDelivery) UpdateBlockPosition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	blockID, err := strconv.ParseUint(vars["block_id"], 10, 64)
	if err != nil {
		apiutils.WriteError(w, http.StatusBadRequest, "invalid block id")
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	var req UpdateBlockPositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiutils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	block, err := d.Usecase.UpdateBlockPosition(r.Context(), userID, blockID, req.BeforeBlockID)
	if err != nil {
		handleBlockError(w, err)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, block)
}

func handleBlockError(w http.ResponseWriter, err error) {
	switch err {
	case namederrors.ErrNotFound:
		apiutils.WriteError(w, http.StatusNotFound, "block or note not found")
	case namederrors.ErrNoAccess:
		apiutils.WriteError(w, http.StatusForbidden, "no access to this note")
	default:
		apiutils.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}

func convertToFormats(inputs []BlockTextFormatInput) []models.BlockTextFormat {
	formats := make([]models.BlockTextFormat, len(inputs))
	for i, input := range inputs {
		formats[i] = models.BlockTextFormat{
			StartOffset:   input.StartOffset,
			EndOffset:     input.EndOffset,
			Bold:          getBool(input.Bold, false),
			Italic:        getBool(input.Italic, false),
			Underline:     getBool(input.Underline, false),
			Strikethrough: getBool(input.Strikethrough, false),
			Link:          input.Link,
			Font:          models.TextFont(getString(input.Font, string(models.FontInter))),
			Size:          getInt(input.Size, 12),
		}
	}
	return formats
}

func getBool(val *bool, def bool) bool {
	if val != nil {
		return *val
	}
	return def
}

func getString(val *string, def string) string {
	if val != nil {
		return *val
	}
	return def
}

func getInt(val *int, def int) int {
	if val != nil {
		return *val
	}
	return def
}
