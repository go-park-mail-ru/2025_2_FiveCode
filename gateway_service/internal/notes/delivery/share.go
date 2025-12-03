package delivery

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/gateway_service/internal/apiutils"
	"backend/gateway_service/internal/middleware"
	"backend/gateway_service/internal/notes/models"
	"backend/pkg/logger"

	"github.com/gorilla/mux"
)

type AddCollaboratorRequest struct {
	Email string          `json:"email"`
	Role  models.NoteRole `json:"role"`
}

func (d *NotesDelivery) AddCollaborator(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	currentUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	vars := mux.Vars(r)
	noteID, err := strconv.ParseUint(vars["note_id"], 10, 64)
	if err != nil {
		log.Warn().Err(err).Str("note_id", vars["note_id"]).Msg("invalid note id")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close request body")
		}
	}()

	var req AddCollaboratorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("invalid request body")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" {
		log.Warn().Msg("email is required")
		apiutils.WriteError(w, http.StatusBadRequest, "email is required")
		return
	}

	input := &models.AddCollaboratorInput{
		CurrentUserID: currentUserID,
		NoteID:        noteID,
		Email:         req.Email,
		Role:          req.Role,
	}

	response, err := d.usecase.AddCollaborator(r.Context(), input)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Str("email", req.Email).Msg("failed to add collaborator")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	apiutils.WriteJSON(w, http.StatusCreated, response)
}

func (d *NotesDelivery) GetCollaborators(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	currentUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	vars := mux.Vars(r)
	noteID, err := strconv.ParseUint(vars["note_id"], 10, 64)
	if err != nil {
		log.Warn().Err(err).Str("note_id", vars["note_id"]).Msg("invalid note id")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	response, err := d.usecase.GetCollaborators(r.Context(), currentUserID, noteID)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("failed to get collaborators")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, response)
}

type UpdateCollaboratorRoleRequest struct {
	Role models.NoteRole `json:"role"`
}

func (d *NotesDelivery) UpdateCollaboratorRole(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	currentUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	vars := mux.Vars(r)
	noteID, err := strconv.ParseUint(vars["note_id"], 10, 64)
	if err != nil {
		log.Warn().Err(err).Str("note_id", vars["note_id"]).Msg("invalid note id")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	permissionID, err := strconv.ParseUint(vars["permission_id"], 10, 64)
	if err != nil {
		log.Warn().Err(err).Str("permission_id", vars["permission_id"]).Msg("invalid permission id")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid permission id")
		return
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close request body")
		}
	}()

	var req UpdateCollaboratorRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("invalid request body")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := &models.UpdateCollaboratorRoleInput{
		CurrentUserID: currentUserID,
		NoteID:        noteID,
		PermissionID:  permissionID,
		NewRole:       req.Role,
	}

	response, err := d.usecase.UpdateCollaboratorRole(r.Context(), input)
	if err != nil {
		log.Error().Err(err).Uint64("permission_id", permissionID).Msg("failed to update collaborator role")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, response)
}

func (d *NotesDelivery) RemoveCollaborator(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	currentUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	vars := mux.Vars(r)
	noteID, err := strconv.ParseUint(vars["note_id"], 10, 64)
	if err != nil {
		log.Warn().Err(err).Str("note_id", vars["note_id"]).Msg("invalid note id")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	permissionID, err := strconv.ParseUint(vars["permission_id"], 10, 64)
	if err != nil {
		log.Warn().Err(err).Str("permission_id", vars["permission_id"]).Msg("invalid permission id")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid permission id")
		return
	}

	err = d.usecase.RemoveCollaborator(r.Context(), currentUserID, noteID, permissionID)
	if err != nil {
		log.Error().Err(err).Uint64("permission_id", permissionID).Msg("failed to remove collaborator")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type SetPublicAccessRequest struct {
	AccessLevel *models.NoteRole `json:"access_level"`
}

func (d *NotesDelivery) SetPublicAccess(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	currentUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	vars := mux.Vars(r)
	noteID, err := strconv.ParseUint(vars["note_id"], 10, 64)
	if err != nil {
		log.Warn().Err(err).Str("note_id", vars["note_id"]).Msg("invalid note id")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close request body")
		}
	}()

	var req SetPublicAccessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("invalid request body")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := &models.SetPublicAccessInput{
		CurrentUserID: currentUserID,
		NoteID:        noteID,
		AccessLevel:   req.AccessLevel,
	}

	response, err := d.usecase.SetPublicAccess(r.Context(), input)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("failed to set public access")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, response)
}

func (d *NotesDelivery) GetPublicAccess(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	currentUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	vars := mux.Vars(r)
	noteID, err := strconv.ParseUint(vars["note_id"], 10, 64)
	if err != nil {
		log.Warn().Err(err).Str("note_id", vars["note_id"]).Msg("invalid note id")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	response, err := d.usecase.GetPublicAccess(r.Context(), currentUserID, noteID)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("failed to get public access")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, response)
}

func (d *NotesDelivery) GetSharingSettings(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	currentUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	vars := mux.Vars(r)
	noteID, err := strconv.ParseUint(vars["note_id"], 10, 64)
	if err != nil {
		log.Warn().Err(err).Str("note_id", vars["note_id"]).Msg("invalid note id")
		apiutils.WriteError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	response, err := d.usecase.GetSharingSettings(r.Context(), currentUserID, noteID)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("failed to get sharing settings")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, response)
}

func (d *NotesDelivery) ActivateAccessByLink(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		log.Error().Msg("user not authenticated")
		apiutils.WriteError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	vars := mux.Vars(r)
	shareUUID := vars["share_uuid"]
	if shareUUID == "" {
		log.Warn().Msg("share_uuid is required")
		apiutils.WriteError(w, http.StatusBadRequest, "share_uuid is required")
		return
	}

	response, err := d.usecase.ActivateAccessByLink(r.Context(), shareUUID, userID)
	if err != nil {
		log.Error().Err(err).Str("share_uuid", shareUUID).Msg("failed to activate access by link")
		apiutils.HandleGrpcError(w, err, log)
		return
	}

	apiutils.WriteJSON(w, http.StatusOK, response)
}
