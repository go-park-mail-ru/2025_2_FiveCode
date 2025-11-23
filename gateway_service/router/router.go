package router

import (
	"net/http"

	authDelivery "backend/gateway_service/internal/auth/delivery"
	"backend/gateway_service/internal/config"
	fileDelivery "backend/gateway_service/internal/file/delivery"
	mw "backend/gateway_service/internal/middleware"
	notesDelivery "backend/gateway_service/internal/notes/delivery"
	userDelivery "backend/gateway_service/internal/user/delivery"

	"github.com/gorilla/mux"
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

	notesRouter := api.PathPrefix("").Subrouter()
	notesRouter.Use(mw.AuthMiddleware(sessionValidator), mw.CSRFMiddleware(conf))

	notesRouter.HandleFunc("/notes", notes.GetAllNotes).Methods("GET")
	notesRouter.HandleFunc("/notes", notes.CreateNote).Methods("POST")
	notesRouter.HandleFunc("/notes/{note_id}", notes.GetNoteById).Methods("GET")
	notesRouter.HandleFunc("/notes/{note_id}", notes.UpdateNote).Methods("PUT")
	notesRouter.HandleFunc("/notes/{note_id}", notes.DeleteNote).Methods("DELETE")
	notesRouter.HandleFunc("/notes/{note_id}/favorite", notes.AddFavorite).Methods("POST")
	notesRouter.HandleFunc("/notes/{note_id}/favorite", notes.RemoveFavorite).Methods("DELETE")

	blocksRouter := api.PathPrefix("").Subrouter()
	blocksRouter.Use(mw.AuthMiddleware(sessionValidator), mw.CSRFMiddleware(conf))

	blocksRouter.HandleFunc("/notes/{note_id}/blocks", notes.CreateBlock).Methods("POST")
	blocksRouter.HandleFunc("/notes/{note_id}/blocks", notes.GetBlocks).Methods("GET")
	blocksRouter.HandleFunc("/blocks/{block_id}", notes.GetBlock).Methods("GET")
	blocksRouter.HandleFunc("/blocks/{block_id}", notes.UpdateBlock).Methods("PATCH")
	blocksRouter.HandleFunc("/blocks/{block_id}", notes.DeleteBlock).Methods("DELETE")
	blocksRouter.HandleFunc("/blocks/{block_id}/position", notes.UpdateBlockPosition).Methods("PUT")

	filesRouter := api.PathPrefix("/files").Subrouter()
	filesRouter.Use(mw.AuthMiddleware(sessionValidator), mw.CSRFMiddleware(conf))

	filesRouter.HandleFunc("/upload", files.UploadFile).Methods("POST")
	filesRouter.HandleFunc("/{file_id}", files.GetFile).Methods("GET")
	filesRouter.HandleFunc("/{file_id}", files.DeleteFile).Methods("DELETE")

	return mw.CORS(r, conf)
}
