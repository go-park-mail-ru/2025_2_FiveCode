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
	api.HandleFunc("/register", deliveries.AuthDelivery.Register).Methods("POST")
	api.HandleFunc("/logout", deliveries.AuthDelivery.Logout).Methods("POST")
	api.HandleFunc("/session", deliveries.UserDelivery.GetProfileBySession).Methods("GET")
	api.HandleFunc("/tickets", deliveries.TicketDelivery.CreateTicket).Methods("POST")
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	profile := api.PathPrefix("").Subrouter()
	profile.Use(mw.AuthMiddleware(s))
	profile.HandleFunc("/profile", deliveries.UserDelivery.GetProfile).Methods("GET")
	profile.HandleFunc("/profile", deliveries.UserDelivery.UpdateProfile).Methods("PUT")
	profile.HandleFunc("/profile", deliveries.UserDelivery.DeleteProfile).Methods("DELETE")

	notes := api.PathPrefix("").Subrouter()
	notes.Use(mw.AuthMiddleware(s))
	notes.HandleFunc("/notes", deliveries.NotesDelivery.GetAllNotes).Methods("GET")
	notes.HandleFunc("/notes", deliveries.NotesDelivery.CreateNote).Methods("POST")
	notes.HandleFunc("/notes/{note_id}", deliveries.NotesDelivery.GetNoteById).Methods("GET")
	notes.HandleFunc("/notes/{note_id}", deliveries.NotesDelivery.UpdateNote).Methods("PUT")
	notes.HandleFunc("/notes/{note_id}", deliveries.NotesDelivery.DeleteNote).Methods("DELETE")
	notes.HandleFunc("/notes/{note_id}/favorite", deliveries.NotesDelivery.AddFavorite).Methods("POST")
	notes.HandleFunc("/notes/{note_id}/favorite", deliveries.NotesDelivery.RemoveFavorite).Methods("DELETE")

	blocks := api.PathPrefix("").Subrouter()
	blocks.Use(mw.AuthMiddleware(s))
	blocks.HandleFunc("/notes/{note_id}/blocks", deliveries.BlocksDelivery.CreateBlock).Methods("POST")
	blocks.HandleFunc("/notes/{note_id}/blocks", deliveries.BlocksDelivery.GetBlocks).Methods("GET")
	blocks.HandleFunc("/blocks/{block_id}", deliveries.BlocksDelivery.GetBlock).Methods("GET")
	blocks.HandleFunc("/blocks/{block_id}", deliveries.BlocksDelivery.UpdateBlock).Methods("PATCH")
	blocks.HandleFunc("/blocks/{block_id}", deliveries.BlocksDelivery.DeleteBlock).Methods("DELETE")
	blocks.HandleFunc("/blocks/{block_id}/position", deliveries.BlocksDelivery.UpdateBlockPosition).Methods("PUT")

	filesRouter := r.PathPrefix("/api/files").Subrouter()
	filesRouter.HandleFunc("/upload", deliveries.FileDelivery.UploadFile).Methods("POST")
	filesRouter.HandleFunc("/{file_id}", deliveries.FileDelivery.GetFile).Methods("GET")
	filesRouter.HandleFunc("/{file_id}", deliveries.FileDelivery.DeleteFile).Methods("DELETE")

	admin := api.PathPrefix("/admin").Subrouter()
	admin.Use(mw.AuthMiddleware(s))
	admin.HandleFunc("/statistics", deliveries.TicketDelivery.GetStatistics).Methods("GET")
	admin.HandleFunc("/tickets", deliveries.TicketDelivery.GetAllTickets).Methods("GET")
	admin.HandleFunc("/tickets/{ticket_id}", deliveries.TicketDelivery.UpdateTicketStatus).Methods("PUT")
	
	ticketProtectRouter := r.PathPrefix("/api/").Subrouter()
	ticketProtectRouter.Use(mw.AuthMiddleware(s))
	ticketProtectRouter.HandleFunc("/tickets", deliveries.TicketDelivery.GetAllTicketsByUserId).Methods("GET")
	ticketProtectRouter.HandleFunc("/tickets/{ticket_id}", deliveries.TicketDelivery.UpdateTicket).Methods("PUT")
	ticketProtectRouter.HandleFunc("/tickets/{ticket_id}", deliveries.TicketDelivery.GetTicketById).Methods("GET")

	return mw.CORS(r)
}
