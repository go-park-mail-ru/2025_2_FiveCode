package server

import (
	"context"
	"errors"
	"strings"

	"backend/notes_service/internal/constants"
	notePB "backend/notes_service/pkg/note/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) GetAllNotes(ctx context.Context, req *notePB.GetAllNotesRequest) (*notePB.GetAllNotesResponse, error) {
	notes, err := s.noteUsecase.GetAllNotes(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get notes")
	}

	protoNotes := make([]*notePB.Note, len(notes))
	for i := range notes {
		protoNotes[i] = noteModelToProto(&notes[i])
	}

	return &notePB.GetAllNotesResponse{
		Notes: protoNotes,
	}, nil
}

func (s *Server) CreateNote(ctx context.Context, req *notePB.CreateNoteRequest) (*notePB.Note, error) {
	note, err := s.noteUsecase.CreateNote(ctx, req.GetUserId(), req.ParentNoteId)
	if err != nil {
		if strings.Contains(err.Error(), "parent note not found") {
			return nil, status.Error(codes.NotFound, "parent note not found")
		}
		if strings.Contains(err.Error(), "cannot create sub-note of a sub-note") {
			return nil, status.Error(codes.InvalidArgument, "cannot create sub-note of a sub-note: maximum nesting level is 1")
		}
		if strings.Contains(err.Error(), "no access") {
			return nil, status.Error(codes.PermissionDenied, "no access to parent note")
		}
		return nil, status.Error(codes.Internal, "failed to create note")
	}

	return noteModelToProto(note), nil
}

func (s *Server) GetNoteById(ctx context.Context, req *notePB.GetNoteByIdRequest) (*notePB.Note, error) {
	note, err := s.noteUsecase.GetNoteById(ctx, req.GetUserId(), req.GetNoteId())
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to get note")
	}

	return noteModelToProto(note), nil
}

func (s *Server) GetNoteByShareUUID(ctx context.Context, req *notePB.GetNoteByShareUUIDRequest) (*notePB.Note, error) {
	note, err := s.noteUsecase.GetNoteByShareUUID(ctx, req.GetShareUuid())
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		return nil, status.Error(codes.Internal, "failed to get note by share uuid")
	}

	return noteModelToProto(note), nil
}

func (s *Server) UpdateNote(ctx context.Context, req *notePB.UpdateNoteRequest) (*notePB.Note, error) {
	var title *string
	var isArchived *bool

	if req.Title != nil {
		title = req.Title
	}
	if req.IsArchived != nil {
		isArchived = req.IsArchived
	}

	note, err := s.noteUsecase.UpdateNote(ctx, req.GetUserId(), req.GetNoteId(), title, isArchived)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to update note")
	}

	return noteModelToProto(note), nil
}

func (s *Server) DeleteNote(ctx context.Context, req *notePB.DeleteNoteRequest) (*emptypb.Empty, error) {
	err := s.noteUsecase.DeleteNote(ctx, req.GetUserId(), req.GetNoteId())
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to delete note")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) AddFavorite(ctx context.Context, req *notePB.FavoriteRequest) (*emptypb.Empty, error) {
	err := s.noteUsecase.AddFavorite(ctx, req.GetUserId(), req.GetNoteId())
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to add favorite")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) RemoveFavorite(ctx context.Context, req *notePB.FavoriteRequest) (*emptypb.Empty, error) {
	err := s.noteUsecase.RemoveFavorite(ctx, req.GetUserId(), req.GetNoteId())
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to remove favorite")
	}

	return &emptypb.Empty{}, nil
}
