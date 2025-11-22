package delivery

import (
	notePB "backend/notes_service/pkg/note/v1"
	"backend/gateway_service/internal/apiutils"
	"backend/gateway_service/internal/middleware"
	"backend/gateway_service/internal/utils"
	"backend/gateway_service/logger"
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

//go:generate mockgen -destination=../mock/mock_note_client.go -package=mock . NoteServiceClient
type NoteServiceClient interface {
	GetAllNotes(ctx context.Context, in *notePB.GetAllNotesRequest, opts ...grpc.CallOption) (*notePB.GetAllNotesResponse, error)
	CreateNote(ctx context.Context, in *notePB.CreateNoteRequest, opts ...grpc.CallOption) (*notePB.Note, error)
	GetNoteById(ctx context.Context, in *notePB.GetNoteByIdRequest, opts ...grpc.CallOption) (*notePB.Note, error)
	UpdateNote(ctx context.Context, in *notePB.UpdateNoteRequest, opts ...grpc.CallOption) (*notePB.Note, error)
	DeleteNote(ctx context.Context, in *notePB.DeleteNoteRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	AddFavorite(ctx context.Context, in *notePB.FavoriteRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	RemoveFavorite(ctx context.Context, in *notePB.FavoriteRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type NotesDelivery struct {
	NoteClient NoteServiceClient
}

func NewNotesDelivery(n NoteServiceClient) *NotesDelivery {
	return &NotesDelivery{
		NoteClient: n,
	}
}

func (d *NotesDelivery) GetAllNotes(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusInternalServerError, "user not authenticated")
		return
	}

	grpcResp, err := d.NoteClient.GetAllNotes(r.Context(), &notePB.GetAllNotesRequest{
		UserId: userID,
	})
	if err != nil {
		log.Error().Err(err).Msg("gRPC call to Note service failed")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	notes := utils.ProtoNotesToModels(grpcResp.Notes)
	apiutils.WriteJSON(w, http.StatusOK, notes)
}

func (d *NotesDelivery) CreateNote(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusInternalServerError, "user not authenticated")
		return
	}

	grpcResp, err := d.NoteClient.CreateNote(r.Context(), &notePB.CreateNoteRequest{
		UserId: userID,
	})
	if err != nil {
		log.Error().Err(err).Msg("gRPC call to Note service failed")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	note := utils.ProtoNoteToModel(grpcResp)
	apiutils.WriteJSON(w, http.StatusCreated, note)
}

func (d *NotesDelivery) GetNoteById(w http.ResponseWriter, r *http.Request) {
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
		apiutils.WriteError(w, http.StatusInternalServerError, "user not authenticated")
		return
	}

	grpcResp, err := d.NoteClient.GetNoteById(r.Context(), &notePB.GetNoteByIdRequest{
		UserId: userID,
		NoteId: noteID,
	})
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("gRPC call to Note service failed")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	note := utils.ProtoNoteToModel(grpcResp)
	apiutils.WriteJSON(w, http.StatusOK, note)
}

type UpdateNoteRequest struct {
	Title      *string `json:"title"`
	IsArchived *bool   `json:"is_archived"`
}

func (d *NotesDelivery) UpdateNote(w http.ResponseWriter, r *http.Request) {
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
		apiutils.WriteError(w, http.StatusInternalServerError, "user not authenticated")
		return
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close request body")
		}
	}()

	var req UpdateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("invalid request body")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == nil && req.IsArchived == nil {
		log.Warn().Msg("invalid request body: title and is_archived are both nil")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	grpcReq := &notePB.UpdateNoteRequest{
		UserId: userID,
		NoteId: noteID,
	}
	if req.Title != nil {
		grpcReq.Title = req.Title
	}
	if req.IsArchived != nil {
		grpcReq.IsArchived = req.IsArchived
	}

	grpcResp, err := d.NoteClient.UpdateNote(r.Context(), grpcReq)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("gRPC call to Note service failed")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	note := utils.ProtoNoteToModel(grpcResp)
	apiutils.WriteJSON(w, http.StatusOK, note)
}

func (d *NotesDelivery) DeleteNote(w http.ResponseWriter, r *http.Request) {
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
		apiutils.WriteError(w, http.StatusInternalServerError, "user not authenticated")
		return
	}

	_, err = d.NoteClient.DeleteNote(r.Context(), &notePB.DeleteNoteRequest{
		UserId: userID,
		NoteId: noteID,
	})
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("gRPC call to Note service failed")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, "note was successfully deleted")
}

func (d *NotesDelivery) AddFavorite(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	vars := mux.Vars(r)
	noteID, _ := strconv.ParseUint(vars["note_id"], 10, 64)

	_, err := d.NoteClient.AddFavorite(r.Context(), &notePB.FavoriteRequest{
		UserId: userID,
		NoteId: noteID,
	})
	if err != nil {
		log.Error().Err(err).Msg("gRPC call to Note service failed")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (d *NotesDelivery) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	vars := mux.Vars(r)
	noteID, _ := strconv.ParseUint(vars["note_id"], 10, 64)

	_, err := d.NoteClient.RemoveFavorite(r.Context(), &notePB.FavoriteRequest{
		UserId: userID,
		NoteId: noteID,
	})
	if err != nil {
		log.Error().Err(err).Msg("gRPC call to Note service failed")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
