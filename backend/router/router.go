package router

import (
	"backend/initialize"
	mw "backend/middleware"
	"backend/store"
	"net/http"

	_ "backend/docs"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(s *store.Store, deliveries *initialize.Deliveries) http.Handler {
	r := mux.NewRouter()
	r.Use(mw.RequestIDMiddleware, mw.AccessLogMiddleware)

	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/login", deliveries.AuthDelivery.Login).Methods("POST")
	api.HandleFunc("/register", deliveries.UserDelivery.Register).Methods("POST")
	api.HandleFunc("/logout", deliveries.AuthDelivery.Logout).Methods("POST")
	api.HandleFunc("/session", deliveries.UserDelivery.GetProfileBySession).Methods("GET")
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	profile := api.PathPrefix("").Subrouter()
	profile.Use(mw.AuthMiddleware(s))
	profile.HandleFunc("/profile", deliveries.UserDelivery.GetProfile).Methods("GET")
	profile.HandleFunc("/profile", deliveries.UserDelivery.UpdateProfile).Methods("PUT")

	notes := api.PathPrefix("").Subrouter()
	notes.Use(mw.AuthMiddleware(s))
	notes.HandleFunc("/notes", deliveries.NotesDelivery.GetAllNotes).Methods("GET")
	notes.HandleFunc("/notes", deliveries.NotesDelivery.CreateNote).Methods("POST")
	notes.HandleFunc("/notes/{note_id}", deliveries.NotesDelivery.GetNoteById).Methods("GET")
	notes.HandleFunc("/notes/{note_id}", deliveries.NotesDelivery.UpdateNote).Methods("PUT")
	notes.HandleFunc("/notes/{note_id}", deliveries.NotesDelivery.DeleteNote).Methods("DELETE")

	blocks := api.PathPrefix("").Subrouter()
	blocks.Use(mw.AuthMiddleware(s))
	blocks.HandleFunc("/notes/{note_id}/blocks", deliveries.BlocksDelivery.CreateBlock).Methods("POST")
	blocks.HandleFunc("/notes/{note_id}/blocks", deliveries.BlocksDelivery.GetBlocks).Methods("GET")
	blocks.HandleFunc("/blocks/{block_id}", deliveries.BlocksDelivery.GetBlock).Methods("GET")
	blocks.HandleFunc("/blocks/{block_id}", deliveries.BlocksDelivery.UpdateBlock).Methods("PATCH")
	blocks.HandleFunc("/blocks/{block_id}", deliveries.BlocksDelivery.DeleteBlock).Methods("DELETE")
	blocks.HandleFunc("/blocks/{block_id}/position", deliveries.BlocksDelivery.UpdateBlockPosition).Methods("PUT")

	return mw.CORS(r)
}
