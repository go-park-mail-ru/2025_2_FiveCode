package server

import (
	"context"
	"errors"

	"backend/notes_service/internal/constants"
	"backend/notes_service/internal/models"
	blockPB "backend/notes_service/pkg/block/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) GetBlocks(ctx context.Context, req *blockPB.GetBlocksRequest) (*blockPB.GetBlocksResponse, error) {
	blocks, err := s.blocksUsecase.GetBlocks(ctx, req.GetUserId(), req.GetNoteId())
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to get blocks")
	}

	protoBlocks := make([]*blockPB.Block, len(blocks))
	for i := range blocks {
		protoBlocks[i] = blockModelToProto(&blocks[i])
	}

	return &blockPB.GetBlocksResponse{
		Blocks: protoBlocks,
	}, nil
}

func (s *Server) GetBlock(ctx context.Context, req *blockPB.GetBlockRequest) (*blockPB.Block, error) {
	block, err := s.blocksUsecase.GetBlock(ctx, req.GetUserId(), req.GetBlockId())
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "block not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to get block")
	}

	return blockModelToProto(block), nil
}

func (s *Server) CreateTextBlock(ctx context.Context, req *blockPB.CreateTextBlockRequest) (*blockPB.Block, error) {
	block, err := s.blocksUsecase.CreateTextBlock(ctx, req.GetUserId(), req.GetNoteId(), req.BeforeBlockId)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to create text block")
	}

	return blockModelToProto(block), nil
}

func (s *Server) CreateCodeBlock(ctx context.Context, req *blockPB.CreateCodeBlockRequest) (*blockPB.Block, error) {
	block, err := s.blocksUsecase.CreateCodeBlock(ctx, req.GetUserId(), req.GetNoteId(), req.BeforeBlockId)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to create code block")
	}

	return blockModelToProto(block), nil
}

func (s *Server) CreateAttachmentBlock(ctx context.Context, req *blockPB.CreateAttachmentBlockRequest) (*blockPB.Block, error) {
	block, err := s.blocksUsecase.CreateAttachmentBlock(ctx, req.GetUserId(), req.GetNoteId(), req.BeforeBlockId, req.GetFileId())
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to create attachment block")
	}

	return blockModelToProto(block), nil
}

func (s *Server) UpdateBlock(ctx context.Context, req *blockPB.UpdateBlockRequest) (*blockPB.Block, error) {
	updateReq := &models.UpdateBlockRequest{
		BlockID: req.GetBlockId(),
		Type:    req.GetType(),
	}

	switch content := req.Content.(type) {
	case *blockPB.UpdateBlockRequest_TextContent:
		updateReq.Content = protoToTextContent(content.TextContent)
	case *blockPB.UpdateBlockRequest_CodeContent:
		updateReq.Content = protoToCodeContent(content.CodeContent)
	}

	block, err := s.blocksUsecase.UpdateBlock(ctx, req.GetUserId(), updateReq)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "block not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to update block")
	}

	return blockModelToProto(block), nil
}

func (s *Server) DeleteBlock(ctx context.Context, req *blockPB.DeleteBlockRequest) (*emptypb.Empty, error) {
	err := s.blocksUsecase.DeleteBlock(ctx, req.GetUserId(), req.GetBlockId())
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "block not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to delete block")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) UpdateBlockPosition(ctx context.Context, req *blockPB.UpdateBlockPositionRequest) (*blockPB.Block, error) {
	block, err := s.blocksUsecase.UpdateBlockPosition(ctx, req.GetUserId(), req.GetBlockId(), req.BeforeBlockId)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "block not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to update block position")
	}

	return blockModelToProto(block), nil
}
