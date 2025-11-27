package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/notes_service/internal/constants"
	"backend/notes_service/internal/models"
	sharePB "backend/notes_service/pkg/sharing/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) AddCollaborator(ctx context.Context, req *sharePB.AddCollaboratorRequest) (*sharePB.CollaboratorResponse, error) {
	permission, err := s.sharingUsecase.AddCollaborator(
		ctx,
		req.GetNoteId(),
		req.GetCurrentUserId(),
		req.GetUserId(),
		noteRoleFromProto(req.GetRole()),
	)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		if errors.Is(err, constants.ErrSubNoteCannotBeShared) {
			return nil, status.Error(codes.InvalidArgument, "sub-notes cannot be shared")
		}
		return nil, status.Error(codes.Internal, "failed to add collaborator")
	}

	collaborator := &models.Collaborator{
		PermissionID: permission.PermissionID,
		UserID:       permission.GrantedTo,
		Role:         permission.Role,
		GrantedBy:    permission.GrantedBy,
		GrantedAt:    permission.CreatedAt,
	}

	return &sharePB.CollaboratorResponse{
		PermissionId: permission.PermissionID,
		Collaborator: collaboratorModelToProto(collaborator),
	}, nil
}

func (s *Server) GetCollaborators(ctx context.Context, req *sharePB.GetCollaboratorsRequest) (*sharePB.GetCollaboratorsResponse, error) {
	ownerID, permissions, publicAccessLevel, err := s.sharingUsecase.GetCollaborators(
		ctx,
		req.GetNoteId(),
		req.GetCurrentUserId(),
	)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to get collaborators")
	}

	collaborators := make([]*sharePB.Collaborator, len(permissions))
	for i, p := range permissions {
		collaborators[i] = &sharePB.Collaborator{
			PermissionId: p.PermissionID,
			UserId:       p.GrantedTo,
			Role:         noteRoleToProto(p.Role),
			GrantedBy:    p.GrantedBy,
			GrantedAt:    timestamppb.New(p.CreatedAt),
		}
	}

	var protoPublicAccessLevel *sharePB.NoteRole
	if publicAccessLevel != nil {
		level := noteRoleToProto(*publicAccessLevel)
		protoPublicAccessLevel = &level
	}

	return &sharePB.GetCollaboratorsResponse{
		NoteId:             req.GetNoteId(),
		OwnerId:            ownerID,
		Collaborators:      collaborators,
		PublicAccessLevel:  protoPublicAccessLevel,
		TotalCollaborators: int32(len(collaborators)),
	}, nil
}

func (s *Server) UpdateCollaboratorRole(ctx context.Context, req *sharePB.UpdateCollaboratorRoleRequest) (*sharePB.CollaboratorResponse, error) {
	permission, err := s.sharingUsecase.UpdateCollaboratorRole(
		ctx,
		req.GetNoteId(),
		req.GetCurrentUserId(),
		req.GetPermissionId(),
		noteRoleFromProto(req.GetNewRole()),
	)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "permission not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		// ← ДОБАВИТЬ
		if errors.Is(err, constants.ErrSubNoteCannotBeShared) {
			return nil, status.Error(codes.InvalidArgument, "sub-notes cannot be shared")
		}
		return nil, status.Error(codes.Internal, "failed to update collaborator role")
	}

	collaborator := &models.Collaborator{
		PermissionID: permission.PermissionID,
		UserID:       permission.GrantedTo,
		Role:         permission.Role,
		GrantedBy:    permission.GrantedBy,
		GrantedAt:    permission.CreatedAt,
	}

	return &sharePB.CollaboratorResponse{
		PermissionId: permission.PermissionID,
		Collaborator: collaboratorModelToProto(collaborator),
	}, nil
}

