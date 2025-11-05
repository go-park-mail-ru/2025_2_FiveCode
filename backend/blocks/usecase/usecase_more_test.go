package usecase

import (
	"backend/blocks/repository"
	namederrors "backend/named_errors"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// reuse fakeBlocksRepo and fakeNotesRepo from helpers test

func Test_CreateAttachmentBlock_fileIDRequired(t *testing.T) {
    fb := &fakeBlocksRepo{blocks: []repository.BlockPositionInfo{}}
    fn := &fakeNotesRepo{noteOwner: 1}
    u := NewBlocksUsecase(fb, fn)
    _, err := u.CreateAttachmentBlock(context.Background(), 1, 1, nil, 0)
    assert.Error(t, err)
}

func Test_checkNoteAccess_denied(t *testing.T) {
    fb := &fakeBlocksRepo{blocks: []repository.BlockPositionInfo{}}
    // note owner differs
    fn := &fakeNotesRepo{noteOwner: 2}
    u := NewBlocksUsecase(fb, fn)
    err := u.checkNoteAccess(context.Background(), 1, 1)
    assert.Equal(t, namederrors.ErrNoAccess, err)
}

func Test_UpdateBlockPosition_beforeNotFound(t *testing.T) {
    fb := &fakeBlocksRepo{blocks: []repository.BlockPositionInfo{{ID: 2, Position: 2.0}}}
    fn := &fakeNotesRepo{noteOwner: 1}
    u := NewBlocksUsecase(fb, fn)
    // trying to update with beforeBlock that doesn't exist
    nf := uint64(999)
    _, err := u.calculatePosition(context.Background(), 1, &nf, 0)
    assert.Error(t, err)
}
