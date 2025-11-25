package websocket

import (
	"context"
	"net/http"
	"strconv"

	"backend/gateway_service/internal/middleware"
	sharePB "backend/notes_service/pkg/sharing/v1"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	hub           *Hub
	logger        *zerolog.Logger
	sharingClient sharePB.SharingServiceClient
}

func NewHandler(hub *Hub, logger *zerolog.Logger, sharingClient sharePB.SharingServiceClient) *Handler {
	return &Handler{
		hub:           hub,
		logger:        logger,
		sharingClient: sharingClient,
	}
}

func (h *Handler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	noteIDStr := vars["note_id"]
	noteID, err := strconv.Atoi(noteIDStr)
	if err != nil {
		h.logger.Error().Err(err).Str("note_id", noteIDStr).Msg("invalid note_id")
		http.Error(w, "Invalid note_id", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		h.logger.Error().Msg("user not authenticated")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	hasAccess, err := h.checkNoteAccess(r.Context(), uint64(noteID), userID)
	if err != nil {
		h.logger.Error().Err(err).Uint64("note_id", uint64(noteID)).Uint64("user_id", userID).Msg("failed to check note access")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !hasAccess {
		h.logger.Warn().Uint64("note_id", uint64(noteID)).Uint64("user_id", userID).Msg("user has no access to note")
		http.Error(w, "Forbidden: you don't have access to this note", http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to upgrade connection")
		return
	}

	client := NewClient(conn, h.hub, int(userID), h.logger)
	client.NoteID = noteID

	h.hub.register <- client
	h.hub.subscribe <- client

	h.logger.Info().
		Uint64("user_id", userID).
		Int("note_id", noteID).
		Msg("websocket connection established and subscribed")

	go client.writePump()
	go client.readPump()
}

func (h *Handler) checkNoteAccess(ctx context.Context, noteID, userID uint64) (bool, error) {
	req := &sharePB.CheckNoteAccessRequest{
		NoteId: noteID,
		UserId: userID,
	}

	resp, err := h.sharingClient.CheckNoteAccess(ctx, req)
	if err != nil {
		return false, err
	}

	return resp.HasAccess, nil
}