func (s *Server) RemoveCollaborator(ctx context.Context, req *sharePB.RemoveCollaboratorRequest) (*emptypb.Empty, error) {
	err := s.sharingUsecase.RemoveCollaborator(
		ctx,
		req.GetNoteId(),
		req.GetCurrentUserId(),
		req.GetPermissionId(),
	)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "permission not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		if errors.Is(err, constants.ErrSubNoteCannotBeShared) {
			return nil, status.Error(codes.InvalidArgument, "sub-notes cannot be shared")
		}
		return nil, status.Error(codes.Internal, "failed to remove collaborator")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) SetPublicAccess(ctx context.Context, req *sharePB.SetPublicAccessRequest) (*sharePB.PublicAccessResponse, error) {
	var accessLevel *models.NoteRole
	if req.AccessLevel != nil {
		level := noteRoleFromProto(*req.AccessLevel)
		accessLevel = &level
	}

	err := s.sharingUsecase.SetPublicAccess(
		ctx,
		req.GetNoteId(),
		req.GetCurrentUserId(),
		accessLevel,
	)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		if errors.Is(err, constants.ErrSubNoteCannotBeShared) {
			return nil, status.Error(codes.InvalidArgument, "sub-notes cannot be shared")
		}
		return nil, status.Error(codes.Internal, "failed to set public access")
	}

	updatedAccess, err := s.sharingUsecase.GetPublicAccess(ctx, req.GetNoteId(), req.GetCurrentUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get updated public access")
	}

	var protoAccessLevel *sharePB.NoteRole
	if updatedAccess != nil {
		level := noteRoleToProto(*updatedAccess)
		protoAccessLevel = &level
	}

	return &sharePB.PublicAccessResponse{
		NoteId:      req.GetNoteId(),
		AccessLevel: protoAccessLevel,
		ShareUrl:    fmt.Sprintf("/notes/%d", req.GetNoteId()),
		UpdatedAt:   timestamppb.New(time.Now()),
	}, nil
}

func (s *Server) GetPublicAccess(ctx context.Context, req *sharePB.GetPublicAccessRequest) (*sharePB.PublicAccessResponse, error) {
	accessLevel, err := s.sharingUsecase.GetPublicAccess(
		ctx,
		req.GetNoteId(),
		req.GetCurrentUserId(),
	)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to get public access")
	}

	var protoAccessLevel *sharePB.NoteRole
	if accessLevel != nil {
		level := noteRoleToProto(*accessLevel)
		protoAccessLevel = &level
	}

	return &sharePB.PublicAccessResponse{
		NoteId:      req.GetNoteId(),
		AccessLevel: protoAccessLevel,
		ShareUrl:    fmt.Sprintf("/notes/%d", req.GetNoteId()),
		UpdatedAt:   timestamppb.New(time.Now()),
	}, nil
}

func (s *Server) ActivateAccessByLink(ctx context.Context, req *sharePB.ActivateAccessByLinkRequest) (*sharePB.ActivateAccessByLinkResponse, error) {
	response, err := s.sharingUsecase.ActivateAccessByLink(
		ctx,
		req.GetShareUuid(),
		req.GetUserId(),
	)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		return nil, status.Error(codes.Internal, "failed to activate access by link")
	}

	return &sharePB.ActivateAccessByLinkResponse{
		NoteId:        response.NoteID,
		AccessGranted: response.AccessGranted,
		AccessInfo:    noteAccessInfoModelToProto(&response.AccessInfo),
	}, nil
}

func (s *Server) GetSharingSettings(ctx context.Context, req *sharePB.GetSharingSettingsRequest) (*sharePB.SharingSettingsResponse, error) {
	settings, err := s.sharingUsecase.GetSharingSettings(
		ctx,
		req.GetNoteId(),
		req.GetCurrentUserId(),
	)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "note not found")
		}
		if errors.Is(err, constants.ErrNoAccess) {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
		return nil, status.Error(codes.Internal, "failed to get sharing settings")
	}

	collaborators := make([]*sharePB.Collaborator, len(settings.Collaborators))
	for i, c := range settings.Collaborators {
		collaborators[i] = &sharePB.Collaborator{
			PermissionId: c.PermissionID,
			UserId:       c.UserID,
			Role:         noteRoleToProto(c.Role),
			GrantedBy:    c.GrantedBy,
			GrantedAt:    timestamppb.New(c.GrantedAt),
		}
	}

	var publicAccessLevel *sharePB.NoteRole
	if settings.PublicAccess.AccessLevel != nil {
		level := noteRoleToProto(*settings.PublicAccess.AccessLevel)
		publicAccessLevel = &level
	}

	publicAccess := &sharePB.PublicAccess{
		NoteId:      settings.PublicAccess.NoteID,
		AccessLevel: publicAccessLevel,
		ShareUrl:    settings.PublicAccess.ShareURL,
	}

	return &sharePB.SharingSettingsResponse{
		NoteId:             settings.NoteID,
		OwnerId:            settings.Owner.UserID,
		PublicAccess:       publicAccess,
		Collaborators:      collaborators,
		TotalCollaborators: int32(settings.TotalCount),
		IsOwner:            settings.IsOwner,
	}, nil
}

func (s *Server) CheckNoteAccess(ctx context.Context, req *sharePB.CheckNoteAccessRequest) (*sharePB.NoteAccessResponse, error) {
	accessInfo, err := s.sharingUsecase.CheckNoteAccess(
		ctx,
		req.GetNoteId(),
		req.GetUserId(),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to check note access")
	}

	return noteAccessInfoModelToProto(accessInfo), nil
}
