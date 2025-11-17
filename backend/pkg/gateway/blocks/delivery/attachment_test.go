package delivery

import (
	"backend/middleware"
	"backend/models"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

type fakeBlocksUsecase2 struct{
    called bool
}
func (f *fakeBlocksUsecase2) GetBlocks(ctx context.Context, userID, noteID uint64) ([]models.BlockWithContent, error) { return nil, nil }
func (f *fakeBlocksUsecase2) GetBlock(ctx context.Context, userID, blockID uint64) (*models.BlockWithContent, error) { return nil, nil }
func (f *fakeBlocksUsecase2) UpdateBlock(ctx context.Context, userID, blockID uint64, text string, formats []models.BlockTextFormat) (*models.BlockWithContent, error) { return nil, nil }
func (f *fakeBlocksUsecase2) DeleteBlock(ctx context.Context, userID, blockID uint64) error { return nil }
func (f *fakeBlocksUsecase2) UpdateBlockPosition(ctx context.Context, userID, blockID uint64, afterBlockID *uint64) (*models.Block, error) { return nil, nil }
func (f *fakeBlocksUsecase2) CreateTextBlock(ctx context.Context, userID, noteID uint64, beforeBlockID *uint64) (*models.BlockWithContent, error) { return nil, nil }
func (f *fakeBlocksUsecase2) CreateAttachmentBlock(ctx context.Context, userID, noteID uint64, beforeBlockID *uint64, fileID uint64) (*models.BlockWithContent, error) { f.called = true; return &models.BlockWithContent{Block: models.Block{ID: 1, NoteID: noteID}}, nil }

func TestCreateAttachmentBlock_MissingFileID(t *testing.T) {
    d := &BlocksDelivery{Usecase: &fakeBlocksUsecase2{}}

    body := `{"type":"attachment"}`
    req := httptest.NewRequest("POST", "/api/notes/1/blocks", strings.NewReader(body))
    req = mux.SetURLVars(req, map[string]string{"note_id": "1"})
    // set user id
    req = req.WithContext(middleware.WithUserID(req.Context(), 1))

    w := httptest.NewRecorder()
    d.CreateBlock(w, req)

    if w.Result().StatusCode != http.StatusBadRequest {
        t.Fatalf("expected 400 got %d body=%s", w.Result().StatusCode, w.Body.String())
    }
}

func TestCreateAttachmentBlock_Success(t *testing.T) {
    fu := &fakeBlocksUsecase2{}
    d := &BlocksDelivery{Usecase: fu}

    body := `{"type":"attachment","file_id":5}`
    req := httptest.NewRequest("POST", "/api/notes/2/blocks", strings.NewReader(body))
    req = mux.SetURLVars(req, map[string]string{"note_id": "2"})
    req = req.WithContext(middleware.WithUserID(req.Context(), 7))

    w := httptest.NewRecorder()
    d.CreateBlock(w, req)

    if w.Result().StatusCode != http.StatusCreated {
        t.Fatalf("expected 201 got %d body=%s", w.Result().StatusCode, w.Body.String())
    }
    if !fu.called {
        t.Fatalf("usecase was not called")
    }
}
