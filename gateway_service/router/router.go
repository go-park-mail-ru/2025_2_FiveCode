package router

import (
	"net/http"

	authDelivery "backend/gateway_service/internal/auth/delivery"
	"backend/gateway_service/internal/config"
	exportDelivery "backend/gateway_service/internal/export/delivery"
	fileDelivery "backend/gateway_service/internal/file/delivery"
	mw "backend/gateway_service/internal/middleware"
	notesDelivery "backend/gateway_service/internal/notes/delivery"
	userDelivery "backend/gateway_service/internal/user/delivery"
	"backend/gateway_service/internal/websocket"
	"backend/pkg/metrics"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

func NewRouter(
	conf *config.Config,
	appLogger *zerolog.Logger,
	sessionValidator mw.SessionValidator,
	auth *authDelivery.AuthDelivery,
	user *userDelivery.UserDelivery,
	notes *notesDelivery.NotesDelivery,
	files *fileDelivery.FileDelivery,
	export *exportDelivery.ExportDelivery,
	wsHandler *websocket.Handler,
) http.Handler {

	r := mux.NewRouter()

	r.Use(mw.RequestIDMiddleware(*appLogger), mw.AccessLogMiddleware)

	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/login", auth.Login).Methods("POST")
	api.HandleFunc("/register", auth.Register).Methods("POST")
	api.HandleFunc("/logout", auth.Logout).Methods("POST")
	api.HandleFunc("/session", user.GetProfileBySession).Methods("GET")

	profile := api.PathPrefix("").Subrouter()
	profile.Use(mw.AuthMiddleware(sessionValidator), mw.CSRFMiddleware(conf))

	profile.HandleFunc("/csrf-token", auth.GetCSRFToken).Methods("GET")
	profile.HandleFunc("/profile", user.GetProfile).Methods("GET")
	profile.HandleFunc("/profile", user.UpdateProfile).Methods("PUT")
	profile.HandleFunc("/profile", user.DeleteProfile).Methods("DELETE")

	api.HandleFunc("/icons", files.GetIcons).Methods("GET")
	api.HandleFunc("/headers", files.GetHeaders).Methods("GET")

	notesRouter := api.PathPrefix("").Subrouter()
	notesRouter.Use(mw.AuthMiddleware(sessionValidator), mw.CSRFMiddleware(conf))

	notesRouter.HandleFunc("/notes", notes.GetAllNotes).Methods("GET")
	notesRouter.HandleFunc("/notes", notes.CreateNote).Methods("POST")
	notesRouter.HandleFunc("/notes/search", notes.SearchNotes).Methods("POST")
	notesRouter.HandleFunc("/notes/{note_id}", notes.GetNoteById).Methods("GET")
	notesRouter.HandleFunc("/notes/{note_id}", notes.UpdateNote).Methods("PUT")
	notesRouter.HandleFunc("/notes/{note_id}", notes.DeleteNote).Methods("DELETE")
	notesRouter.HandleFunc("/notes/{note_id}/favorite", notes.AddFavorite).Methods("POST")
	notesRouter.HandleFunc("/notes/{note_id}/favorite", notes.RemoveFavorite).Methods("DELETE")
	notesRouter.HandleFunc("/notes/{note_id}/icons", notes.SetIcon).Methods("PUT")
	notesRouter.HandleFunc("/notes/{note_id}/headers", notes.SetHeader).Methods("PUT")

	exportRouter := api.PathPrefix("").Subrouter()
	exportRouter.Use(mw.AuthMiddleware(sessionValidator), mw.CSRFMiddleware(conf))
	exportRouter.HandleFunc("/notes/{note_id}/export/pdf", export.ExportNoteToPDF).Methods("GET")

	wsRouter := api.PathPrefix("/ws").Subrouter()
	wsRouter.Use(mw.AuthMiddleware(sessionValidator))
	wsRouter.HandleFunc("/notes/{note_id}", wsHandler.HandleConnection).Methods("GET")

	blocksRouter := api.PathPrefix("").Subrouter()
	blocksRouter.Use(mw.AuthMiddleware(sessionValidator), mw.CSRFMiddleware(conf))

	blocksRouter.HandleFunc("/notes/{note_id}/blocks", notes.CreateBlock).Methods("POST")
	blocksRouter.HandleFunc("/notes/{note_id}/blocks", notes.GetBlocks).Methods("GET")
	blocksRouter.HandleFunc("/blocks/{block_id}", notes.GetBlock).Methods("GET")
	blocksRouter.HandleFunc("/blocks/{block_id}", notes.UpdateBlock).Methods("PATCH")
	blocksRouter.HandleFunc("/blocks/{block_id}", notes.DeleteBlock).Methods("DELETE")
	blocksRouter.HandleFunc("/blocks/{block_id}/position", notes.UpdateBlockPosition).Methods("PUT")

	sharingRouter := api.PathPrefix("").Subrouter()
	sharingRouter.Use(mw.AuthMiddleware(sessionValidator), mw.CSRFMiddleware(conf))

	sharingRouter.HandleFunc("/notes/{note_id}/collaborators", notes.AddCollaborator).Methods("POST")
	sharingRouter.HandleFunc("/notes/{note_id}/collaborators", notes.GetCollaborators).Methods("GET")
	sharingRouter.HandleFunc("/notes/{note_id}/collaborators/{permission_id}", notes.UpdateCollaboratorRole).Methods("PATCH")
	sharingRouter.HandleFunc("/notes/{note_id}/collaborators/{permission_id}", notes.RemoveCollaborator).Methods("DELETE")
	sharingRouter.HandleFunc("/notes/{note_id}/public-access", notes.SetPublicAccess).Methods("PUT")
	sharingRouter.HandleFunc("/notes/{note_id}/public-access", notes.GetPublicAccess).Methods("GET")
	sharingRouter.HandleFunc("/notes/{note_id}/sharing", notes.GetSharingSettings).Methods("GET")
	sharingRouter.HandleFunc("/notes/activate/{share_uuid}", notes.ActivateAccessByLink).Methods("POST")

	filesRouter := api.PathPrefix("/files").Subrouter()
	filesRouter.Use(mw.AuthMiddleware(sessionValidator), mw.CSRFMiddleware(conf))

	filesRouter.HandleFunc("/upload", files.UploadFile).Methods("POST")
	filesRouter.HandleFunc("/{file_id}", files.GetFile).Methods("GET")
	filesRouter.HandleFunc("/{file_id}", files.DeleteFile).Methods("DELETE")

	r.Handle("/metrics", promhttp.HandlerFor(metrics.Registry(), promhttp.HandlerOpts{})).Methods("GET")

	return mw.CORS(r, conf)
}
