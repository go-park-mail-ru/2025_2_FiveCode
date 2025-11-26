package delivery

import (
	"backend/gateway_service/internal/middleware"
	"backend/gateway_service/internal/notes/delivery/mock"
	"backend/gateway_service/internal/notes/models"
	"backend/gateway_service/internal/websocket"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNotesDelivery_AddCollaborator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	delivery := NewNotesDelivery(mockUsecase, nil)

	currentUserID := uint64(1)
	noteID := uint64(10)
	targetEmail := "collaborator@example.com"
	role := models.RoleEditor
	response := &models.CollaboratorResponse{
		PermissionID: 1,
		Collaborator: models.Collaborator{
			UserID:    2,
			Role:      role,
			GrantedBy: currentUserID,
		},
	}

	t.Run("Success", func(t *testing.T) {
		reqBody, _ := json.Marshal(AddCollaboratorRequest{Email: targetEmail, Role: role})
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/notes/%d/collaborators", noteID), bytes.NewBuffer(reqBody))
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		input := &models.AddCollaboratorInput{
			CurrentUserID: currentUserID,
			NoteID:        noteID,
			Email:         targetEmail,
			Role:          role,
		}
		mockUsecase.EXPECT().AddCollaborator(gomock.Any(), input).Return(response, nil)

		delivery.AddCollaborator(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("NotAuthenticated", func(t *testing.T) {
		reqBody, _ := json.Marshal(AddCollaboratorRequest{Email: targetEmail, Role: role})
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/notes/%d/collaborators", noteID), bytes.NewBuffer(reqBody))
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		rr := httptest.NewRecorder()

		delivery.AddCollaborator(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("InvalidNoteID", func(t *testing.T) {
		reqBody, _ := json.Marshal(AddCollaboratorRequest{Email: targetEmail, Role: role})
		req, _ := http.NewRequest(http.MethodPost, "/notes/invalid/collaborators", bytes.NewBuffer(reqBody))
		req = mux.SetURLVars(req, map[string]string{"note_id": "invalid"})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		delivery.AddCollaborator(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("InvalidBody", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/notes/%d/collaborators", noteID), bytes.NewBuffer([]byte("invalid")))
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		delivery.AddCollaborator(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		reqBody, _ := json.Marshal(AddCollaboratorRequest{Email: targetEmail, Role: role})
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/notes/%d/collaborators", noteID), bytes.NewBuffer(reqBody))
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().AddCollaborator(gomock.Any(), gomock.Any()).Return(nil, errors.New("usecase error"))

		delivery.AddCollaborator(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestNotesDelivery_GetCollaborators(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	logger := zerolog.New(zerolog.NewConsoleWriter())
	hub := websocket.NewHub(&logger)
	delivery := NewNotesDelivery(mockUsecase, hub)

	currentUserID := uint64(1)
	noteID := uint64(10)

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/notes/%d/share", noteID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().GetCollaborators(gomock.Any(), currentUserID, noteID).Return(&models.GetCollaboratorsResponse{}, nil)

		delivery.GetCollaborators(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("NotAuthenticated", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/notes/%d/share", noteID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		rr := httptest.NewRecorder()

		delivery.GetCollaborators(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("InvalidNoteID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/notes/invalid/share", nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": "invalid"})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		delivery.GetCollaborators(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/notes/%d/share", noteID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().GetCollaborators(gomock.Any(), currentUserID, noteID).Return(nil, errors.New("usecase error"))

		delivery.GetCollaborators(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestNotesDelivery_UpdateCollaboratorRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	logger := zerolog.New(zerolog.NewConsoleWriter())
	hub := websocket.NewHub(&logger)
	delivery := NewNotesDelivery(mockUsecase, hub)

	currentUserID := uint64(1)
	noteID := uint64(10)
	permissionID := uint64(5)
	newRole := models.RoleViewer

	t.Run("Success", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"role": newRole,
		})
		req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/notes/%d/share/%d", noteID, permissionID), bytes.NewBuffer(reqBody))
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID), "permission_id": fmt.Sprintf("%d", permissionID)})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		expectedInput := &models.UpdateCollaboratorRoleInput{
			CurrentUserID: currentUserID,
			NoteID:        noteID,
			PermissionID:  permissionID,
			NewRole:       newRole,
		}
		mockUsecase.EXPECT().UpdateCollaboratorRole(gomock.Any(), expectedInput).Return(&models.CollaboratorResponse{}, nil)

		delivery.UpdateCollaboratorRole(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("NotAuthenticated", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/notes/%d/share/%d", noteID, permissionID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID), "permission_id": fmt.Sprintf("%d", permissionID)})
		rr := httptest.NewRecorder()

		delivery.UpdateCollaboratorRole(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("InvalidNoteID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/notes/invalid/share/%d", permissionID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": "invalid", "permission_id": fmt.Sprintf("%d", permissionID)})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		delivery.UpdateCollaboratorRole(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("InvalidPermissionID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/notes/%d/share/invalid", noteID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID), "permission_id": "invalid"})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		delivery.UpdateCollaboratorRole(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("InvalidBody", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/notes/%d/share/%d", noteID, permissionID), bytes.NewBuffer([]byte("invalid")))
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID), "permission_id": fmt.Sprintf("%d", permissionID)})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		delivery.UpdateCollaboratorRole(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"role": newRole,
		})
		req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/notes/%d/share/%d", noteID, permissionID), bytes.NewBuffer(reqBody))
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID), "permission_id": fmt.Sprintf("%d", permissionID)})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().UpdateCollaboratorRole(gomock.Any(), gomock.Any()).Return(nil, errors.New("usecase error"))

		delivery.UpdateCollaboratorRole(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestNotesDelivery_RemoveCollaborator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	logger := zerolog.New(zerolog.NewConsoleWriter())
	hub := websocket.NewHub(&logger)
	delivery := NewNotesDelivery(mockUsecase, hub)

	currentUserID := uint64(1)
	noteID := uint64(10)
	permissionID := uint64(5)

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/notes/%d/share/%d", noteID, permissionID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID), "permission_id": fmt.Sprintf("%d", permissionID)})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().RemoveCollaborator(gomock.Any(), currentUserID, noteID, permissionID).Return(nil)

		delivery.RemoveCollaborator(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("NotAuthenticated", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/notes/%d/share/%d", noteID, permissionID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID), "permission_id": fmt.Sprintf("%d", permissionID)})
		rr := httptest.NewRecorder()

		delivery.RemoveCollaborator(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("InvalidNoteID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/notes/invalid/share/%d", permissionID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": "invalid", "permission_id": fmt.Sprintf("%d", permissionID)})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		delivery.RemoveCollaborator(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("InvalidPermissionID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/notes/%d/share/invalid", noteID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID), "permission_id": "invalid"})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		delivery.RemoveCollaborator(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/notes/%d/share/%d", noteID, permissionID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID), "permission_id": fmt.Sprintf("%d", permissionID)})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().RemoveCollaborator(gomock.Any(), currentUserID, noteID, permissionID).Return(errors.New("usecase error"))

		delivery.RemoveCollaborator(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestNotesDelivery_SetPublicAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	logger := zerolog.New(zerolog.NewConsoleWriter())
	hub := websocket.NewHub(&logger)
	delivery := NewNotesDelivery(mockUsecase, hub)

	currentUserID := uint64(1)
	noteID := uint64(10)
	accessLevel := models.RoleViewer

	t.Run("Success", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"access_level": accessLevel,
		})
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/notes/%d/share/public", noteID), bytes.NewBuffer(reqBody))
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		expectedInput := &models.SetPublicAccessInput{
			CurrentUserID: currentUserID,
			NoteID:        noteID,
			AccessLevel:   &accessLevel,
		}
		mockUsecase.EXPECT().SetPublicAccess(gomock.Any(), expectedInput).Return(&models.PublicAccessResponse{}, nil)

		delivery.SetPublicAccess(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("NotAuthenticated", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/notes/%d/share/public", noteID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		rr := httptest.NewRecorder()

		delivery.SetPublicAccess(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("InvalidNoteID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/notes/invalid/share/public", nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": "invalid"})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		delivery.SetPublicAccess(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("InvalidBody", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/notes/%d/share/public", noteID), bytes.NewBuffer([]byte("invalid")))
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		delivery.SetPublicAccess(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"access_level": accessLevel,
		})
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/notes/%d/share/public", noteID), bytes.NewBuffer(reqBody))
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().SetPublicAccess(gomock.Any(), gomock.Any()).Return(nil, errors.New("usecase error"))

		delivery.SetPublicAccess(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestNotesDelivery_GetPublicAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	logger := zerolog.New(zerolog.NewConsoleWriter())
	hub := websocket.NewHub(&logger)
	delivery := NewNotesDelivery(mockUsecase, hub)

	currentUserID := uint64(1)
	noteID := uint64(10)

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/notes/%d/share/public", noteID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().GetPublicAccess(gomock.Any(), currentUserID, noteID).Return(&models.PublicAccessResponse{}, nil)

		delivery.GetPublicAccess(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("NotAuthenticated", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/notes/%d/share/public", noteID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		rr := httptest.NewRecorder()

		delivery.GetPublicAccess(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("InvalidNoteID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/notes/invalid/share/public", nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": "invalid"})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		delivery.GetPublicAccess(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/notes/%d/share/public", noteID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().GetPublicAccess(gomock.Any(), currentUserID, noteID).Return(nil, errors.New("usecase error"))

		delivery.GetPublicAccess(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestNotesDelivery_GetSharingSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	logger := zerolog.New(zerolog.NewConsoleWriter())
	hub := websocket.NewHub(&logger)
	delivery := NewNotesDelivery(mockUsecase, hub)

	currentUserID := uint64(1)
	noteID := uint64(10)

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/notes/%d/share/settings", noteID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().GetSharingSettings(gomock.Any(), currentUserID, noteID).Return(&models.SharingSettingsResponse{}, nil)

		delivery.GetSharingSettings(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("NotAuthenticated", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/notes/%d/share/settings", noteID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		rr := httptest.NewRecorder()

		delivery.GetSharingSettings(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("InvalidNoteID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/notes/invalid/share/settings", nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": "invalid"})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		delivery.GetSharingSettings(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/notes/%d/share/settings", noteID), nil)
		req = mux.SetURLVars(req, map[string]string{"note_id": fmt.Sprintf("%d", noteID)})
		ctx := middleware.WithUserID(req.Context(), currentUserID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().GetSharingSettings(gomock.Any(), currentUserID, noteID).Return(nil, errors.New("usecase error"))

		delivery.GetSharingSettings(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestNotesDelivery_ActivateAccessByLink(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockNotesUsecase(ctrl)
	logger := zerolog.New(zerolog.NewConsoleWriter())
	hub := websocket.NewHub(&logger)
	delivery := NewNotesDelivery(mockUsecase, hub)

	userID := uint64(1)
	shareUUID := "uuid-123"

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/share/%s", shareUUID), nil)
		req = mux.SetURLVars(req, map[string]string{"share_uuid": shareUUID})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().ActivateAccessByLink(gomock.Any(), shareUUID, userID).Return(&models.ActivateAccessResponse{AccessGranted: true}, nil)

		delivery.ActivateAccessByLink(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("NotAuthenticated", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/share/%s", shareUUID), nil)
		req = mux.SetURLVars(req, map[string]string{"share_uuid": shareUUID})
		rr := httptest.NewRecorder()

		delivery.ActivateAccessByLink(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("UsecaseError", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/share/%s", shareUUID), nil)
		req = mux.SetURLVars(req, map[string]string{"share_uuid": shareUUID})
		ctx := middleware.WithUserID(req.Context(), userID)
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockUsecase.EXPECT().ActivateAccessByLink(gomock.Any(), shareUUID, userID).Return(nil, errors.New("usecase error"))

		delivery.ActivateAccessByLink(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
